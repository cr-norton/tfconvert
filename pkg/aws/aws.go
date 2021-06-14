package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	cloudformationTypes "github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/firehose"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/cr-norton/tfconvert/pkg/types"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	cloudformation *cloudformation.Client
	dynamodb       *dynamodb.Client
	iam            *iam.Client
	firehose       *firehose.Client
	lambda         *lambda.Client
	logs           *cloudwatchlogs.Client
	sqs            *sqs.Client
	sns            *sns.Client
}

// New ...
func New(region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &Client{
		cloudformation: cloudformation.NewFromConfig(cfg),
		dynamodb:       dynamodb.NewFromConfig(cfg),
		iam:            iam.NewFromConfig(cfg),
		firehose:       firehose.NewFromConfig(cfg),
		lambda:         lambda.NewFromConfig(cfg),
		logs:           cloudwatchlogs.NewFromConfig(cfg),
		sqs:            sqs.NewFromConfig(cfg),
		sns:            sns.NewFromConfig(cfg),
	}, nil
}

func (aws *Client) GetStack(ctx context.Context, options types.Options) (*Stack, error) {
	resources, err := aws.GetStackResources(ctx, options.StackName)
	if err != nil {
		return nil, err
	}

	stackres := &StackResources{}
	for _, r := range resources {
		switch *r.ResourceType {
		case "AWS::DynamoDB::Table":
			table, err := aws.GetDynamoTable(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.DynamoTables = append(stackres.DynamoTables, *table)
		case "AWS::IAM::Role":
			role, err := aws.GetRole(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, errors.Wrap(err, "unable to get IAM role")
			}
			stackres.Roles = append(stackres.Roles, *role)
		case "AWS::KinesisFirehose::DeliveryStream":
			stream, err := aws.GetFirehoseDeliveryStream(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.FirehoseDeliveryStreams = append(stackres.FirehoseDeliveryStreams, *stream)
		case "AWS::Lambda::Function":
			function, err := aws.GetLambdaFunction(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.LambdaFunctions = append(stackres.LambdaFunctions, *function)
		case "AWS::Lambda::EventSourceMapping":
			event, err := aws.GetLambdaEventSource(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.LambdaEventSources = append(stackres.LambdaEventSources, *event)
		case "AWS::Logs::LogGroup":
			logs, err := aws.GetLogGroup(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.LogGroups = append(stackres.LogGroups, *logs)
		case "AWS::SQS::Queue":
			queue, err := aws.GetQueue(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.Queues = append(stackres.Queues, *queue)
		case "AWS::SNS::Topic":
			topic, err := aws.GetTopic(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.Topics = append(stackres.Topics, *topic)
		case "AWS::SNS::Subscription":
			subscription, err := aws.GetTopicSubscription(ctx, *r.LogicalResourceId, *r.PhysicalResourceId)
			if err != nil {
				return nil, err
			}
			stackres.TopicSubscriptions = append(stackres.TopicSubscriptions, *subscription)
		case "AWS::ApiGateway::Authorizer", "AWS::ApiGateway::Deployment", "AWS::ApiGateway::Method", "AWS::ApiGateway::Resource", "AWS::ApiGateway::RestApi":
			continue
		case "AWS::ApiGatewayV2::Api", "AWS::ApiGatewayV2::Integration", "AWS::ApiGatewayV2::Route", "AWS::ApiGatewayV2::Stage":
			continue
		default:
			log.WithFields(log.Fields{
				"resource_type": *r.ResourceType,
				"logical_id":    *r.LogicalResourceId,
				"physical_id":   *r.PhysicalResourceId,
			}).Warn("unsupported aws resource")
		}
	}

	stack := &Stack{
		Name:           options.StackName,
		ServiceName:    options.ServiceName,
		AdditionalTags: options.AdditionalTags,
		Index:          index(stackres),
		StackResources: stackres,
	}
	return stack, nil
}

func (aws *Client) GetStackResources(ctx context.Context, stackName string) ([]cloudformationTypes.StackResourceSummary, error) {
	input := &cloudformation.ListStackResourcesInput{
		StackName: &stackName,
	}
	resources := []cloudformationTypes.StackResourceSummary{}
	for {
		res, err := aws.cloudformation.ListStackResources(ctx, input)
		if err != nil {
			return nil, err
		}

		resources = append(resources, res.StackResourceSummaries...)
		if res.NextToken != nil {
			input.NextToken = res.NextToken
		} else {
			break
		}
	}
	return resources, nil
}

func (aws *Client) GetDynamoTable(ctx context.Context, logicalID string, tableName string) (*DynamoTable, error) {
	table, err := aws.dynamodb.DescribeTable(ctx, &dynamodb.DescribeTableInput{
		TableName: &tableName,
	})
	if err != nil {
		return nil, err
	}
	return &DynamoTable{
		LogicalID:        logicalID,
		TableDescription: *table.Table,
	}, nil
}

func (aws *Client) GetRole(ctx context.Context, logicalID string, roleName string) (*Role, error) {
	r := &Role{LogicalID: logicalID, PolicyDocuments: map[string]string{}}

	role, err := aws.iam.GetRole(ctx, &iam.GetRoleInput{
		RoleName: &roleName,
	})
	if err != nil {
		return nil, err
	}
	r.Role = *role.Role

	rolePolicies, err := aws.iam.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: &roleName,
	})
	if err != nil {
		return nil, err
	}

	for _, pname := range rolePolicies.PolicyNames {
		policy, err := aws.iam.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
			RoleName:   &roleName,
			PolicyName: &pname,
		})
		if err != nil {
			return nil, err
		}
		r.PolicyDocuments[pname] = *policy.PolicyDocument
	}

	attachedPolicies, err := aws.iam.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
		RoleName: &roleName,
	})
	if err != nil {
		return nil, err
	}

	for _, apolicy := range attachedPolicies.AttachedPolicies {
		policy, err := aws.iam.GetPolicy(ctx, &iam.GetPolicyInput{
			PolicyArn: apolicy.PolicyArn,
		})
		if err != nil {
			return nil, err
		}
		r.AttachedPolicies = append(r.AttachedPolicies, *policy.Policy)
	}

	return r, nil
}

func (aws *Client) GetFirehoseDeliveryStream(ctx context.Context, logicalID string, deliveryStreamName string) (*FirehoseDeliveryStream, error) {
	res, err := aws.firehose.DescribeDeliveryStream(ctx, &firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: &deliveryStreamName,
	})
	if err != nil {
		return nil, err
	}
	return &FirehoseDeliveryStream{
		LogicalID:                 logicalID,
		DeliveryStreamDescription: *res.DeliveryStreamDescription,
	}, nil
}

