package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/nithin-gith/email-scheduler/email_service"
	scheduler "github.com/nithin-gith/email-scheduler/scheduler_service"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log = logrus.New()

func main() {

	// Load the configuration file
	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		log.Error("Error reading config file: ", err)
	}
	// Connect to Redis
	redisAddr := viper.GetString("redis.address")
	redisPassword := viper.GetString("redis.password")
	redisDB := viper.GetInt("redis.db")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	ctx := context.Background()
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Error("Could not connect to Redis: ", err)
		os.Exit(1)
	}
	log.Info("Connected to Redis")
	defer redisClient.Close()

	// Connect to DynamoDB
	dynamoRegion := viper.GetString("dynamodb.region")
	dynamoEndpoint := viper.GetString("dynamodb.endpoint")
	dynamoAccessKeyID := viper.GetString("dynamodb.access_key_id")
	dynamoSecretAccessKey := viper.GetString("dynamodb.secret_access_key")
	dynamoSessionToken := viper.GetString("dynamodb.session_token")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(dynamoRegion),
		Endpoint:    aws.String(dynamoEndpoint),
		Credentials: credentials.NewStaticCredentials(dynamoAccessKeyID, dynamoSecretAccessKey, dynamoSessionToken),
	})
	if err != nil {
		log.Error("Failed to create AWS session: ", err)
		os.Exit(1)
	}

	dynamoDBClient := dynamodb.New(sess)
	tablesNames := &dynamodb.ListTablesOutput{}
	if tablesNames, err = dynamoDBClient.ListTables(&dynamodb.ListTablesInput{}); err != nil {
		log.Error("Failed to connect to DynamoDB: ", err)
		os.Exit(1)
	}
	log.Info("Connected to DynamoDB", tablesNames)
	log.Info("Connected to DynamoDB")

	emailSchedulerService := scheduler.NewService(redisClient, dynamoDBClient)

	r := mux.NewRouter()
	r.HandleFunc("/email", emailSchedulerService.ScheduleEmailNotification).Methods("POST")

	go emailWorker(redisClient)
	http.Handle("/", r)
	log.Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}

}

func emailWorker(redisClient *redis.Client) {
	ctx := context.Background()
	for {
		currentUnixTime := time.Now().UTC().Unix()

		emailDataStr, err := redisClient.ZRangeByScore(ctx, "email:schedule", &redis.ZRangeBy{
			Max:   strconv.FormatInt(currentUnixTime, 10),
			Count: 1,
		}).Result()
		if err != nil {
			log.Error("Failed to get scheduled emails: ", emailDataStr)
			redisClient.ZRem(ctx, "email:schedule", emailDataStr)
			continue
		}

		if len(emailDataStr) == 0 {
			log.Error("No Emails pending", emailDataStr)
			time.Sleep(1 * time.Second)
			continue
		}

		emailData := email_service.SendEmailRequest{}
		err = json.Unmarshal([]byte(emailDataStr[0]), &emailData)
		if err != nil {
			log.Error("Failed to unmarshal email: ", err)
			redisClient.ZRem(ctx, "email:schedule", emailDataStr)
			continue
		}

		err = email_service.SendEmail(email_service.SendEmailRequest{
			Email:   emailData.Email,
			Subject: emailData.Subject,
			Message: emailData.Message,
		})
		if err != nil {
			log.Error("Failed to send email: ", err)
			redisClient.ZRem(ctx, "email:schedule", emailDataStr)
			continue
		}

		redisClient.ZRem(ctx, "email:schedule", emailDataStr)
	}
}
