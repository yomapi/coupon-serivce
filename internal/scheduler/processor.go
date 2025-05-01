package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yomapi/coupon-service/internal/redis"
	"github.com/yomapi/coupon-service/internal/sqs"
)

type Processor struct {
	RedisClient *redis.RedisClient
	SQSClient   *sqs.SQSClient
	DB          *pgxpool.Pool
	MaxPerTick  int64
}

func NewProcessor(rdb *redis.RedisClient, sqsClient *sqs.SQSClient, db *pgxpool.Pool, maxPerTick int64) *Processor {
	return &Processor{
		RedisClient: rdb,
		SQSClient:   sqsClient,
		DB:          db,
		MaxPerTick:  maxPerTick,
	}
}

func (p *Processor) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⛔ scheduler 종료")
			return
		case <-ticker.C:
			campaignIDs, err := p.getActiveCampaignIDs(ctx)
			if err != nil {
				log.Printf("❌ 캠페인 조회 실패: %v", err)
				continue
			}
			for _, campaignID := range campaignIDs {
				p.processCampaign(ctx, campaignID)
			}
		}
	}
}

func (p *Processor) getActiveCampaignIDs(ctx context.Context) ([]int32, error) {
	query := `
		SELECT id FROM campaigns
		WHERE start_at <= NOW() AND end_at >= NOW()
	`

	rows, err := p.DB.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int32
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (p *Processor) processCampaign(ctx context.Context, campaignID int32) {
	key := fmt.Sprintf("coupon_queue:%d", campaignID)

	// ZRANGE로 상위 N개 추출
	users, err := p.RedisClient.Client.ZRange(ctx, key, 0, p.MaxPerTick-1).Result()
	if err != nil || len(users) == 0 {
		return
	}

	// ZREM으로 큐에서 제거
	if _, err := p.RedisClient.Client.ZRem(ctx, key, users).Result(); err != nil {
		log.Printf("❌ ZREM 실패: %v", err)
		return
	}

	// SQS로 전송
	for _, userID := range users {
		err := p.SQSClient.Enqueue(ctx, campaignID, userID)
		if err != nil {
			log.Printf("❌ SQS enqueue 실패: campaign=%d user=%s err=%v", campaignID, userID, err)
		} else {
			log.Printf("✅ SQS 전송: campaign=%d user=%s", campaignID, userID)
		}
	}
}
