package solr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostSolr(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()

	createCollection(t, keyset)

	tags := make(map[string]string)
	tags["host"] = "host"

	expected, err := makeDocs(metric, keyset, "", tags, 10, true)
	if !assert.NoError(t, err) {
		return
	}

	actual := httpSearchSolr(keyset, metric, 10)

	testHTTPDocument(t, expected, actual)

}
