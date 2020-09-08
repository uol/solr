package solr

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestSearchByDate(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()

	httpCreateCollection(keyset)

	tags := make(map[string]string)
	tags["host"] = "host"

	now := time.Now()
	fiveHours := now.Add(-5 * time.Hour)
	expectedFiveHours, err := makeDocs(metric, keyset, fiveHours.Format("2006-01-02T15:04:05Z"), tags, 30, false)
	if !assert.NoError(t, err) {
		return
	}

	tags2 := make(map[string]string)
	tags2["service"] = "solr"

	oneMinute := now.Add(-1 * time.Minute)
	expectedOneMinute, err := makeDocs(metric, keyset, oneMinute.Format("2006-01-02T15:04:05Z"), tags2, 10, false)
	if !assert.NoError(t, err) {
		return
	}

	res := searchByDate(t, fiveHours, fiveHours, keyset, metric, 60, true)

	if !testDocumentRaw(t, expectedFiveHours, res.Docs.([]solr.DocumentRaw)) {
		return
	}

	res2 := searchByDate(t, oneMinute, oneMinute, keyset, metric, 60, true)

	if !testDocumentRaw(t, expectedOneMinute, res2.Docs.([]solr.DocumentRaw)) {
		return
	}

	resOneMinuteExclude := searchByDate(t, fiveHours.Add(-time.Second), fiveHours.Add(time.Second), keyset, metric, 60, false)

	if !testDocumentRaw(t, expectedFiveHours, resOneMinuteExclude.Docs.([]solr.DocumentRaw)) {
		return
	}

	resFiveHoursExclude := searchByDate(t, oneMinute.Add(-time.Second), oneMinute.Add(time.Second), keyset, metric, 60, false)

	if !testDocumentRaw(t, expectedOneMinute, resFiveHoursExclude.Docs.([]solr.DocumentRaw)) {
		return
	}

	resExcludeAll := searchByDate(t, fiveHours.Add(time.Second), oneMinute, keyset, metric, 60, false)

	if !assert.Equal(t, int(0), len(resExcludeAll.Docs.([]solr.DocumentRaw))) {
		return
	}

}

func searchByDate(t *testing.T, start, last time.Time, keyset, metric string, rows int, isInclude bool) *solr.Response {

	var fqs []string
	if isInclude {
		fqs = append(fqs, fmt.Sprintf("creation_date:[%v TO %v]", start.Format("2006-01-02T15:04:05Z"), last.Format("2006-01-02T15:04:05Z")))
	} else {
		fqs = append(fqs, fmt.Sprintf("creation_date:{%v TO %v}", start.Format("2006-01-02T15:04:05Z"), last.Format("2006-01-02T15:04:05Z")))
	}

	fqs = append(fqs, fmt.Sprintf("metric:%s", metric))
	searchParams := &solr.SearchParams{
		Q:             "*:*",
		FilterQueries: fqs,
		Rows:          rows,
	}

	resp, err := defaultInstance.Search(searchParams, keyset)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	return resp

}
