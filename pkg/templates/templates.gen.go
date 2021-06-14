package templates

import "text/template"

var templates = map[string]string{"dynamodb.tmpl": `{{- $stackName := .Name}}
{{- $serviceName := .ServiceName}}
{{- $additionalTags := .AdditionalTags}}
{{- range .DynamoTables}}
resource "aws_dynamodb_table" "{{tfName .LogicalID}}" {
  name           = "{{.TableName}}"
  read_capacity  = {{.ProvisionedThroughput.ReadCapacityUnits}}
  write_capacity = {{.ProvisionedThroughput.WriteCapacityUnits}}
  hash_key       = "{{keySchemaElement .KeySchema "HASH"}}"
  {{- $rkey := keySchemaElement .KeySchema "RANGE"}}
  {{- if $rkey}}
  range_key    = "{{$rkey}}"

  {{- end}}
  {{- range .AttributeDefinitions}}
  
  attribute {
    name = "{{.AttributeName}}"
    type = "{{.AttributeType}}"
  }
  {{- end}}

  {{- range .GlobalSecondaryIndexes}}

  global_secondary_index {
    name               = "{{.IndexName}}"
    hash_key           = "{{keySchemaElement .KeySchema "HASH"}}"
    {{- $rkey := keySchemaElement .KeySchema "RANGE"}}
    {{- if $rkey}}
    range_key          = "{{$rkey}}"
    {{- end}}
    write_capacity     = {{.ProvisionedThroughput.WriteCapacityUnits}}
    read_capacity      = {{.ProvisionedThroughput.ReadCapacityUnits}}
    projection_type    = "{{.Projection.ProjectionType}}"
    {{- if eq .Projection.ProjectionType "INCLUDE"}}
    non_key_attributes = [.Projection.NonKeyAttributes] // TODO test this
    {{- end}}
  }
  {{- end}}

  tags = {
    Service          = "{{$serviceName}}"
    {{- range $key, $value := $additionalTags}}
    {{$key}}        = "{{$value}}"
    {{- end}}
  }
}

{{end}}`,
	"firehose.tmpl": `{{- $stackName := .Name}}
{{- $serviceName := .ServiceName}}
{{- $additionalTags := .AdditionalTags}}
{{- range .FirehoseDeliveryStreams}}
resource "aws_kinesis_firehose_delivery_stream" "{{tfName .LogicalID}}" {
  name        = "{{.DeliveryStreamName}}"
  destination = "extended_s3"
  {{- range .Destinations}}
  {{- if .ExtendedS3DestinationDescription}}
  {{- $dest := .ExtendedS3DestinationDescription}}
  extended_s3_configuration {
    role_arn   = "{{$dest.RoleARN}}"
    bucket_arn = "{{$dest.BucketARN}}"
    prefix     = "{{$dest.Prefix}}"
    error_output_prefix = "{{$dest.ErrorOutputPrefix}}"
    buffering_size      = {{$dest.BufferingHints.SizeInMBs}}
    buffering_interval  = {{$dest.BufferingHints.IntervalInSeconds}}
    compression_format  = "{{$dest.CompressionFormat}}"
  }
  {{- end}}
  {{- end}}

  tags = {
    Service          = "{{$serviceName}}"
    {{- range $key, $value := $additionalTags}}
    {{$key}}        = "{{$value}}"
    {{- end}}
  }
}

{{- end}}`,
	"iam.tmpl": `{{$serviceName := .ServiceName}}
{{$additionalTags := .AdditionalTags}}
{{- range .Roles}}
{{$id := tfName .LogicalID}}
resource "aws_iam_role" "{{tfName .LogicalID}}" {
  name               = "{{.RoleName}}"
  assume_role_policy = data.aws_iam_policy_document.{{tfName .LogicalID}}_assume_policy.json

  tags = {
    Service          = "{{$serviceName}}"
    {{- range $key, $value := $additionalTags}}
    {{$key}}        = "{{$value}}"
    {{- end}}
  }
}

{{- if .PolicyDocuments}}
{{- range $key, $value := .PolicyDocuments}}
{{- $pd := parsePolicyDocument $value}}

resource "aws_iam_role_policy" "{{$id}}" {
  name   = "{{$key}}" 
  role   = aws_iam_role.{{$id}}.id
  policy = data.aws_iam_policy_document.{{$id}}_role_policy.json
}

data "aws_iam_policy_document" "{{$id}}_role_policy" {
  {{- range $pd.Statement}}

  statement {
    {{- $actions := typeStringSlice .Action}}
    {{- $resources := typeStringSlice .Resource}}
    actions = [
      {{- range $actions}}
      "{{.}}",
      {{- end}}
    ]
    resources = [
      {{- range $resources}}
      "{{.}}",
      {{- end}}
    ]
  }
  {{- end}}
}
{{- end}}
{{- end}}
{{- $apd := parsePolicyDocument .AssumeRolePolicyDocument}}

data "aws_iam_policy_document" "{{tfName .LogicalID}}_assume_policy" {
  {{- range $apd.Statement}}
  statement {
    actions = [
      "{{.Action}}"
    ]
    principals {
    {{- if .Principal.Service}}
      type        = "Service"
      identifiers = [
        "{{.Principal.Service}}"
      ]
    {{- end}}
    }
    {{- if .Condition}}
    {{- if .Condition.StringEquals}}
    {{- range $key, $value := .Condition.StringEquals}}
    condition {
      test     = "StringEquals" 
      variable = "{{$key}}"
      values = [
        "{{$value}}"
      ]
    }
    {{- end}}
    {{- end}}
    {{- end}}
  }
  {{- end}}
}
{{- end}}`,
	"lambda.tmpl": `{{- $stack := .}}
{{- $stackName := .Name}}
{{- $serviceName := .ServiceName}}
{{- $additionalTags := .AdditionalTags}}
{{- range .LambdaFunctions}}
resource "aws_lambda_function" "{{tfName .LogicalID}}" {
  filename      = "lambda_function_payload.zip"
  function_name = "{{.FunctionName}}"
  role          = "{{.Role}}"
  handler       = "{{.Handler}}"
  runtime       = "{{.Runtime}}"
  memory_size   = {{.MemorySize}}
  timeout       = {{.Timeout}}

  tags = {
    Service          = "{{$serviceName}}"
    {{- range $key, $value := $additionalTags}}
    {{$key}}        = "{{$value}}"
    {{- end}}
  }

{{- if .Environment.Variables}}
  environment {
    variables = { 
      {{- range $key, $value := .Environment.Variables}} 
      "{{$key}}" = "{{$value}}" 
      {{- end}}
    }
  }
}
{{- end}}
{{- end}}
{{- range .LambdaEventSources}}

resource "aws_lambda_event_source_mapping" "{{tfName .LogicalID}}" {
  event_source_arn  = {{lookup $stack .EventSourceArn}}
  function_name     = {{lookup $stack .FunctionArn}}
}
{{- end}}

{{- range .LogGroups}}

resource "aws_cloudwatch_log_group" "{{tfName .LogicalID}}" {
  name = "{{.LogGroupName}}"
  retention_in_days = {{.RetentionInDays}}
}
{{- end}}`,
	"sns.tmpl": `{{- $stack := .}}
{{- $stackName := .Name}}
{{- $serviceName := .ServiceName}}
{{- $additionalTags := .AdditionalTags}}
{{- range .Topics}}
resource "aws_sns_topic" "{{tfName .LogicalID}}" {
  name = "{{.TopicName}}"
  tags = {
    Service          = "{{$serviceName}}"
    {{- range $key, $value := $additionalTags}}
    {{$key}}        = "{{$value}}"
    {{- end}}
  } 
}

{{- end}}
{{- range .TopicSubscriptions}}

resource "aws_sns_topic_subscription" "{{tfName .LogicalID}}" {
  topic_arn = {{lookup $stack (index .Attributes "TopicArn")}}
  protocol  = "{{index .Attributes "Protocol"}}"
  endpoint  = {{lookup $stack (index .Attributes "Endpoint")}}
}
{{- end}}`,
	"sqs.tmpl": `{{- $stack := .}}
{{- $stackName := .Name}}
{{- $serviceName := .ServiceName}}
{{- $additionalTags := .AdditionalTags}}
{{- range .Queues}}
resource "aws_sqs_queue" "{{tfName .LogicalID}}" {
  name = "{{.QueueName}}"
  {{- $rdp := .RedrivePolicy}}
  {{- if $rdp}}

  redrive_policy = jsonencode({
    deadLetterTargetArn = {{lookup $stack $rdp.DeadLetterTargetArn}}
    maxReceiveCount     = {{$rdp.MaxReceiveCount}}
  })
  {{- end}}
  {{- if .Policy}}

    policy = <<EOT
      {{formatJSON .Policy}}
    EOT
  {{- end}}

  tags = {
    Service          = "{{$stackName}}"
    {{- range $key, $value := $additionalTags}}
    {{$key}}        = "{{$value}}"
    {{- end}}
  }
}
{{end}}`,
}

// Parse parses declared templates.
func Parse(t *template.Template) (*template.Template, error) {
	for name, s := range templates {
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		if _, err := tmpl.Parse(s); err != nil {
			return nil, err
		}
	}
	return t, nil
}

