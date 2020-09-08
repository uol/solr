package solr

import (
	"fmt"

	"github.com/buger/jsonparser"
)

//Decode - decode a raw byte from solr instance and return a formated response
func (s *Instance) Decode(raw []byte, facet bool) (*Response, error) {

	res := &Response{}
	var err error

	if res.NumFound, res.Status, res.QTime, err = s.parserNumbers(raw); err != nil {
		return nil, fmt.Errorf("error parsing numbers: %v", err.Error())
	}

	if res.Docs, err = s.documentParser.Parse(raw); err != nil {
		return nil, fmt.Errorf("error parsing docs: %v", err.Error())
	}

	if facet {
		var errors []error
		if res.Facets, errors = s.parserFacets(raw); err != nil {
			return nil, fmt.Errorf("error parsing facets: %v", errors)
		}
	}

	return res, err

}

func (s *Instance) parserFacets(raw []byte) ([]FacetField, []error) {

	facetValueArray := []FacetValue{}
	facetValues := FacetValue{}
	facetFields := FacetField{}
	var facetFieldsArray []FacetField
	var err error
	var k string
	var v int64
	var errors []error

	err = jsonparser.ObjectEach(raw, func(key, value []byte, dataType jsonparser.ValueType, offset int) error {

		_, err = jsonparser.ArrayEach(value, func(tvalue []byte, dataType jsonparser.ValueType, offset int, err error) {

			if err != nil {
				errors = append(errors, err)
			}

			switch dataType {
			case jsonparser.String:
				k = string(tvalue)
			case jsonparser.Number:
				v, err = jsonparser.GetInt(tvalue)
				if err != nil {
					errors = append(errors, err)
				}
				facetValues.Name = k
				facetValues.Value = v
				facetValueArray = append(facetValueArray, facetValues)
			}

		})
		if err != nil {
			return err
		}

		facetFields.Name = string(key)
		facetFields.List = facetValueArray
		facetFieldsArray = append(facetFieldsArray, facetFields)
		return nil
	}, rawFacetsCount, rawFacetFields)
	if err != nil {
		return nil, errors
	}

	return facetFieldsArray, nil

}

func (s *Instance) parserNumbers(raw []byte) (found, status, qtime int64, err error) {

	if found, err = jsonparser.GetInt(raw, rawResponse, rawNumFound); err != nil {
		return found, status, qtime, err
	}
	if status, err = jsonparser.GetInt(raw, rawResponseHeader, rawStatus); err != nil {
		return found, status, qtime, err
	}
	if qtime, err = jsonparser.GetInt(raw, rawResponseHeader, rawQtime); err != nil {
		return found, status, qtime, err
	}

	return found, status, qtime, err

}
