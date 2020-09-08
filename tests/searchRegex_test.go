package solr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestSearchSolrLibRegex(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()

	createCollection(t, keyset)

	tags := make(map[string]string)
	tags["host"] = "host"

	expected, err := makeDocs(metric, keyset, "", tags, 10, false)
	if !assert.NoError(t, err) {
		return
	}

	names := []string{"andre", "bernardo", "carlos", "deise", "edmundo", "fabio", "geraldo", "helcio", "ivanir", "joao"}

	expectedNames := makeDocsByArray(metric, keyset, "", "name", names, false)
	if !assert.NoError(t, err) {
		return
	}

	var queries = []string{"tag_value:/host[0-9]+/", "tag_value:/host.*/", "tag_value:/host[0-9]{1,2}/",
		"tag_value:/host1|host2|host3|host4|host5|host6|host7|host8|host9|host10/",
		"tag_value:/hos[a-z].*/"}

	for _, q := range queries {
		searchParams := &solr.SearchParams{
			Q:    q,
			Rows: 10,
		}

		res, err := defaultInstance.Search(searchParams, keyset)
		if !assert.NoError(t, err) {
			return
		}

		actual := res.Docs.([]solr.DocumentRaw)

		testDocumentRaw(t, expected, actual)

	}

	searchParams := &solr.SearchParams{
		Q:    "tag_value:/a.*|b.*|c.*|d.*|e.*|f.*|g.*|hel.*|i.*|j.*/",
		Rows: 10,
	}

	res, err := defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	actualNames := res.Docs.([]solr.DocumentRaw)

	testDocumentRaw(t, expectedNames, actualNames)

}
