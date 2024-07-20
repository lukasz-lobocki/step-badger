package cmd

type tDbRecord struct {
	Key   string `json:"key"`
	Value []byte `json:"value,omitempty"`
}
