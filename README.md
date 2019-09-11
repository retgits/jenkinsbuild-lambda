# lambda-jenkinsbuild

Trigger builds using the [Jenkins REST API](https://wiki.jenkins.io/display/JENKINS/Remote+access+API) based on messages published on an [Amazon SQS](https://aws.amazon.com/sqs/) queue.

## Prerequisites

### Environment Variables

The app relies on [AWS Systems Manager Parameter Store](https://aws.amazon.com/systems-manager/features/) (SSM) to store encrypted variables on how to connect to Jenkins. The variables it relies on are:

* `/prod/jenkins/token`: The token to authenticate to Jenkins
* `/prod/jenkins/user`: The username to authenticate to Jenkins
* `/prod/jenkins/url`: The URL to Jenkins

These parameters are encrypted using [Amazon KMS](https://aws.amazon.com/kms/) and retrieved from the Parameter Store on deployment. This way the encrypted variables are given to the Lambda function and the function needs to take care of decrypting them at runtime.

To create the encrypted variables, run the below command for all of the variables

```bash
aws ssm put-parameter                       \
   --type String                            \
   --name "/prod/jenkins/token"             \
   --value $(aws kms encrypt                \
              --output text                 \
              --query CiphertextBlob        \
              --key-id <YOUR_KMS_KEY_ID>    \
              --plaintext "PLAIN TEXT HERE")
```

To test the function locally, using `SAM`, you'll have to uncomment lines 45-47 in the `template.yaml` file and update these with base64 encoded values from the Parameter Store. Only during deployment to AWS Lambda will these variables get their actual values from the Parameter Store.

### SQS queue

The SSM parameter `/prod/jenkins/sqsqueue` doesn't have to be encrypted but does need to reference a valid SQS ARN. The payload of the message the function expects is

```json
{
   "BuildID": "MyAwesomeBuild"
}
```

In this case, _MyAwesomeBuild_ is the name of the job to trigger in Jenkins.

## Build and Deploy

There are several `Make` targets available to help build and deploy the function

| Target | Description                                       |
|--------|---------------------------------------------------|
| build  | Build the executable for Lambda                   |
| clean  | Remove all generated files                        |
| deploy | Deploy the app to AWS Lambda                      |
| deps   | Get the Go modules from the GOPROXY               |
| help   | Displays the help for each target (this message). |
| local  | Run SAM to test the Lambda function using Docker  |
| test   | Run all unit tests and print coverage             |
