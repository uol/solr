package solr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

const (
	stringHTTP     string = "http://"
	stringSolrPort string = ":8983"
)

var (
	newSource       = rand.NewSource(time.Now().UnixNano())
	newRand         = rand.New(newSource)
	defaultInstance = createInstance(&solr.CloudParams{
		CollectionConfigName: "mycenae",
	})
)

//SimpleResponse - simple solr response
type SimpleResponse struct {
	ResponseHeader struct {
		ZkConnected bool `json:"zkConnected"`
		Status      int  `json:"status"`
		QTime       int  `json:"QTime"`
		Params      struct {
			Q string `json:"q"`
		} `json:"params"`
	} `json:"responseHeader"`
	Response struct {
		NumFound int `json:"numFound"`
		Start    int `json:"start"`
		Docs     []struct {
			ID           string    `json:"id"`
			Metric       string    `json:"metric"`
			Type         string    `json:"type"`
			TagKey       string    `json:"tag_key"`
			TagValue     string    `json:"tag_value"`
			CreationDate time.Time `json:"creation_date"`
			Version      int64     `json:"_version_"`
		} `json:"docs"`
	} `json:"response"`
}

// DefaultDocument - default document to post
type DefaultDocument struct {
	ID           string `json:"id"`
	Metric       string `json:"metric"`
	Type         string `json:"type"`
	ParentDoc    bool   `json:"parent_doc"`
	TagKey       string `json:"tag_key"`
	TagValue     string `json:"tag_value"`
	CreationDate string `json:"creation_date"`
}

// ResponseSearchChildDocs - solr with child documents response
type ResponseSearchChildDocs struct {
	ResponseHeader struct {
		ZkConnected bool `json:"zkConnected"`
		Status      int  `json:"status"`
		QTime       int  `json:"QTime"`
		Params      struct {
			Q  string `json:"q"`
			Fl string `json:"fl"`
		} `json:"params"`
	} `json:"responseHeader"`
	Response struct {
		NumFound int `json:"numFound"`
		Start    int `json:"start"`
		Docs     []struct {
			ID             string    `json:"id"`
			Metric         string    `json:"metric"`
			Type           string    `json:"type"`
			CreationDate   time.Time `json:"creation_date"`
			Version        int64     `json:"_version_"`
			ChildDocuments []struct {
				ID           string    `json:"id"`
				TagKey       string    `json:"tag_key"`
				TagValue     string    `json:"tag_value"`
				CreationDate time.Time `json:"creation_date"`
				Version      int64     `json:"_version_"`
			} `json:"_childDocuments_"`
		} `json:"docs"`
	} `json:"response"`
}

// NewChildDocuments - structs for ChildDocuments tests
type NewChildDocuments struct {
	ID             string           `json:"id"`
	Metric         string           `json:"metric"`
	Type           string           `json:"type"`
	ParentDoc      bool             `json:"parent_doc"`
	ChildDocuments []ChildDocuments `json:"_childDocuments_"`
}

//ChildDocuments - childDocuments
type ChildDocuments struct {
	ID       string `json:"id"`
	TagKey   string `json:"tag_key"`
	TagValue string `json:"tag_value"`
}

//Collections - list collections
type Collections struct {
	ResponseHeader struct {
		Status int `json:"status"`
		QTime  int `json:"QTime"`
	} `json:"responseHeader"`
	Collections []string `json:"collections"`
}

func getSolrAddress() string {
	out, err := exec.Command("docker", "inspect", "--format", "{{ .NetworkSettings.Networks.timeseriesNetwork.IPAddress }}", "solr1").Output()
	if err != nil {
		panic(err)
	}

	solrAddress := strings.Trim(string(out), "\n")

	baseURL := strings.Builder{}
	baseURL.Grow(len(stringHTTP) + len(solrAddress) + len(stringSolrPort))
	baseURL.WriteString("http://")
	baseURL.WriteString(solrAddress)
	baseURL.WriteString(":8983")
	return baseURL.String()
}

