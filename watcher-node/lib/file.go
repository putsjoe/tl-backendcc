package lib

type BaseMessage struct {
	Instance string `json:"instance"`
}

type ListResponse struct {
	BaseMessage
	Files    []FileMetadata `json:"files"`
	Sequence int            `json:"seqno"`
}

type FileMetadata struct {
	Filename string `json:"filename"`
}

type PatchOperation struct {
	BaseMessage
	Op       string       `json:"op"`
	Value    FileMetadata `json:"value"`
	Sequence int          `json:"seqno"`
}

type HelloOperation struct {
	BaseMessage
	Port uint `json:"port"`
}

type ByeOperation struct {
	BaseMessage
}
