package email_service

import (
	"fmt"
	"net/smtp"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var log = logrus.New()

func SendEmail(sendEmailRequest SendEmailRequest) (err error) {

	smtpServer := viper.GetString("smtp.smtp_server")
	smtpPort := viper.GetInt("smtp.smtp_port")
	senderEmail := viper.GetString("smtp.smtp_user")
	senderPassword := viper.GetString("smtp.smtp_password")

	toEmail := sendEmailRequest.Email
	emailSubject := sendEmailRequest.Subject
	emailMessage := sendEmailRequest.Message

	auth := smtp.PlainAuth("", senderEmail, senderPassword, smtpServer)
	msg := []byte("To: " + toEmail + "\r\n" +
		"Subject: " + emailSubject + "\r\n" +
		"\r\n" +
		emailMessage + "\r\n")

	// Send the email.
	err = smtp.SendMail(smtpServer+":"+fmt.Sprintf("%d", smtpPort), auth, senderEmail, []string{toEmail}, msg)
	if err != nil {
		log.Error(smtpServer + ":" + fmt.Sprintf("%d", smtpPort))
		log.Error("err", err)
		return err
	}

	log.Info("msg", "Email sent successfully", "emailData", sendEmailRequest)

	return nil
}
