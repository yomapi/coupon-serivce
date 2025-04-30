package service

import (
	"context"

	couponv1 "github.com/yomapi/coupon-service/gen/go/coupon/v1"

	"github.com/yomapi/coupon-service/internal/db"
	"github.com/yomapi/coupon-service/internal/model"
)

type CampaignService struct {
	Repo *db.CampaignRepository
}

func NewCampaignService(repo *db.CampaignRepository) *CampaignService {
	return &CampaignService{Repo: repo}
}

func (s *CampaignService) CreateCampaign(ctx context.Context, req *couponv1.CreateCampaignRequest) (*model.Campaign, error) {
	campaign := &model.Campaign{
		Name:        req.Name,
		CouponCount: int(req.CouponCount),
		StartAt:     req.StartAt.AsTime(),
		EndAt:       req.EndAt.AsTime(),
	}
	err := s.Repo.InsertCampaign(ctx, campaign)
	if err != nil {
		return nil, err
	}
	return campaign, nil
}
