package service

import (
	"context"
	"errors"
	"time"

	"github.com/yomapi/coupon-service/internal/repository"

	couponv1 "github.com/yomapi/coupon-service/gen/go/coupon/v1"

	"github.com/bufbuild/connect-go"
)


type CouponService struct {
	queueRepo repository.CouponQueueRepository
	couponRepo *repository.CouponRepository
}

func NewCouponService(queueRepo repository.CouponQueueRepository, couponRepo *repository.CouponRepository) *CouponService {
	return &CouponService{
		queueRepo: queueRepo,
		couponRepo: couponRepo,
	}
}

func (s *CouponService) IssueCoupon(ctx context.Context, req *couponv1.IssueCouponRequest) (*connect.Response[couponv1.IssueCouponResponse], error) {
	campaignID := req.CampaignId
	userID := req.UserId

	
	requested, err := s.queueRepo.IsUserRequested(ctx, campaignID, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	if requested {
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("이미 쿠폰을 요청했습니다"))
	}
	
	coupon, err := s.couponRepo.GetCouponByUserAndCampaign(ctx, userID, campaignID)
	if err != nil {
			return nil, connect.NewError(connect.CodeInternal, err)
	} 
	if coupon != nil {
		return nil, connect.NewError(connect.CodeAlreadyExists, errors.New("이미 쿠폰을 발급받았습니다"))
	}

	
	score := float64(time.Now().UnixNano())
	rank, err := s.queueRepo.EnqueueUser(ctx, campaignID, userID, score)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	
	resp := connect.NewResponse(&couponv1.IssueCouponResponse{
		Rank: rank, 
	})
	return resp, nil
}
