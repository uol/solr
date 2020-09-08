# Solr Lib
A simple solr client lib.


## Example:
```
inst, err := .NewCloud(getSolrAddress(), time.Duration(20*time.Second), time.Duration(20*time.Second), 100, 100, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
if err != nil {
		panic(err)
	}
	
err := inst.Create("CollectionName")
if err != nil {
		panic(err)
	}

params := make(map[string]string)

payload := `{
	"id": 1,
	"metric": "solr_metric"
}`
err = inst.UpdateDocument("CollectionName", params, payload)
if err != nil {
	panic(err)
}
	
searchParams := &solr.SearchParams{
    Q:      "*:*",
	Rows:   10,
}
inst.Search(searchParams, "CollectionName")

```
