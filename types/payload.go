package types

type Payload struct {
	Metadata map[string]interface{} `json:"metadata"`
	Data     interface{}            `json:"data"`
}
