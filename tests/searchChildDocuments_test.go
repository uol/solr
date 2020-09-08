package solr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestChildDocuments(t *testing.T) {

	keyset := randomKeyset()
	var metric string

	createCollection(t, keyset)

	tags := make(map[string]string)
	tags["host"] = fmt.Sprintf("host%d", newRand.Intn(511213432))
	tags["ttl"] = "1"
	tags["name"] = fmt.Sprintf("test%d", newRand.Intn(20000618621))
	tags["system"] = fmt.Sprintf("test_documents%d", newRand.Intn(5982736))

	docSize := 500

	for i := 0; i < docSize; i++ {

		metric = randomMetric()

		err := createChildDocuments(keyset, metric, tags)
		if !assert.NoError(t, err) {
			return
		}

	}

	tags2 := make(map[string]string)
	tags2["server"] = fmt.Sprintf("host%d", newRand.Intn(511213432))
	tags2["ttl"] = "30"
	tags2["service"] = fmt.Sprintf("solr%d", newRand.Intn(20000618621))

	for i := 0; i < docSize; i++ {

		metric = randomMetric()

		err := createChildDocuments(keyset, metric, tags2)
		if !assert.NoError(t, err) {
			return
		}

	}

	res := httpSearchChildDocumentsSolr(keyset, metric, "server", docSize)

	libRes := searchDocuments(keyset, metric, "server", docSize)

	if !testChildDocuments(t, res, libRes.Docs.([]solr.DocumentRaw)) {
		return
	}

}

func createChildDocuments(keyset, metric string, tags map[string]string) error {

	id := fmt.Sprintf("%d", newRand.Intn(50000))

	var childs []ChildDocuments
	i := 0
	for key, value := range tags {

		childs = append(childs, ChildDocuments{
			ID:       fmt.Sprintf("%s-t%d", id, i),
			TagKey:   key,
			TagValue: value,
		})

		i++
	}

	newChildDocument := NewChildDocuments{
		ID:             id,
		Metric:         metric,
		Type:           "meta",
		ParentDoc:      true,
		ChildDocuments: childs,
	}

	var send []interface{}
	send = append(send, newChildDocument)

	httpPostDocs(keyset, send)

	return nil

}

func testChildDocuments(t *testing.T, expected interface{}, actual interface{}) bool {

	if !assert.NotNil(t, expected, "expected value cannot be null") {
		return false
	}

	if !assert.NotNil(t, actual, "actual value cannot be null") {
		return false
	}

	expectedInterface, ok := expected.(ResponseSearchChildDocs)
	if !ok && !assert.True(t, ok, "expected interface must be a ResponseSearchChildDocs") {
		return false
	}

	actualInterface, ok := actual.([]solr.DocumentRaw)
	if !ok && !assert.True(t, ok, "actual interface must be a SimpleResponse") {
		return false
	}

	if !assert.Len(t, actualInterface, len(expectedInterface.Response.Docs), "expected %d documents", len(expectedInterface.Response.Docs)) {
		return false
	}

	result := true

	for i := 0; i < len(expectedInterface.Response.Docs); i++ {

		result = result && assert.Equal(t, expectedInterface.Response.Docs[i].Metric, actualInterface[i]["metric"])
		result = result && assert.Equal(t, expectedInterface.Response.Docs[i].Type, actualInterface[i]["type"])

		expectedChilds := expectedInterface.Response.Docs[i].ChildDocuments
		actualChilds := actualInterface[i]["_childDocuments_"].([]interface{})
		for j := 0; j < len(expectedChilds); j++ {
			result = result && assert.Equal(t, expectedChilds[j].TagKey, actualChilds[j].(map[string]interface{})["tag_key"])
			result = result && assert.Equal(t, expectedChilds[j].TagValue, actualChilds[j].(map[string]interface{})["tag_value"])
		}

		if !result {
			return false
		}
	}

	return result
}
