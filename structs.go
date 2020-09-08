package solr

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// The *Raw structs are used to unmarshall the JSON from Solr
// via Go's built-in functions. They are not exposed outside
// the solr package.
type headerRaw struct {
	Status int `json:"status"`
	QTime  int `json:"QTime"`
	// Use interface{} because some params are strings and
	// others (e.g. fq) are arrays of strings.
	Params map[string]interface{} `json:"params"`
}

type dataRaw struct {
	NumFound  int           `json:"numFound"`
	Start     int           `json:"start"`
	Documents []DocumentRaw `json:"docs"`
}

//DocumentRaw - Just as it comes from Solr
type DocumentRaw map[string]interface{}

type errorRaw struct {
	Trace string `json:"trace"`
	Code  int    `json:"code"`
}

type responseRaw struct {
	Header      headerRaw      `json:"responseHeader"`
	Data        dataRaw        `json:"response"`
	Error       errorRaw       `json:"error"`
	FacetCounts facetCountsRaw `json:"facet_counts"`
}

type facetCountsRaw struct {
	Queries interface{}              `json:"facet_queries"`
	Fields  map[string][]interface{} `json:"facet_fields"`
}

//Response - response data from solr instance
type Response struct {
	Status   int64        `json:"status,omitempty"`
	QTime    int64        `json:"Qtime,omitempty"`
	NumFound int64        `json:"numFound,omitempty"`
	Docs     interface{}  `json:"Docs,omitempty"`
	Facets   []FacetField `json:"Facets,omitempty"`
}

//FacetField - struct for facets
type FacetField struct {
	Name string
	List []FacetValue
}

//FacetValue - struct for facets name and value
type FacetValue struct {
	Name  string
	Value int64
}

// Instance The main class to instance solr
type Instance struct {
	coreURL           string            //CoreURL - ex: http://localhost:8983
	coreConfig        *SettingsSolrCore //CoreConfig - Configs required for admin solr core
	cloudParams       *CloudParams      //CloudParams - Params used for admin solr cloud
	isCloud           bool
	documentParser    DocumentParser
	documentWriter    DocumentWriter
	listCollectionURL string
	httpGetClient     *http.Client
	httpPostClient    *http.Client
}

// SearchParams - Params for solr queries
type SearchParams struct {
	Q                 string
	FL                string
	FilterQueries     []string
	BlockJoinFaceting bool
	Sort              string
	Facets            map[string]string
	Rows              int
	Start             int
}

func (params SearchParams) toQueryString() string {

	qs := strings.Builder{}

	qs.Grow(len(stringSort) + len(params.Sort) + len(stringQ) + len(params.Q) +
		len(stringFL) + len(params.FL) + getLen(params.FilterQueries, len(stringFQ)) +
		getLen(params.Facets, len(stringFacetTrue+stringEqual)) + len(stringStart) +
		len(strconv.Itoa(params.Start)) + len(stringRows) + len(strconv.Itoa(params.Rows)))

	if len(params.FilterQueries) > 0 {

		for i := 0; i < len(params.FilterQueries); i++ {

			qs.WriteString(stringFQ)
			qs.WriteString(url.QueryEscape(params.FilterQueries[i]))

		}

	}

	if params.Facets != nil {

		qs.WriteString(stringFacetTrue)

		for k, v := range params.Facets {

			qs.WriteString(url.QueryEscape(k))
			qs.WriteString(stringEqual)
			qs.WriteString(url.QueryEscape(v))

		}

	}

	if params.Sort != "" {

		qs.WriteString(stringSort)
		qs.WriteString(url.QueryEscape(params.Sort))

	}

	if params.Q != "" {

		qs.WriteString(stringQ)
		qs.WriteString(url.QueryEscape(params.Q))

	}

	if params.FL != "" {

		qs.WriteString(stringFL)
		qs.WriteString(url.QueryEscape(params.FL))

	}

	qs.WriteString(stringStart)
	qs.WriteString(strconv.Itoa(params.Start))

	qs.WriteString(stringRows)
	qs.WriteString(strconv.Itoa(params.Rows))

	return qs.String()

}

//DeleteResponse - response delete
type DeleteResponse struct {
	ResponseHeader struct {
		Status int `json:"status"`
		QTime  int `json:"QTime"`
	} `json:"responseHeader"`
}

// SettingsSolrCore - Configurations for create a new solr core for more information visit https://lucene.apache.org/solr/guide/6_6/coreadmin-api.html
type SettingsSolrCore struct {
	CoreName    string //CoreName - The name of the new core. Same as "name" on the <core> element.
	InstanceDir string //InstanceDir - The directory where files for this SolrCore should be stored. Same as instanceDir on the <core> element.
	Config      string //Config - Name of the config file (i.e., solrconfig.xml) relative to instanceDir
	Schema      string //Schema - Name of the schema file to use for the core.
	DataDir     string //DataDir - Name of the data directory relative to instanceDir
}

//CloudParams - parameters for creating collection
type CloudParams struct {
	CollectionConfigName string
	NumShards            int
	MaxShardsPerNode     int
	ReplicationFactor    int
	AdvancedOptions      map[string]string
}