func createInstance(params *solr.CloudParams) *solr.Instance {

	inst, err := solr.NewCloud(getSolrAddress(), time.Duration(20*time.Second), time.Duration(20*time.Second), 100, 100, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
	if err != nil {
		panic(err)
	}
	return inst
}

func randomKeyset() string {

	return "keyset_test_solr" + fmt.Sprint(newRand.Intn(20000))
}

func randomMetric() string {

	return "metric_test_solr" + fmt.Sprint(newRand.Intn(20000))
}

func checkCollections(t *testing.T, ksid string, exists bool) {

	url := getSolrAddress() + "/solr/admin/collections?action=LIST&indexInfo=false&wt=json"

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != http.StatusOK {
		panic(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var collections Collections
	err = json.Unmarshal(body, &collections)
	if err != nil {
		panic(err)
	}
	var isExists bool
	for i := 0; i < len(collections.Collections); i++ {

		if collections.Collections[i] == ksid {
			isExists = true
			break
		}

	}

	if exists {

		if isExists {
			return
		}
		t.Logf("collection %s not exists", ksid)
		t.Fail()

	} else {
		if isExists {
			t.Logf("collection %s exists", ksid)
			t.Fail()
		}
	}

}

func solrLibPost(ksid string, json interface{}) {

	params := make(map[string]string)

	err := defaultInstance.UpdateDocument(ksid, params, json)
	if err != nil {
		panic(err)
	}

}

func httpSearchSolr(ksid, metric string, rows int) SimpleResponse {

	url := getSolrAddress() + "/solr/" + ksid + "/select?q=%7B%21parent+which%3D%22%28parent_doc%3Atrue+AND+type%3Ameta+AND+metric%3A" + metric + "%29%22%7D&fl=*%2C%5Bchild+parentFilter%3Dparent_doc%3Atrue+limit%3D10000%5D&rows=" + strconv.Itoa(rows)

	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var response SimpleResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}
	return response

}

func httpSearchChildDocumentsSolr(keyset, metric, tagKey string, docSize int) ResponseSearchChildDocs {
	url := getSolrAddress() + "/solr/" + keyset + "/select?fl=*,[child%20parentFilter=parent_doc:true]&fq=metric:" + metric + "&fq={!parent%20which=\"parent_doc:true\"}tag_key:" + tagKey + "&q={!parent%20which=%22parent_doc:true%22}&rows=" + strconv.Itoa(docSize)

	method := "GET"
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	defer res.Body.Close()
	var response ResponseSearchChildDocs
	err = json.Unmarshal(body, &response)
	if err != nil {
		panic(err)
	}
	return response
}

func makeDocs(metric, keyset string, creation string, tags map[string]string, docSize int, isLib bool) ([]DefaultDocument, error) {

	var newDoc []DefaultDocument

	if docSize == 0 {
		docSize = len(tags)
	}

	if creation == "" {
		creation = time.Now().Format("2006-01-02T15:04:05Z")
	}

	for i := 0; i < docSize; i++ {

		for k, v := range tags {

			newDoc = append(newDoc, DefaultDocument{
				ID:           fmt.Sprintf("%d", newRand.Intn(50000)),
				Metric:       metric,
				Type:         "meta",
				ParentDoc:    true,
				TagKey:       k,
				TagValue:     fmt.Sprintf("%s%d", v, i+1),
				CreationDate: creation,
			})

		}

	}

	if isLib {

		solrLibPost(keyset, newDoc)

	} else {

		httpPostDocs(keyset, newDoc)

	}

	return newDoc, nil

}

func makeDocsByArray(metric, keyset, creation string, tagName string, tags []string, isLib bool) []DefaultDocument {

	var newDoc []DefaultDocument

	for i := 0; i < len(tags); i++ {

		if creation == "" {
			creation = time.Now().Format("2006-01-02T15:04:05Z")
		}

		newDoc = append(newDoc, DefaultDocument{
			ID:           fmt.Sprintf("%d", i),
			Metric:       metric,
			Type:         "meta",
			ParentDoc:    true,
			TagKey:       tagName,
			TagValue:     tags[i],
			CreationDate: creation,
		})

	}

	httpPostDocs(keyset, newDoc)

	return newDoc

}

func searchDocuments(keyset, metric, tagKey string, rows int) *solr.Response {
	q := "{!parent which=\"(parent_doc:true AND type:meta AND metric:" + fmt.Sprintf("%s", metric) + ")\"}"
	fl := "*,[child parentFilter=parent_doc:true limit=10000]"
	fq := fmt.Sprintf("{!parent which=\"parent_doc:true\"}tag_key:%s", tagKey)
	searchParams := &solr.SearchParams{
		Q:             q,
		FilterQueries: []string{fq},
		FL:            fl,
		Rows:          rows,
	}
	libres, err := defaultInstance.Search(searchParams, keyset)
	if err != nil {
		panic(err)
	}

	return libres
}

func httpCreateCollection(keyset string) {

	url := getSolrAddress() + "/solr/admin/collections?action=CREATE&name=" + keyset + "&collection.configName=mycenae&maxShardsPerNode=1&numShards=1&replicationFactor=1&wt=json"

	_, err := http.Get(url)
	if err != nil {
		panic(err)
	}

}

func httpPostDocs(keyset string, payload interface{}) {
	url := getSolrAddress() + "/solr/" + keyset + "/update?commit=true"

	byte, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	_, err = http.Post(url, "application/json", bytes.NewBuffer(byte))
	if err != nil {
		panic(err)
	}

}

func testHTTPDocument(t *testing.T, expected interface{}, actual interface{}) bool {

	if !assert.NotNil(t, expected, "expected value cannot be null") {
		return false
	}

	if !assert.NotNil(t, actual, "actual value cannot be null") {
		return false
	}

	expectedInterface, ok := expected.([]DefaultDocument)
	if !ok && !assert.True(t, ok, "expected interface must be a DefaultDocument array") {
		return false
	}

	actualInterface, ok := actual.(SimpleResponse)
	if !ok && !assert.True(t, ok, "actual interface must be a SimpleResponse") {
		return false
	}

	if !assert.Len(t, actualInterface.Response.Docs, len(expectedInterface), "expected %d documents", len(expectedInterface)) {
		return false
	}

	result := true

	for i := 0; i < len(expectedInterface); i++ {

		result = result && assert.Equal(t, expectedInterface[i].Metric, actualInterface.Response.Docs[i].Metric)
		result = result && assert.Equal(t, expectedInterface[i].Type, actualInterface.Response.Docs[i].Type)
		result = result && assert.Equal(t, expectedInterface[i].TagKey, actualInterface.Response.Docs[i].TagKey)
		result = result && assert.Equal(t, expectedInterface[i].TagValue, actualInterface.Response.Docs[i].TagValue)

		if !result {
			return false
		}
	}

	return result
}

func testDocumentRaw(t *testing.T, expected interface{}, actual interface{}) bool {

	if !assert.NotNil(t, expected, "expected value cannot be null") {
		return false
	}

	if !assert.NotNil(t, actual, "actual value cannot be null") {
		return false
	}

	expectedInterface, ok := expected.([]DefaultDocument)
	if !ok && !assert.True(t, ok, "expected interface must be a DefaultDocument array") {
		return false
	}

	actualInterface, ok := actual.([]solr.DocumentRaw)
	if !ok && !assert.True(t, ok, "actual interface must be a SimpleResponse") {
		return false
	}

	if !assert.Len(t, actualInterface, len(expectedInterface), "expected %d documents", len(expectedInterface)) {
		return false
	}

	result := true

	for i := 0; i < len(expectedInterface); i++ {

		result = result && assert.Equal(t, expectedInterface[i].CreationDate, actualInterface[i]["creation_date"])
		result = result && assert.Equal(t, expectedInterface[i].Metric, actualInterface[i]["metric"])
		result = result && assert.Equal(t, expectedInterface[i].Type, actualInterface[i]["type"])
		result = result && assert.Equal(t, expectedInterface[i].TagKey, actualInterface[i]["tag_key"])
		result = result && assert.Equal(t, expectedInterface[i].TagValue, actualInterface[i]["tag_value"])

		if !result {
			return false
		}
	}

	return result
}
