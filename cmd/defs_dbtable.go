package cmd

/*
tDbRecord defines arbitrary structure for any key/value record.
*/
type tDbRecord struct {
	Key   string `json:"key"`
	Value []byte `json:"value,omitempty"`
}
