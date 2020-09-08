package solr

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestPostUpdate(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()
	createCollection(t, keyset)

	var newDoc []DefaultDocument
	creation := time.Now().Format("2006-01-02T15:04:05Z")
	newDoc = append(newDoc, DefaultDocument{
		ID:           "1",
		Metric:       fmt.Sprintf("%s", metric),
		Type:         "meta",
		ParentDoc:    true,
		TagKey:       "host",
		TagValue:     "host1",
		CreationDate: creation,
	})

	httpPostDocs(keyset, newDoc)

	var newDoc2 []DefaultDocument
	newDoc2 = append(newDoc2, DefaultDocument{
		ID:           "1",
		Metric:       fmt.Sprintf("%s", metric),
		Type:         "meta",
		ParentDoc:    true,
		TagKey:       "host",
		TagValue:     "host2",
		CreationDate: creation,
	})

	solrLibPost(keyset, newDoc2)

	params := *&solr.SearchParams{
		Q:    "*:*",
		Rows: 10,
	}
	res, err := defaultInstance.Search(&params, keyset)
	if !assert.NoError(t, err) {
		return
	}

	actual := res.Docs.([]solr.DocumentRaw)

	testDocumentRaw(t, newDoc2, actual)

}
