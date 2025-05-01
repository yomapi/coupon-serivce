package model

import "time"

type Coupon struct {
	ID          int32
	Code        string
	UserID 			int32
	CampaignID	int32
	IssuedAt    time.Time
	ExpireAt    time.Time
}