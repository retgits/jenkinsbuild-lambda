package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/valyala/fastjson"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var (
	queueName          = os.Getenv("QUEUE")
	awsAccessKeyID     = os.Getenv("AWSACCESSKEYID")
	awsSecretAccessKey = os.Getenv("AWSSECRETACCESSKEY")
	awsRegion          = os.Getenv("AWSREGION")
	jenkinsUser        = os.Getenv("JENKINSUSER")
	jenkinsToken       = os.Getenv("JENKINSTOKEN")
	jenkinsServer      = os.Getenv("JENKINSSERVER")
)

func main() {
	// Log start
	log.Printf("Starting SQS Receiver process.")

	// Channel on which to send received messages.
	msgChan := make(chan sqs.Message, 100)

	// Channel to signal done
	doneChan := make(chan bool)

	// Create new credentials using the accessKey and secretKey
	awsCredentials := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")

	// Create a new session with AWS credentials
	awsSession := session.Must(session.NewSession(&aws.Config{
		Credentials: awsCredentials,
		Region:      aws.String(awsRegion),
	}))

	// Create a SQS service client.
	sqsSvcs := sqs.New(awsSession)

	// Need to convert the queue name into a URL. Make the GetQueueUrl
	// API call to retrieve the URL. This is needed for receiving messages
	// from the queue.
	queueURL, err := sqsSvcs.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			log.Printf("Unable to find queue %q.", queueName)
			os.Exit(2)
		}
		log.Printf("Unable to connect to %q: %v.", queueName, err)
		os.Exit(2)
	}

	log.Printf("Connected to Amazon SQS.")

	// Receive a message from the SQS queue and send that on the msgChan
	go func() {
		log.Printf("Starting queue receiver goroutine.")
		for {
			result, err := sqsSvcs.ReceiveMessage(&sqs.ReceiveMessageInput{
				QueueUrl: queueURL.QueueUrl,
				AttributeNames: aws.StringSlice([]string{
					"SentTimestamp",
				}),
				MaxNumberOfMessages: aws.Int64(1),
				MessageAttributeNames: aws.StringSlice([]string{
					"All",
				}),
			})
			if err != nil {
				log.Printf("Unable to receive message from %q: %v.", queueName, err)
				doneChan <- true
			}
			if len(result.Messages) > 0 {
				for _, msg := range result.Messages {
					msgChan <- *msg
				}
			}
		}
	}()

	// Listen to messages on the msgChan
	go func() {
		log.Printf("Starting message processor goroutine.")
		for {
			select {
			case msg := <-msgChan:
				log.Println("received message!")
				err := messageProcessor(msg)
				if err == nil {
					_, err := sqsSvcs.DeleteMessage(&sqs.DeleteMessageInput{
						QueueUrl:      queueURL.QueueUrl,
						ReceiptHandle: msg.ReceiptHandle,
					})
					if err != nil {
						log.Println(err.Error())
					}
				}
			default:
			}
		}
	}()

	<-doneChan
}

func messageProcessor(msg sqs.Message) error {
	// Get the body
	body := *msg.Body
	body = body[8:]

	// Parse the string
	body, err := url.QueryUnescape(body)
	if err != nil {
		fmt.Printf("Error while parsing query string: %s\n", err.Error())
		return err
	}

	// Extract the payload
	str := fastjson.GetString([]byte(body), "repository", "full_name")
	return triggerJenkinsBuild(str)
}

func triggerJenkinsBuild(reponame string) error {
	// Prepare the API request
	reponame = strings.Replace(reponame, "/", "-", -1)
	auth := []byte(fmt.Sprintf("%s:%s", jenkinsUser, jenkinsToken))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/job/%s/build", jenkinsServer, reponame), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(auth)))

	log.Printf("%s", req.URL)

	// Prepare the HTTP client
	client := &http.Client{}

	// Execute the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	log.Printf("Message for %s to Jenkins resulted in\nHTTP StatusCode %v\nHTTP Body %v\n", reponame, resp.StatusCode, buf.String())

	return nil
}
