package main

import (
	"context"
	"log"
	"net/http"

	"github.com/bufbuild/connect-go"
	couponv1 "github.com/yomapi/coupon-service/gen/go/coupon/v1"
	"github.com/yomapi/coupon-service/gen/go/coupon/v1/couponv1connect"
)

type couponService struct{}

func (s *couponService) IssueCoupon(
    ctx context.Context,
    req *connect.Request[couponv1.IssueCouponRequest],
) (*connect.Response[couponv1.IssueCouponResponse], error) {
    log.Printf("Issuing coupon for user %s in campaign %s", req.Msg.UserId, req.Msg.CampaignId)

    resp := connect.NewResponse(&couponv1.IssueCouponResponse{
        CouponCode: "DUMMY123",
    })

    return resp, nil
}

func main() {
    mux := http.NewServeMux()

    path, handler := couponv1connect.NewCouponServiceHandler(&couponService{})
    mux.Handle(path, handler)

    http.ListenAndServe(":8080", mux)
}
