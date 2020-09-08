package solr

import (
	"encoding/json"
)

// DefaultDocumentParser - Default parser
type DefaultDocumentParser struct {
}

// Parse - parses a document and return a default struct
func (p *DefaultDocumentParser) Parse(raw []byte) (interface{}, error) {

	var resp responseRaw

	err := json.Unmarshal(raw, &resp)
	if err != nil {
		return nil, err
	}

	return resp.Data.Documents, err

}

// DefaultDocumentWriter - default update document writer
type DefaultDocumentWriter struct {
}

// Writer - default document writer
func (u *DefaultDocumentWriter) Writer(payload interface{}) ([]byte, error) {

	byte, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return byte, nil
}
