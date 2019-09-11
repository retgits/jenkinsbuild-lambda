// Package main contains the main logic to receive messages from Amazon SQS
// and trigger a Jenkins webhook based on that
package main

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/kelseyhightower/envconfig"
	"github.com/retgits/jenkinsbuild-lambda/jenkins"
)

type config struct {
	AWSRegion    string `required:"true" split_words:"true"`
	JenkinsURL   string `required:"true" split_words:"true"`
	JenkinsUser  string `required:"true" split_words:"true"`
	JenkinsToken string `required:"true" split_words:"true"`
}

var c config

func handler(request events.SQSEvent) error {
	// Get configuration set using environment variables
	err := envconfig.Process("", &c)
	if err != nil {
		return err
	}

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.AWSRegion),
	}))

	kmsSvc := kms.New(awsSession)

	val, err := decodeString(kmsSvc, c.JenkinsURL)
	if err != nil {
		return err
	}
	c.JenkinsURL = val

	val, err = decodeString(kmsSvc, c.JenkinsUser)
	if err != nil {
		return err
	}
	c.JenkinsUser = val

	val, err = decodeString(kmsSvc, c.JenkinsToken)
	if err != nil {
		return err
	}
	c.JenkinsToken = val

	srv := jenkins.NewServer(c.JenkinsURL, c.JenkinsUser, c.JenkinsToken)

	for _, req := range request.Records {
		evt, err := jenkins.UnmarshalBuildRequest([]byte(req.Body))
		if err != nil {
			return err
		}
		res, err := srv.TriggerBuild(&evt)
		if err != nil {
			return err
		}
		fmt.Printf("%d/%s: %s\n", res.HTTPStatusCode, res.HTTPStatusMessage, res.JenkinsResponse)
	}

	return nil
}

// decodeString uses AWS Key Management Service (AWS KMS) to decrypt environment variables.
// In order for this method to work, the function needs access to the kms:Decrypt capability.
func decodeString(kmsSvc *kms.KMS, payload string) (string, error) {
	sDec, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}
	out, err := kmsSvc.Decrypt(&kms.DecryptInput{
		CiphertextBlob: sDec,
	})
	if err != nil {
		return "", err
	}
	return string(out.Plaintext), nil
}

// The main method is executed by AWS Lambda and points to the handler
func main() {
	lambda.Start(handler)
}
