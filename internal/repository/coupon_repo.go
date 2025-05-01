package repository

import (
	"context"
	"errors"

	"github.com/yomapi/coupon-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CouponRepository struct {
	DB *pgxpool.Pool
}

func NewCouponRepository(db *pgxpool.Pool) *CouponRepository {
	return &CouponRepository{DB: db}
}

func (r *CouponRepository) GetCouponByUserAndCampaign(ctx context.Context, userID int32, campaignID int32) (*model.Coupon, error) {
	query := `
		SELECT id, code, user_id, campaign_id, issued_at, expired_at
		FROM coupons
		WHERE user_id = $1 AND campaign_id = $2
	`
	row := r.DB.QueryRow(ctx, query, userID, campaignID)

	var c model.Coupon
	err := row.Scan(&c.ID, &c.Code, &c.UserID, &c.CampaignID, &c.IssuedAt, &c.ExpireAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil 
		}
		return nil, err
	}
	return &c, nil
}
