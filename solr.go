package solr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/buger/jsonparser"
	"github.com/uol/funks"
)

//NewCore Create a new instance core Solr
func NewCore(coreURL string, httpGeTimeout, httpPosTimeout time.Duration, httpGetmaxConn, httpPotmaxConn int, coreConfig *SettingsSolrCore, documentParser DocumentParser, documentWriter DocumentWriter) *Instance {

	if documentParser == nil {
		documentParser = &DefaultDocumentParser{}
	}
	if documentWriter == nil {
		documentWriter = &DefaultDocumentWriter{}
	}

	listURL := strings.Builder{}
	listURL.Grow(len(coreURL) + len(actionStatus))
	listURL.WriteString(coreURL)
	listURL.WriteString(actionStatus)

	clientGet := funks.CreateHTTPClientAdv(httpGeTimeout, true, httpGetmaxConn)
	clientPost := funks.CreateHTTPClientAdv(httpPosTimeout, true, httpPotmaxConn)
	return &Instance{
		coreURL:        coreURL,
		isCloud:        false,
		httpGetClient:  clientGet,
		httpPostClient: clientPost,
		coreConfig: &SettingsSolrCore{
			CoreName:    coreConfig.CoreName,
			InstanceDir: coreConfig.InstanceDir,
			Config:      coreConfig.Config,
			Schema:      coreConfig.Schema,
			DataDir:     coreConfig.DataDir,
		},
		documentParser:    documentParser,
		documentWriter:    documentWriter,
		listCollectionURL: listURL.String(),
	}
}

//NewCloud - create new cloud instance
func NewCloud(coreURL string, httpGeTimeout, httpPosTimeout time.Duration, httpGetmaxConn, httpPotmaxConn int, cloudParams *CloudParams, documentParser DocumentParser, documentWriter DocumentWriter) (*Instance, error) {

	if documentParser == nil {
		documentParser = &DefaultDocumentParser{}
	}

	if documentWriter == nil {
		documentWriter = &DefaultDocumentWriter{}
	}

	if cloudParams == nil {
		return nil, fmt.Errorf("cloudParams cannot be null")
	}

	if cloudParams.CollectionConfigName == "" {
		return nil, fmt.Errorf("CollectionConfigName not defined")
	}

	if cloudParams.NumShards == 0 {
		cloudParams.NumShards = 1
	}

	if cloudParams.MaxShardsPerNode == 0 {
		cloudParams.MaxShardsPerNode = 1
	}

	if cloudParams.ReplicationFactor == 0 {
		cloudParams.ReplicationFactor = 1
	}

	listURL := strings.Builder{}

	listURL.Grow(len(coreURL) + len(listCollection))
	listURL.WriteString(coreURL)
	listURL.WriteString(listCollection)

	clientGet := funks.CreateHTTPClientAdv(httpGeTimeout, true, httpGetmaxConn)
	clientPost := funks.CreateHTTPClientAdv(httpPosTimeout, true, httpPotmaxConn)

	return &Instance{
		coreURL:           coreURL,
		httpGetClient:     clientGet,
		httpPostClient:    clientPost,
		isCloud:           true,
		cloudParams:       cloudParams,
		documentParser:    documentParser,
		documentWriter:    documentWriter,
		listCollectionURL: listURL.String(),
	}, nil
}

// List - List cores/collections
func (s *Instance) List() ([]string, error) {

	raw, err := s.httpGet(s.listCollectionURL)
	if err != nil {
		return nil, err
	}

	cores, err := s.listParser(raw)
	if err != nil {
		return nil, err
	}

	return cores, nil

}

func (s *Instance) listParser(raw []byte) ([]string, error) {

	var list []string
	var parseError error

	if s.coreConfig != nil {
		err := jsonparser.ObjectEach(raw, func(key, value []byte, dataType jsonparser.ValueType, offset int) error {
			name, _, _, err := jsonparser.Get(value, "name")
			if err != nil {
				return err
			}
			list = append(list, string(name))
			return nil
		}, rawParserStatus)
		if err != nil {
			return nil, err
		}
	} else {
		_, err := jsonparser.ArrayEach(raw, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				parseError = err
				return
			}
			list = append(list, string(value))
		}, rawParserCollections)
		if err != nil {
			return nil, err
		}
		if parseError != nil {
			return nil, parseError
		}
	}

	return list, nil

}

