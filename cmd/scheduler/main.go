package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yomapi/coupon-service/internal/db"
	"github.com/yomapi/coupon-service/internal/redis"
	"github.com/yomapi/coupon-service/internal/scheduler"
	"github.com/yomapi/coupon-service/internal/sqs"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("❌ SQS_QUEUE_URL 환경변수가 필요합니다")
	}
	sqsClient, err := sqs.NewSQSClient(ctx, queueURL)
	if err != nil {
		log.Fatalf("❌ SQS 클라이언트 초기화 실패: %v", err)
	}

	redisClient := redis.NewRedisClient()

	pgPool, err := db.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("❌ DB 연결 실패: %v", err)
	}
	defer pgPool.Close()

	processor := scheduler.NewProcessor(redisClient, sqsClient, pgPool, 10) // 예: 초당 최대 10개 처리

	log.Println("🚀 Scheduler 시작")
	processor.Run(ctx)
}
