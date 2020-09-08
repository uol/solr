package solr

// DocumentWriter - a generic document writer
type DocumentWriter interface {

	// Writer - writes a document
	Writer(payload interface{}) ([]byte, error)
}

// DocumentParser - a generic document parser
type DocumentParser interface {

	// Parse - parses the pure document input from JSON
	Parse(documents []byte) (interface{}, error)
}
