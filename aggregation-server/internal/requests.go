package internal

type HelloRequest struct {
	Instance string `json:"instance"`
	Port     int    `json:"port"`
}

type ByeRequest struct {
	Instance string `json:"instance"`
}

type PutRequest struct {
	Instance  string            `json:"instance"`
	Operation string            `json:"op"`
	Sequence  int               `json:"seqno"`
	Value     map[string]string `json:"value"`
}
