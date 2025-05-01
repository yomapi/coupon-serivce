package api

import (
	"context"
	"errors"
	"log"

	couponv1 "github.com/yomapi/coupon-service/gen/go/coupon/v1"

	"github.com/yomapi/coupon-service/gen/go/coupon/v1/couponv1connect"
	"github.com/yomapi/coupon-service/internal/service"

	"github.com/bufbuild/connect-go"
)

type CampaignHandler struct {
	CampaignService *service.CampaignService
	CouponService   *service.CouponService
}

func NewCampaignHandler(
	campaignService *service.CampaignService,
	couponService *service.CouponService,
) couponv1connect.CouponServiceHandler {
	return &CampaignHandler{
		CampaignService: campaignService,
		CouponService:   couponService,
	}
}

func (h *CampaignHandler) CreateCampaign(
	ctx context.Context,
	req *connect.Request[couponv1.CreateCampaignRequest],
) (*connect.Response[couponv1.CreateCampaignResponse], error) {
	log.Printf("CreateCampaign: %s", req.Msg.Name)

	if req.Msg.Name == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("name is required"))
	}
	if req.Msg.CouponCount <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("coupon_count must be greater than 0"))
	}
	if req.Msg.StartAt == nil || req.Msg.EndAt == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("start_at and end_at are required"))
	}
	startTime := req.Msg.StartAt.AsTime()
	endTime := req.Msg.EndAt.AsTime()
	if !startTime.Before(endTime) {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("start_at must be before end_at"))
	}

	campaign, err := h.CampaignService.CreateCampaign(ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&couponv1.CreateCampaignResponse{
		CampaignId: campaign.ID,
		Name:       campaign.Name,
	})
	return res, nil
}


func (h *CampaignHandler) IssueCoupon(
	ctx context.Context,
	req *connect.Request[couponv1.IssueCouponRequest],
) (*connect.Response[couponv1.IssueCouponResponse], error) {
	log.Printf("IssueCoupon: campaignID=%d, userID=%d", req.Msg.CampaignId, req.Msg.UserId)

	if req.Msg.CampaignId <= 0 || req.Msg.UserId <= 0 {
		return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("campaign_id and user_id must be positive"))
	}

	return h.CouponService.IssueCoupon(ctx, req.Msg)
}