func (aws *Client) GetLambdaFunction(ctx context.Context, logicalID string, functionName string) (*LambdaFunctionConfiguration, error) {
	res, err := aws.lambda.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: &functionName,
	})
	if err != nil {
		return nil, err
	}
	return &LambdaFunctionConfiguration{
		LogicalID:             logicalID,
		FunctionConfiguration: *res.Configuration,
	}, nil
}

func (aws *Client) GetLambdaEventSource(ctx context.Context, logicalID string, uuid string) (*LambdaEventSource, error) {
	res, err := aws.lambda.GetEventSourceMapping(ctx, &lambda.GetEventSourceMappingInput{
		UUID: &uuid,
	})
	if err != nil {
		return nil, err
	}
	return &LambdaEventSource{
		LogicalID: logicalID, // TODO move this to types.go
		EventSourceMappingConfiguration: lambdaTypes.EventSourceMappingConfiguration{
			BatchSize:                      res.BatchSize,
			BisectBatchOnFunctionError:     res.BisectBatchOnFunctionError,
			DestinationConfig:              res.DestinationConfig,
			EventSourceArn:                 res.EventSourceArn,
			FunctionArn:                    res.FunctionArn,
			FunctionResponseTypes:          res.FunctionResponseTypes,
			LastModified:                   res.LastModified,
			LastProcessingResult:           res.LastProcessingResult,
			MaximumBatchingWindowInSeconds: res.MaximumBatchingWindowInSeconds,
			MaximumRecordAgeInSeconds:      res.MaximumRecordAgeInSeconds,
			MaximumRetryAttempts:           res.MaximumRetryAttempts,
			ParallelizationFactor:          res.ParallelizationFactor,
			Queues:                         res.Queues,
			SelfManagedEventSource:         res.SelfManagedEventSource,
			SourceAccessConfigurations:     res.SourceAccessConfigurations,
			StartingPosition:               res.StartingPosition,
			StartingPositionTimestamp:      res.StartingPositionTimestamp,
			State:                          res.State,
			StateTransitionReason:          res.StateTransitionReason,
			Topics:                         res.Topics,
			TumblingWindowInSeconds:        res.TumblingWindowInSeconds,
			UUID:                           res.UUID,
		},
	}, nil
}

func (aws *Client) GetLogGroup(ctx context.Context, logicalID string, logGroupName string) (*LogGroup, error) {
	res, err := aws.logs.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: &logGroupName,
	})
	if err != nil || len(res.LogGroups) == 0 {
		return nil, err
	}
	return &LogGroup{
		LogicalID: logicalID,
		LogGroup:  res.LogGroups[0],
	}, nil
}

func (aws *Client) GetQueue(ctx context.Context, logicalID string, queueURL string) (*Queue, error) {
	res, err := aws.sqs.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []sqsTypes.QueueAttributeName{"All"},
	})
	if err != nil {
		return nil, err
	}
	return &Queue{
		LogicalID:  logicalID,
		Attributes: res.Attributes,
	}, nil
}

func (aws *Client) GetTopic(ctx context.Context, logicalID string, topicArn string) (*Topic, error) {
	res, err := aws.sns.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: &topicArn,
	})
	if err != nil {
		return nil, err
	}
	return &Topic{
		LogicalID:  logicalID,
		Attributes: res.Attributes,
	}, nil
}

func (aws *Client) GetTopicSubscription(ctx context.Context, logicalID string, subscriptionArn string) (*TopicSubscription, error) {
	res, err := aws.sns.GetSubscriptionAttributes(ctx, &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: &subscriptionArn,
	})
	if err != nil {
		return nil, err
	}
	return &TopicSubscription{
		LogicalID:  logicalID,
		Attributes: res.Attributes,
	}, nil
}
