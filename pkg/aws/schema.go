package aws

import "github.com/cr-norton/tfconvert/pkg/types"

type Stack struct {
	Name           string
	ServiceName    string
	AdditionalTags map[string]string
	Index          map[string]types.Resource
	*StackResources
}

func (s Stack) Lookup(id string) *types.Resource {
	if r, has := s.Index[id]; has {
		return &r
	}
	return nil
}

func (s Stack) Resources() []types.Resource {
	resources := []types.Resource{}
	for _, res := range s.Index {
		resources = append(resources, res)
	}
	return resources
}

type StackResources struct {
	DynamoTables            []DynamoTable
	Roles                   []Role
	FirehoseDeliveryStreams []FirehoseDeliveryStream
	LambdaFunctions         []LambdaFunctionConfiguration
	LambdaEventSources      []LambdaEventSource
	LogGroups               []LogGroup
	Queues                  []Queue
	Topics                  []Topic
	TopicSubscriptions      []TopicSubscription
}

func index(stack *StackResources) map[string]types.Resource {
	index := map[string]types.Resource{}
	for _, r := range stack.DynamoTables {
		index[r.Key()] = r.Resource()
	}
	for _, r := range stack.Roles {
		index[r.Key()] = r.Resource()
	}
	for _, r := range stack.LambdaFunctions {
		index[r.Key()] = r.Resource()
	}
	for _, r := range stack.Queues {
		index[r.Key()] = r.Resource()
	}
	for _, r := range stack.Topics {
		index[r.Key()] = r.Resource()
	}
	return index
}
