package types

type Resource struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
	ImportKey  string `json:"import_key"`
	OutputKey  string `json:"output_key"`
}
