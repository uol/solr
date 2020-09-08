# Solr Lib
A simple solr client lib.


## Create a solr instance:
```

inst, err := NewCloud(getSolrAddress(), time.Duration(20*time.Second), time.Duration(20*time.Second), 100, 100, params, &solr.DefaultDocumentParser{}, &solr.DefaultDocumentWriter{})
if err != nil {
		panic(err)
	}
	
```

## Create a collection:
```
err := inst.Create("CollectionName")
if err != nil {
		panic(err)
	}

```
## Send documents to solr server:
```
params := make(map[string]string)
payload := `{
	"id": 1,
	"metric": "solr_metric"
}`
err = inst.UpdateDocument("CollectionName", params, payload)
if err != nil {
	panic(err)
}
```

## Search documents:
```
searchParams := &solr.SearchParams{
    Q:      "*:*",
	Rows:   10,
}
inst.Search(searchParams, "CollectionName")

```
