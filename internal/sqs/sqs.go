package sqs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSClient(ctx context.Context, queueURL string) (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("dummy", "dummy", "")),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 3)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("AWS config 로드 실패: %w", err)
	}

	// SQS 클라이언트 생성 시 endpoint 주입
	client := sqs.NewFromConfig(cfg, sqs.WithEndpointResolver(
		sqs.EndpointResolverFromURL("http://localstack:4566"),
	))

	return &SQSClient{
		client:   client,
		queueURL: queueURL,
	}, nil
}

type CouponMessage struct {
	CampaignID int32  `json:"campaign_id"`
	UserID     string `json:"user_id"`
}

func (s *SQSClient) Enqueue(ctx context.Context, campaignID int32, userID string) error {
	msg := CouponMessage{
		CampaignID: campaignID,
		UserID:     userID,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("메시지 직렬화 실패: %w", err)
	}

	_, err = s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &s.queueURL,
		MessageBody: aws.String(string(body)),
	})
	if err != nil {
		return fmt.Errorf("SQS 전송 실패: %w", err)
	}

	return nil
}
