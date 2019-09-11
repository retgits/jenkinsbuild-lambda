package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

const (
	AWSRegion = "us-west-2"
	URL       = "http://jenkins.dev:8080/jenkins"
	User      = "admin"
	AuthToken = "12345"
	BuildID   = "MyAwesomeBuild"
)

func TestHandler(t *testing.T) {
	assert := assert.New(t)

	bytes, err := ioutil.ReadFile("./test/event.json")
	assert.NoError(err)

	var request events.SQSEvent
	err = json.Unmarshal(bytes, &request)
	assert.NoError(err)

	err = handler(request)
	assert.Error(err)

	awsSession := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(AWSRegion),
	}))

	kmsSvc := kms.New(awsSession)

	os.Setenv("AWS_REGION", AWSRegion)
	os.Setenv("JENKINS_URL", encodeString(kmsSvc, URL))
	os.Setenv("JENKINS_USER", encodeString(kmsSvc, User))
	os.Setenv("JENKINS_TOKEN", encodeString(kmsSvc, AuthToken))

	err = handler(request)
	assert.Error(err)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("POST", fmt.Sprintf("/job/%s/build", BuildID), httpmock.NewStringResponder(200, `success...`))
	httpmock.RegisterResponder("POST", "https://kms.us-west-2.amazonaws.com/", httpmock.NewStringResponder(200, `{"plaintext":"bla"}`))

	err = handler(request)
	assert.NoError(err)
}

func encodeString(kmsSvc *kms.KMS, payload string) string {
	output, err := kmsSvc.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String("alias/retgits/lambda"),
		Plaintext: []byte(payload),
	})
	if err != nil {
		panic(err.Error())
	}

	return base64.StdEncoding.EncodeToString(output.CiphertextBlob)
}