// Delete - delete a core/collection
func (s *Instance) Delete(instanceName string) error {

	deleteURL := strings.Builder{}
	if s.isCloud {
		deleteURL.Grow(len(s.coreURL) + len(deleteCollection) + len(instanceName))
		deleteURL.WriteString(s.coreURL)
		deleteURL.WriteString(deleteCollection)
		deleteURL.WriteString(instanceName)
	} else {
		deleteURL.Grow(len(s.coreURL) + len(unloadCore) + len(instanceName) + len(deleteInstanceTrue))
		deleteURL.WriteString(s.coreURL)
		deleteURL.WriteString(unloadCore)
		deleteURL.WriteString(instanceName)
		deleteURL.WriteString(deleteInstanceTrue)
	}

	res, err := s.httpGetClient.Get(deleteURL.String())
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	var response *DeleteResponse
	err = json.Unmarshal([]byte(body), &response)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK || response.ResponseHeader.Status != 0 {
		return fmt.Errorf("error deleting collection, solr status is %v", response.ResponseHeader.Status)
	}

	return nil

}

// Create - create cores/collections
func (s *Instance) Create(instanceName string) error {

	newInstanceURL := strings.Builder{}

	if s.isCloud {

		lenMaxShardsPerNode := strconv.Itoa(s.cloudParams.MaxShardsPerNode)
		lenNumShards := strconv.Itoa(s.cloudParams.NumShards)
		lenReplicationFactor := strconv.Itoa(s.cloudParams.ReplicationFactor)

		newInstanceURL.Grow(len(s.coreURL) + len(actionCreateCollection) + len(instanceName) +
			len(collectionConfigName) + len(s.cloudParams.CollectionConfigName) + len(maxShardsPerNode) +
			len(lenMaxShardsPerNode) + len(numShards) + len(lenNumShards) + len(replicationFactor) +
			len(lenReplicationFactor) + getLen(s.cloudParams.AdvancedOptions, 2) + len(wtJSON))

		newInstanceURL.WriteString(s.coreURL)

		newInstanceURL.WriteString(actionCreateCollection)
		newInstanceURL.WriteString(instanceName)

		newInstanceURL.WriteString(collectionConfigName)
		newInstanceURL.WriteString(s.cloudParams.CollectionConfigName)

		newInstanceURL.WriteString(maxShardsPerNode)
		newInstanceURL.WriteString(lenMaxShardsPerNode)

		newInstanceURL.WriteString(numShards)
		newInstanceURL.WriteString(lenNumShards)

		newInstanceURL.WriteString(replicationFactor)
		newInstanceURL.WriteString(lenReplicationFactor)

		if len(s.cloudParams.AdvancedOptions) > 0 {

			for key, value := range s.cloudParams.AdvancedOptions {

				newInstanceURL.WriteString("&")
				newInstanceURL.WriteString(url.QueryEscape(key))
				newInstanceURL.WriteString("=")
				newInstanceURL.WriteString(url.QueryEscape(value))

			}

		}

		newInstanceURL.WriteString(wtJSON)

	} else {

		newInstanceURL.Grow(len(s.coreURL) + len(actionCreateCore) + len(s.coreConfig.CoreName) + len(instanceDir) + len(s.coreConfig.InstanceDir) + len(config) + len(schema) + len(s.coreConfig.Config) + len(s.coreConfig.Schema) + len(dataDir) + len(s.coreConfig.DataDir))
		newInstanceURL.WriteString(s.coreURL)
		newInstanceURL.WriteString(actionCreateCore)
		newInstanceURL.WriteString(s.coreConfig.CoreName)
		newInstanceURL.WriteString(instanceDir)
		newInstanceURL.WriteString(s.coreConfig.InstanceDir)
		newInstanceURL.WriteString(config)
		newInstanceURL.WriteString(s.coreConfig.Config)
		newInstanceURL.WriteString(schema)
		newInstanceURL.WriteString(s.coreConfig.Schema)
		newInstanceURL.WriteString(dataDir)
		newInstanceURL.WriteString(s.coreConfig.DataDir)

	}

	_, err := s.httpGet(newInstanceURL.String())
	if err != nil {
		return err
	}

	return nil

}

