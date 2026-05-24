package sqs

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

func NewSQSClient(sdkConfig aws.Config) *sqs.Client {
	return sqs.NewFromConfig(sdkConfig)
}
