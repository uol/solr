package solr

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	gotesthttp "github.com/uol/gotest/http"
	"github.com/uol/solr"
)

const (
	solrQueryFacets                  string = "/solr/solr_test/select?&facet=true&facet.field=metric&rows=10"
	solrQueryFacetsBlockJoinFaceting string = "/solr/solr_test/bjqfacet?&facet=true&facet.field=metric&rows=10"
)

func createResponseData(uri, method string) gotesthttp.ResponseData {

	return gotesthttp.ResponseData{
		RequestData: gotesthttp.RequestData{
			URI:    "/",
			Method: method,
		},
		Status: http.StatusOK,
		Wait:   1 * time.Second,
	}

}

var defaultConf gotesthttp.Configuration = gotesthttp.Configuration{
	Host:        "localhost",
	Port:        18080,
	ChannelSize: 100,
}

func createSolrBackend(uri string) *gotesthttp.Server {

	response := createResponseData(uri, "GET")
	response.URI = uri
	response.Method = "GET"

	defaultConf.Responses = map[string][]gotesthttp.ResponseData{
		"default": {response},
	}

	return gotesthttp.NewServer(&defaultConf)
}

func TestURLEscapeQ(t *testing.T) {

	keyset := randomKeyset()
	solrQueryQ := "/solr/" + keyset + "/select?&q=%2A%3A%2A&rows=10"

	server := createSolrBackend(solrQueryQ)

	params := &solr.CloudParams{CollectionConfigName: "mycenae"}

	inst, err := solr.NewCloud("http://localhost:18080", time.Duration(10*time.Second), time.Duration(10*time.Second), 10000, 10000, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
	if !assert.NoError(t, err) {
		return
	}

	searchParams := &solr.SearchParams{
		Q:    "*:*",
		Rows: 10,
	}

	inst.Search(searchParams, keyset)

	gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    solrQueryQ,
		Method: "GET",
		Host:   defaultConf.Host,
		Port:   defaultConf.Port,
	})

	requestData := gotesthttp.WaitForServerRequest(server, time.Duration(20*time.Second), time.Duration(20*time.Second))

	if !assert.Equal(t, solrQueryQ, requestData.URI) {
		return
	}

	server.Close()

}

func TestURLEscapeFL(t *testing.T) {

	keyset := randomKeyset()

	solrQueryFL := "/solr/" + keyset + "/select?&fl=%2A%2C%5Bchild+parentFilter%3Dparent_doc%3Atrue+limit%3D10000%5D&rows=10"
	s := createSolrBackend(solrQueryFL)

	params := &solr.CloudParams{CollectionConfigName: "mycenae"}

	inst, err := solr.NewCloud("http://localhost:18080", time.Duration(20*time.Second), time.Duration(20*time.Second), 10000, 10000, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
	if !assert.NoError(t, err) {
		return
	}

	searchParams := &solr.SearchParams{
		FL:   "*,[child parentFilter=parent_doc:true limit=10000]",
		Rows: 10,
	}

	inst.Search(searchParams, keyset)

	gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    solrQueryFL,
		Method: "GET",
		Host:   defaultConf.Host,
		Port:   defaultConf.Port,
	})

	requestData := gotesthttp.WaitForServerRequest(s, time.Duration(20*time.Second), time.Duration(20*time.Second))

	if !assert.Equal(t, solrQueryFL, requestData.URI) {
		return
	}

	s.Close()

}

func TestURLEscapeFacets(t *testing.T) {

	s := createSolrBackend(solrQueryFacets)

	keyset := randomKeyset()

	params := &solr.CloudParams{CollectionConfigName: "mycenae"}
	inst, err := solr.NewCloud("http://localhost:18080", time.Duration(20*time.Second), time.Duration(20*time.Second), 100, 100, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
	if !assert.NoError(t, err) {
		return
	}

	fc := make(map[string]string)
	fc["facet.field"] = "metric"
	searchParams := &solr.SearchParams{
		Facets: fc,
		Rows:   10,
	}

	inst.Search(searchParams, keyset)

	gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    solrQueryFacets,
		Method: "GET",
		Host:   defaultConf.Host,
		Port:   defaultConf.Port,
	})

	requestData := gotesthttp.WaitForServerRequest(s, time.Duration(20*time.Second), time.Duration(20*time.Second))

	if !assert.Equal(t, solrQueryFacets, requestData.URI) {
		return
	}

	s.Close()

}

func TestURLEscapeFacetsBlockJoinFaceting(t *testing.T) {

	s := createSolrBackend(solrQueryFacetsBlockJoinFaceting)

	keyset := randomKeyset()

	params := &solr.CloudParams{CollectionConfigName: "mycenae"}
	inst, err := solr.NewCloud("http://localhost:18080", time.Duration(20*time.Second), time.Duration(20*time.Second), 10000, 10000, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
	if !assert.NoError(t, err) {
		return
	}

	fc := make(map[string]string)
	fc["facet.field"] = "metric"
	searchParams := &solr.SearchParams{
		BlockJoinFaceting: true,
		Facets:            fc,
		Rows:              10,
	}

	inst.Search(searchParams, keyset)

	gotesthttp.DoRequest(defaultConf.Host, defaultConf.Port, &gotesthttp.RequestData{
		URI:    solrQueryFacetsBlockJoinFaceting,
		Method: "GET",
		Host:   defaultConf.Host,
		Port:   defaultConf.Port,
	})

	requestData := gotesthttp.WaitForServerRequest(s, time.Duration(20*time.Second), time.Duration(20*time.Second))

	if !assert.Equal(t, solrQueryFacetsBlockJoinFaceting, requestData.URI) {
		return
	}

	s.Close()

}
