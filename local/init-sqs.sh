#!/bin/bash

set -e

# 변수 정의
QUEUE_NAME="coupon-queue"
QUEUE_URL=""
FUNCTION_NAME="coupon-processor"
ZIP_FILE="function.zip"
ROLE_ARN="arn:aws:iam::000000000000:role/lambda-role"
REGION="us-east-1"

echo "✅ Step 1: SQS 큐 생성"
docker exec -it local-localstack awslocal sqs create-queue --queue-name $QUEUE_NAME

QUEUE_URL=$(docker exec -it local-localstack awslocal sqs get-queue-url --queue-name $QUEUE_NAME | jq -r '.QueueUrl')

echo "✅ SQS 큐 URL: $QUEUE_URL"

echo "✅ Step 2: Lambda Go 바이너리 빌드"
GOOS=linux GOARCH=amd64 go build -o main cmd/lambda/main.go

echo "✅ Step 3: Lambda zip 패키징"
zip -j $ZIP_FILE main

echo "✅ Step 4: Lambda zip 파일 LocalStack 컨테이너로 복사"
docker cp $ZIP_FILE local-localstack:/tmp/$ZIP_FILE

echo "✅ Step 5: Lambda 함수 등록"
docker exec -it local-localstack awslocal lambda create-function \
  --function-name coupon-processor \
  --runtime go1.x \
  --handler main \
  --role arn:aws:iam::000000000000:role/lambda-role \
  --zip-file fileb:///tmp/function.zip \
  --environment Variables="{DB_HOST=local-postgres,DB_PORT=5432,DB_USER=postgres,DB_PASSWORD=password,DB_NAME=coupon_service,REDIS_HOST=local-redis,REDIS_PORT=6379}"


echo "✅ Step 6: Lambda ↔ SQS 이벤트 소스 매핑"
docker exec -it local-localstack awslocal lambda create-event-source-mapping \
  --function-name $FUNCTION_NAME \
  --batch-size 10 \
  --event-source-arn arn:aws:sqs:$REGION:000000000000:$QUEUE_NAME

echo "🎉 초기 등록 완료!"
