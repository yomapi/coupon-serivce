package repository

import (
	"context"
	"fmt"

	rds "github.com/redis/go-redis/v9"
	"github.com/yomapi/coupon-service/internal/redis"
)

type CouponQueueRepository interface {
	IsUserRequested(ctx context.Context, campaignID int32, userID int32) (bool, error)
	EnqueueUser(ctx context.Context, campaignID int32, userID int32, score float64) (int64, error)
	GetUserRank(ctx context.Context, campaignID int32, userID int32) (int64, error)
}

type couponQueueRepository struct {
	rdb *redis.RedisClient
}

func NewCouponQueueRepository(rdb *redis.RedisClient) CouponQueueRepository {
	return &couponQueueRepository{rdb: rdb}
}

func (r *couponQueueRepository) IsUserRequested(ctx context.Context, campaignID int32, userID int32) (bool, error) {
	campaignKey := fmt.Sprintf("coupon_requested:%d", campaignID)
	userKey := fmt.Sprintf("%d", userID)
	return r.rdb.Client.SIsMember(ctx, campaignKey, userKey).Result()
}

func (r *couponQueueRepository) EnqueueUser(ctx context.Context, campaignID int32, userID int32, score float64) (int64, error) {
	requestedKey := fmt.Sprintf("coupon_requested:%d", campaignID)
	queueKey := fmt.Sprintf("coupon_queue:%d", campaignID)
	userKey := fmt.Sprintf("%d", userID)

	pipe := r.rdb.Client.TxPipeline()
	pipe.SAdd(ctx, requestedKey, userKey)
	pipe.ZAdd(ctx, queueKey, rds.Z{
		Score:  score,
		Member: userKey,
	})
	_, err := pipe.Exec(ctx)
	if err != nil {
		return -1, err
	}

	rank, err := r.rdb.Client.ZRank(ctx, queueKey, userKey).Result()
	return rank, err
}

func (r *couponQueueRepository) GetUserRank(ctx context.Context, campaignID int32, userID int32) (int64, error) {
	key := fmt.Sprintf("coupon_queue:%d", campaignID)
	userKey := fmt.Sprintf("%d", userID)
	return r.rdb.Client.ZRank(ctx, key, userKey).Result()
}
