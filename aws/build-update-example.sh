#!/bin/bash
go get github.com/aws/aws-lambda-go/lambda
GOOS=linux go build main.go
zip function.zip main
aws lambda update-function-code --function-name {{YOUR-LAMBDA-FUNCTION}} \
  --zip-file fileb://function.zip \


./build-cleanup.sh
