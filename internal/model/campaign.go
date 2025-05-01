package model

import "time"

type Campaign struct {
	ID          int32
	Name        string
	CouponCount int
	StartAt     time.Time
	EndAt       time.Time
}

type CampaignWithCoupons struct {
	ID          int32
	Name        string
	CouponCodes []string
}
