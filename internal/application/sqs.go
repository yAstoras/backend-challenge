package application

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSClient struct {
	client *sqs.Client
}

func (s *SQSClient) QueueURL(ctx context.Context, queueName string) (string, error) {
	result, err := s.client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return "", fmt.Errorf("get SQS queue URL: %w", err)
	}

	return aws.ToString(result.QueueUrl), nil
}

func NewSQSClient() (*SQSClient, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				"test",
				"test",
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	endpoint := "http://localhost:4566"

	client := sqs.NewFromConfig(cfg, func(options *sqs.Options) {
		options.BaseEndpoint = aws.String(endpoint)
	})

	return &SQSClient{
		client: client,
	}, nil
}

func (s *SQSClient) SendMessage(
	ctx context.Context,
	queueURL string,
	body string,
) error {
	_, err := s.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(body),
	})
	if err != nil {
		return fmt.Errorf("send SQS message: %w", err)
	}

	return nil
}

func (s *SQSClient) ReceiveMessage(
	ctx context.Context,
	queueURL string,
) (string, string, string, error) {
	result, err := s.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     2,
	})
	if err != nil {
		return "", "", "", fmt.Errorf("receive SQS message: %w", err)
	}

	if len(result.Messages) == 0 {
		return "", "", "", fmt.Errorf("no SQS messages available")
	}

	message := result.Messages[0]

	return aws.ToString(message.MessageId),
		aws.ToString(message.Body),
		aws.ToString(message.ReceiptHandle),
		nil
}

func (s *SQSClient) DeleteMessage(
	ctx context.Context,
	queueURL string,
	receiptHandle string,
) error {
	_, err := s.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("delete SQS message: %w", err)
	}

	return nil
}

func (s *SQSClient) Client() *sqs.Client {
	return s.client
}

var _ = aws.String

func NewSQSQueueURL(client *SQSClient) (string, error) {
	return client.QueueURL(context.Background(), "wager-transactions")
}
