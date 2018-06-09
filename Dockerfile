# Usage:
# docker run -d -e QUEUE='' \
#               -e AWSACCESSKEYID='' \
#               -e AWSSECRETACCESSKEY='' \
#               -e AWSREGION='' \
#               -e JENKINSUSER='' \
#               -e JENKINSTOKEN='' \
#               -e JENKINSSERVER='' \
#               --name=sqsreceiver retgits/sqsreceiver
#
# docker build . -t retgits/sqsreceiver

FROM alpine
LABEL maintainer "retgits"
RUN apk update && apk add ca-certificates
ADD bin/main .
CMD ./main