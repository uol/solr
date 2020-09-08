package solr

import (
	"testing"
)

func TestCreateCollection(t *testing.T) {

	keyset := randomKeyset()

	createCollection(t, keyset)

	checkCollections(t, keyset, true)

}

func createCollection(t *testing.T, ksid string) {

	err := defaultInstance.Create(ksid)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

}
