package solr

import (
	"testing"

	"github.com/uol/solr"
)

func TestDeleteCollection(t *testing.T) {

	keyset := randomKeyset()
	keyset2 := randomKeyset()

	httpCreateCollection(keyset)

	httpCreateCollection(keyset2)

	checkCollections(t, keyset, true)

	checkCollections(t, keyset2, true)

	deleteCollection(t, defaultInstance, keyset)

	checkCollections(t, keyset, false)

	checkCollections(t, keyset2, true)

}

func deleteCollection(t *testing.T, inst *solr.Instance, instanceName string) {

	err := defaultInstance.Delete(instanceName)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

}