//Search A basic search in solr
func (s *Instance) Search(params *SearchParams, instanceName string) (*Response, error) {

	var res *Response

	if params != nil {
		var facet bool
		var searchType, stringParams string

		if params.toQueryString() != "" {
			stringParams = params.toQueryString()
		}
		if params.BlockJoinFaceting {
			searchType = stringFacet
		} else {
			searchType = stringSelect
		}

		if len(params.Facets) > 0 {
			facet = true
		}

		url := strings.Builder{}

		url.Grow(len(s.coreURL) + len(stringBar)*3 + len(searchType) + len(stringSolrBase) + len(instanceName) + len(stringParams))

		url.WriteString(s.coreURL)
		url.WriteString(stringBar)
		url.WriteString(stringSolrBase)
		url.WriteString(stringBar)
		url.WriteString(instanceName)
		url.WriteString(stringBar)
		url.WriteString(searchType)
		url.WriteString(stringParams)

		raw, err := s.httpGet(url.String())
		if err != nil {
			return nil, err
		}

		res, err = s.Decode(raw, facet)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, fmt.Errorf("params cannot be null")
	}

	return res, nil
}

func getLen(x interface{}, count int) int {

	xType := reflect.TypeOf(x)
	var lenInterface int
	switch xType.Kind() {
	case reflect.Slice:

		if reflect.ValueOf(x).Len() == 0 {
			return 0
		}
		for i := 0; i < reflect.ValueOf(x).Len(); i++ {
			lenInterface += len(reflect.ValueOf(x).Index(i).String()) + count
		}

	case reflect.Map:

		if len(reflect.ValueOf(x).MapKeys()) == 0 {
			return 0
		}

		for _, e := range reflect.ValueOf(x).MapKeys() {
			v := reflect.ValueOf(x).MapIndex(e)
			lenInterface += len(e.String()) + len(v.String()) + count
		}

	default:
		return 0

	}

	return lenInterface

}

// UpdateDocument - post json on solr, if the postParams is passed it will be adding in the request. For deleting items you can use the post function using the json format: https://lucene.apache.org/solr/guide/6_6/uploading-data-with-index-handlers.html#UploadingDatawithIndexHandlers-SendingJSONUpdateCommands
func (s *Instance) UpdateDocument(instanceName string, postParams map[string]string, payload interface{}) error {

	pp := strings.Builder{}

	pp.Grow(len(s.coreURL) + getLen(postParams, len(stringAmpersand+stringEqual)) + len(stringBar)*2 + len(stringSolrBase) + len(instanceName) + len(stringUpdate) + len(stringCommitTrue) + len(wtJSON))

	pp.WriteString(s.coreURL)
	pp.WriteString(stringBar)
	pp.WriteString(stringSolrBase)
	pp.WriteString(stringBar)
	pp.WriteString(instanceName)
	pp.WriteString(stringUpdate)

	if len(postParams) > 0 {
		for k, v := range postParams {
			pp.WriteString(stringAmpersand)
			pp.WriteString(url.QueryEscape(k))
			pp.WriteString(stringEqual)
			pp.WriteString(url.QueryEscape(v))
		}
	}

	pp.WriteString(stringCommitTrue)
	pp.WriteString(wtJSON)

	writer, err := s.documentWriter.Writer(payload)
	if err != nil {
		return err
	}

	resp, err := s.httpPost(pp.String(), contentType, string(writer))
	if err != nil {
		return err
	}

	var response responseRaw

	err = json.Unmarshal([]byte(resp), &response)
	if err != nil {
		return err
	}

	if response.Header.Status != 0 {
		return fmt.Errorf("Solr returned status %d", response.Header.Status)
	}

	return nil
}

func (s *Instance) httpPost(url, contentType, body string) (string, error) {

	payload := bytes.NewBufferString(body)
	res, err := http.Post(url, contentType, payload)
	if err != nil {
		return "", err
	}

	defer res.Body.Close()
	resSTR, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(resSTR), nil

}

func (s *Instance) httpGet(url string) ([]byte, error) {

	client := s.httpGetClient
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {

		msg := fmt.Sprintf("HTTP Status: %s. ", res.Status)
		if len(body) > 0 {
			msg += fmt.Sprintf("Body: %s", body)
		}

		return nil, fmt.Errorf("HTTP Status: %s", msg)
	}

	return body, nil

}
