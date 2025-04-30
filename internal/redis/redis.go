package redis

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

func NewRedisClient() *RedisClient {
	addr := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),	// no password set
		DB:       0,
	})
	return &RedisClient{Client: rdb}
}

func (r *RedisClient) IsUserRequested(ctx context.Context, campaignID, userID string) (bool, error) {
	key := fmt.Sprintf("coupon_requested:%s", campaignID)
	return r.Client.SIsMember(ctx, key, userID).Result()
}

func (r *RedisClient) EnqueueUser(ctx context.Context, campaignID, userID string, score float64) error {
	reqKey := fmt.Sprintf("coupon_requested:%s", campaignID)
	queueKey := fmt.Sprintf("coupon_queue:%s", campaignID)

	pipe := r.Client.TxPipeline()
	pipe.SAdd(ctx, reqKey, userID)
	pipe.ZAdd(ctx, queueKey, redis.Z{
		Score:  score,
		Member: userID,
	})
	_, err := pipe.Exec(ctx)
	return err
}

func (r *RedisClient) GetUserRank(ctx context.Context, campaignID, userID string) (int64, error) {
	key := fmt.Sprintf("coupon_queue:%s", campaignID)
	return r.Client.ZRank(ctx, key, userID).Result()
}
