# sqsreceiver

A simple Amazon SQS receiver which receives messages from an SQS queue and triggers a build in Jenkins

## Build
To build simply run
```bash
./build.sh deps
./build.sh build
```

## Deploy
To deploy simply run these commands
```bash
docker build . -t <yourname>/sqsreceiver
```

## Run container
The Docker container relies on three values:

* QUEUE: The name of the Amazon SQS queue
* AWSACCESSKEYID: The Access Key ID to connect to AWS SQS
* AWSSECRETACCESSKEY: The Secret Access Key to connect to AWS SQS
* AWSREGION: The region you're your Amazon SQS queue in
* JENKINSUSER: The username to connect to Jenkins
* JENKINSTOKEN: The API token to connect to Jenkins
* JENKINSSERVER: The base URL of your Jenkins server

You can run the docker container using:
```bash
docker run -d -e QUEUE='' \
              -e AWSACCESSKEYID='' \
              -e AWSSECRETACCESSKEY='' \
              -e AWSREGION='' \
              -e JENKINSUSER='' \
              -e JENKINSTOKEN='' \
              -e JENKINSSERVER='' \
              --name=sqsreceiver <yourname>/sqsreceiver
```