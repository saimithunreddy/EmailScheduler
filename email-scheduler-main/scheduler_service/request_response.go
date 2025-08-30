package service

import "time"

type ScheduleEmailNotificationRequest struct {
	Email   string    `json:"email" dynamodbav:"pk"`
	Subject string    `json:"subject" dynamodbav:"subject"`
	Message string    `json:"message" dynamodbav:"message"`
	Date    time.Time `json:"date" dynamodbav:"sk"`
}

type ScheduleEmailNotificationResponse struct {
	Message string `json:"message"`
}
