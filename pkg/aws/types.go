package aws

import (
	"encoding/json"
	"fmt"
	"strings"

	logs "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	dynamodb "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	firehose "github.com/aws/aws-sdk-go-v2/service/firehose/types"
	iam "github.com/aws/aws-sdk-go-v2/service/iam/types"
	lambda "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/cr-norton/tfconvert/pkg/types"
)

// todo refactor handling

type DynamoTable struct {
	LogicalID string
	dynamodb.TableDescription
}

func (d DynamoTable) Key() string {
	return *d.TableDescription.TableArn
}

func (d DynamoTable) Resource() types.Resource {
	return types.Resource{
		Type:       "aws_dynamodb_table",
		Identifier: d.LogicalID,
		ImportKey:  *d.TableName,
		OutputKey:  "arn",
	}
}

type Role struct {
	LogicalID        string
	PolicyDocuments  map[string]string
	AttachedPolicies []iam.Policy
	iam.Role
}

func (r Role) Key() string {
	return *r.Role.Arn
}

func (r Role) Resource() types.Resource {
	return types.Resource{
		Type:       "aws_iam_role",
		Identifier: r.LogicalID,
		ImportKey:  *r.Role.RoleName,
		OutputKey:  "arn",
	}
}

type FirehoseDeliveryStream struct {
	LogicalID string
	firehose.DeliveryStreamDescription
}

func (f FirehoseDeliveryStream) Key() string {
	return *f.DeliveryStreamDescription.DeliveryStreamARN
}

func (f FirehoseDeliveryStream) Resource() types.Resource {
	return types.Resource{
		Type:       "aws_kinesis_firehose_delivery_stream",
		Identifier: f.LogicalID,
		ImportKey:  *f.DeliveryStreamARN,
		OutputKey:  "arn",
	}
}

type LambdaFunctionConfiguration struct {
	LogicalID string
	lambda.FunctionConfiguration
}

func (l LambdaFunctionConfiguration) Resource() types.Resource {
	return types.Resource{
		Type:       "lambda_function",
		Identifier: l.LogicalID,
		ImportKey:  *l.FunctionName,
		OutputKey:  "arn",
	}
}

func (l LambdaFunctionConfiguration) Key() string {
	return *l.FunctionConfiguration.FunctionArn
}

type LambdaEventSource struct {
	LogicalID string
	lambda.EventSourceMappingConfiguration
}

func (l LambdaEventSource) Resource() types.Resource {
	return types.Resource{
		Type:       "lambda_event_source_mapping",
		Identifier: l.LogicalID,
		ImportKey:  *l.UUID,
		OutputKey:  "arn",
	}
}

type LogGroup struct {
	LogicalID string
	logs.LogGroup
}

func (l LogGroup) Resource() types.Resource {
	return types.Resource{
		Type:       "aws_cloudwatch_log_group",
		Identifier: l.LogicalID,
		ImportKey:  *l.LogGroupName,
		OutputKey:  "arn",
	}
}

type Topic struct {
	LogicalID  string
	Attributes map[string]string
}

func (t Topic) Key() string {
	return t.Attributes["TopicArn"]
}

func (t Topic) Resource() types.Resource {
	return types.Resource{
		Type:       "aws_sns_topic",
		Identifier: t.LogicalID,
		ImportKey:  t.Attributes["TopicArn"],
		OutputKey:  "arn",
	}
}

func (t Topic) TopicName() string {
	s := strings.Split(t.Attributes["TopicArn"], ":")
	return s[len(s)-1]
}

type TopicSubscription struct {
	LogicalID  string
	Attributes map[string]string
}

func (t TopicSubscription) Resource() types.Resource {
	return types.Resource{
		Type:       "sns_topic_subscription",
		Identifier: t.LogicalID,
		ImportKey:  t.Attributes["SubscriptionArn"],
		OutputKey:  "arn",
	}
}

type Queue struct {
	LogicalID  string
	Attributes map[string]string
}

func (q Queue) Key() string {
	return q.Attributes["QueueArn"]
}

func (q Queue) Resource() types.Resource {
	return types.Resource{
		Type:       "aws_sqs_queue",
		Identifier: q.LogicalID,
		ImportKey:  q.QueueUrl(),
		OutputKey:  "arn",
	}
}

func (q Queue) QueueName() string {
	s := strings.Split(q.Attributes["QueueArn"], ":")
	return s[len(s)-1]
}

func (q Queue) QueueUrl() string {
	s := strings.Split(q.Attributes["QueueArn"], ":")
	region, account, name := s[3], s[4], s[5]
	return fmt.Sprintf("https://%s.queue.amazonaws.com/%s/%s", region, account, name)
}

func (q Queue) Policy() string {
	return q.Attributes["Policy"]
}

func (q Queue) RedrivePolicy() *RedrivePolicy {
	pjson, has := q.Attributes["RedrivePolicy"]
	if !has {
		return nil
	}
	var policy RedrivePolicy
	if err := json.Unmarshal([]byte(pjson), &policy); err != nil {
		return nil
	}
	return &policy
}

type RedrivePolicy struct {
	DeadLetterTargetArn string `json:"deadLetterTargetArn"`
	MaxReceiveCount     int    `json:"maxReceiveCount"`
}
