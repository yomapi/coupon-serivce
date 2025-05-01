package service

import (
	"context"

	couponv1 "github.com/yomapi/coupon-service/gen/go/coupon/v1"

	"github.com/yomapi/coupon-service/internal/model"
	repository "github.com/yomapi/coupon-service/internal/repository"
)

type CampaignService struct {
	Repo *repository.CampaignRepository
}

func NewCampaignService(repo *repository.CampaignRepository) *CampaignService {
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

func (s *CampaignService) GetCampaignWithCoupons(ctx context.Context, campaignID int32) (*model.CampaignWithCoupons, error) {
	campaign, err := s.Repo.GetCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	coupons, err := s.Repo.GetCouponCodesByCampaign(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	return &model.CampaignWithCoupons{
		ID:          campaign.ID,
		Name:        campaign.Name,
		CouponCodes: coupons,
	}, nil
}

