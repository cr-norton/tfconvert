package types

type Options struct {
	StackName      string            `json:"stack_name"`
	ServiceName    string            `json:"service_name"`
	Region         string            `json:"region"`
	AdditionalTags map[string]string `json:"additional_tags"`
}
