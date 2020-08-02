#!/bin/bash
go get github.com/aws/aws-lambda-go/lambda
GOOS=linux go build main.go
zip function.zip main
aws lambda create-function --function-name {{YOUR-LAMBDA-FUNCTION}} --runtime go1.x \
  --zip-file fileb://function.zip --handler main \
  --role {{YOUR-LAMBDA-ARN}}

./build-cleanup.sh