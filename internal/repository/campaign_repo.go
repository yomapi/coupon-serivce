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

func (r *CampaignRepository) GetCampaign(ctx context.Context, campaignID int32) (*model.Campaign, error) {
	query := `
		SELECT id, name
		FROM campaigns
		WHERE id = $1
	`

	var campaign model.Campaign
	err := r.DB.QueryRow(ctx, query, campaignID).Scan(&campaign.ID, &campaign.Name)
	if err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (r *CampaignRepository) GetCouponCodesByCampaign(ctx context.Context, campaignID int32) ([]string, error) {
	query := `
		SELECT code
		FROM coupons
		WHERE campaign_id = $1
	`

	rows, err := r.DB.Query(ctx, query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var codes []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		codes = append(codes, code)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return codes, nil
}
