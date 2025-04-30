package repository

import (
	"context"

	"github.com/yomapi/coupon-service/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type CampaignRepository struct {
	DB *pgxpool.Pool
}

func NewCampaignRepository(db *pgxpool.Pool) *CampaignRepository {
	return &CampaignRepository{DB: db}
}

func (r *CampaignRepository) InsertCampaign(ctx context.Context, c *model.Campaign) error {
	query := `
		INSERT INTO campaigns (name, coupon_count, start_at, end_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return r.DB.QueryRow(ctx, query,
		c.Name,
		c.CouponCount,
		c.StartAt,
		c.EndAt,
	).Scan(&c.ID)
}
