# Email Scheduler

This is a simple email service sends email at specified time uses Gorilla Mux to handle HTTP requests. The service allows you to send emails by making a POST request to the `/email` endpoint.

Utilised Redis sorted list to create a queue of emails and used go routine to fetch the pending emails continuosly. Utilised DynamoDB to store the email data of all the emails.

## Installation

1. Clone the repository:

	```sh
	git clone https://github.com/nithin-gith/email_service.git
	cd email_service
	```

2. Install the dependencies:

	```sh
	go mod download
	```

## Usage

1. Start the server:

	```sh
	go run main.go
	```

2. Send a POST request to the `/send-email` endpoint with the following JSON payload:

	```json
	{
		"email": "recipient@example.com",
		"subject": "Your Subject",
		"message": "Your message here",
        "date":"2024-07-14T17:02:17Z"
	}
	```

	Example using `curl`:

	```sh
	curl -X POST http://localhost:8080/email -H "Content-Type: application/json" -d '{
		"email": "recipient@example.com",
		"subject": "Your Subject",
		"message": "Your message here",
        "date":"2024-07-14T17:02:17Z"
	}'
	```



