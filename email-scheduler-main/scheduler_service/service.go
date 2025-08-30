package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type Service interface {
	ScheduleEmailNotification(w http.ResponseWriter, r *http.Request)
}

type service struct {
	redisClient    *redis.Client
	dynamoDBClient *dynamodb.DynamoDB
}

func NewService(redisClient *redis.Client, dynamoDBClient *dynamodb.DynamoDB) Service {
	return &service{
		redisClient:    redisClient,
		dynamoDBClient: dynamoDBClient,
	}
}

var log = logrus.New()

func (s *service) ScheduleEmailNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var req ScheduleEmailNotificationRequest
	if err := DecodeScheduleEmailNotification(w, r, &req); err != nil {
		log.Error("err", err, "Failed to decode request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Info("req", req, "msg ", "Received data")

	reqString, err := json.Marshal(req)
	if err != nil {
		log.Error("err", err, "Failed to marshal request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	redisRes := s.redisClient.ZAdd(ctx, "email:schedule", &redis.Z{
		Score:  float64(req.Date.UTC().Unix()),
		Member: reqString,
	})

	if redisRes.Err() != nil {
		log.Error("err", redisRes.Err(), "Failed to schedule email")
		http.Error(w, redisRes.Err().Error(), http.StatusInternalServerError)
		return
	}

	av, err := dynamodbattribute.MarshalMap(req)
	if err != nil {
		log.Error("err", err, "Failed to dynamodb request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dynamoInput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("scheduled_emails"),
	}

	dynamoRes, err := s.dynamoDBClient.PutItem(dynamoInput)

	if err != nil {
		log.Error("err", err, "Failed to save email to DynamoDB")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("dynamoRes", dynamoRes, "msg", "Successfully saved email to DynamoDB")

	resData := ScheduleEmailNotificationResponse{
		Message: "Successfully scheduled email",
	}
	if err := EncodeJSONBody(w, resData); err != nil {
		log.Error("err", err, "Failed to encode response body")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
