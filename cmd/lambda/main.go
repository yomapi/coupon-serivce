package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

type CouponMessage struct {
	CampaignID int32  `json:"campaign_id"`
	UserID     string `json:"user_id"`
}

func getDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	return sql.Open("postgres", dsn)
}

func getRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	port := os.Getenv("REDIS_PORT")
	addr := fmt.Sprintf("%s:%s", host, port)

	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	db, err := getDB()
	if err != nil {
		return fmt.Errorf("❌ DB 연결 실패: %v", err)
	}
	defer db.Close()

	rdb := getRedisClient()
	defer rdb.Close()

	for _, record := range sqsEvent.Records {
		var msg CouponMessage
		if err := json.Unmarshal([]byte(record.Body), &msg); err != nil {
			fmt.Printf("❌ 메시지 파싱 실패: %v\n", err)
			continue
		}

		fmt.Printf("✅ 메시지 수신: campaign_id=%d user_id=%s\n", msg.CampaignID, msg.UserID)

		code := generateCouponCode()
		err := issueCoupon(ctx, db, msg.CampaignID, msg.UserID, code)
		if err != nil {
			fmt.Printf("❌ 쿠폰 발급 실패: %v\n", err)
			continue
		}

		// ✅ Redis에서 SREM 처리
		requestedKey := fmt.Sprintf("coupon_requested:%d", msg.CampaignID)
		_, err = rdb.SRem(ctx, requestedKey, msg.UserID).Result()
		if err != nil {
			fmt.Printf("❌ Redis SREM 실패: %v\n", err)
			continue
		}

		fmt.Printf("🎉 쿠폰 발급 완료 및 Redis 정리: campaign_id=%d user_id=%s code=%s\n", msg.CampaignID, msg.UserID, code)
	}

	return nil
}

func issueCoupon(ctx context.Context, db *sql.DB, campaignID int32, userID string, code string) error {
	query := `
		INSERT INTO coupons (code, user_id, campaign_id, issued_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := db.ExecContext(ctx, query, code, userID, campaignID, time.Now())
	return err
}

func generateCouponCode() string {
	return fmt.Sprintf("COUPON-%d", time.Now().UnixNano())
}

func main() {
	lambda.Start(handler)
}
