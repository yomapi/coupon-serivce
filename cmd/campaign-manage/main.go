package main

import (
	"context"
	"log"
	"net/http"

	"github.com/yomapi/coupon-service/gen/go/coupon/v1/couponv1connect"
	"github.com/yomapi/coupon-service/internal/api"
	"github.com/yomapi/coupon-service/internal/db"
	"github.com/yomapi/coupon-service/internal/repository"
	"github.com/yomapi/coupon-service/internal/service"
)

func main() {
	ctx := context.Background()

	pool, err := db.NewPostgresPool(ctx)
	if err != nil {
		log.Fatalf("DB 연결 실패: %v", err)
	}
	defer pool.Close()

	repo := repository.NewCampaignRepository(pool)
	campaignService := service.NewCampaignService(repo)
	handler := api.NewCampaignHandler(campaignService)

	mux := http.NewServeMux()
	path, h := couponv1connect.NewCouponServiceHandler(handler)
	mux.Handle(path, h)

	log.Println("campaign-manage server running at http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("run server failed: %v", err)
	}
}
