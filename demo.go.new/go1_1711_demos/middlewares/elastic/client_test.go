package elastic

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/olivere/elastic/v7"
)

const indexTag = "index_zz_test"

func ElasticInitForTest() *elastic.Client {
	hosts := strings.Split(os.Getenv("ES_HOSTS"), ",")
	userName := os.Getenv("ES_USERNAME")
	password := os.Getenv("ES_PASSOWRD")
	return NewElastic(hosts, userName, password)
}

func TestPingElastic(t *testing.T) {
	esClient := ElasticInitForTest()
	ver, err := esClient.ElasticsearchVersion("http://127.0.0.1:10200")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Elasticsearch version: %s\n", ver)

	isExist, err := esClient.IndexExists(indexTag).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Index exist:", isExist)
}

func TestElasticCreateIndex(t *testing.T) {
	mapping := `
{
	"settings": {
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings": {
		"properties": {
			"name": {
				"type": "keyword"
			},
			"identify": {
				"type": "long"
			},
			"message": {
				"type": "text"
			},
			"tags": {
				"type": "keyword"
			}
		}
	}
}
`

	esClient := ElasticInitForTest()
	createIndex, err := esClient.CreateIndex(indexTag).Body(mapping).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !createIndex.Acknowledged {
		t.Fatal(fmt.Errorf("Create index not acknowledged"))
	}

	res, err := esClient.ClusterHealth().Index(indexTag).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Cluster status is %q\n", res.Status)
}

type indexMapping struct {
	User     string    `json:"user"`
	Message  string    `json:"message"`
	Identify int       `json:"identify"`
	Tags     []string  `json:"tags,omitempty"`
	Created  time.Time `json:"created,omitempty"`
}

func TestElasticAddRecord(t *testing.T) {
	esClient := ElasticInitForTest()
	record1 := indexMapping{User: "foo", Message: "Take Five", Identify: 0}
	put1, err := esClient.Index().Index(indexTag).Id("1").BodyJson(record1).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

	record2 := `{"user": "bar", "message": "It's a Raggy Waltz"}`
	put2, err := esClient.Index().Index(indexTag).Id("2").BodyString(record2).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put2.Id, put2.Index, put2.Type)
}

func TestElasticGetByIndex(t *testing.T) {
	esClient := ElasticInitForTest()
	get, err := esClient.Get().Index(indexTag).Id("1").Do(context.Background())
	if err != nil {
		switch {
		case elastic.IsNotFound(err):
			t.Fatal(fmt.Errorf("Document not found: %v", err))
		case elastic.IsTimeout(err):
			t.Fatal(fmt.Errorf("Timeout retrieving document: %v", err))
		case elastic.IsConnErr(err):
			t.Fatal(fmt.Errorf("Connection problem: %v", err))
		default:
			t.Fatal(err)
		}
	}
	fmt.Printf("Got document %s in version %d from index %s, type %s\n", get.Id, get.Version, get.Index, get.Type)
}

func TestElasticSearch(t *testing.T) {
	esClient := ElasticInitForTest()
	_, err := esClient.Refresh().Index(indexTag).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Search with a term query
	termQuery := elastic.NewTermQuery("user", "foo")
	searchResult, err := esClient.Search().
		Index(indexTag).         // search in index "index_zz_test"
		Query(termQuery).        // specify the query
		Sort("user", true).      // sort by "user" field, ascending
		From(0).Size(10).        // take documents 0-9
		Pretty(true).            // pretty print request and response JSON
		Do(context.Background()) // execute
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	var data indexMapping
	for _, item := range searchResult.Each(reflect.TypeOf(data)) {
		d := item.(indexMapping)
		fmt.Printf("Data for %s: %s\n", d.User, d.Message)
	}

	if searchResult.TotalHits() > 0 {
		fmt.Printf("Found a total of %d data\n", searchResult.TotalHits())
		for _, hit := range searchResult.Hits.Hits {
			var d indexMapping
			if err := json.Unmarshal(hit.Source, &d); err != nil {
				t.Fatal(err)
			}
			fmt.Printf("Data for %s: %s\n", d.User, d.Message)
		}
	} else {
		fmt.Println("Found no data")
	}
}

func TestElasticDeleteIndex(t *testing.T) {
	esClient := ElasticInitForTest()
	deleteIndex, err := esClient.DeleteIndex(indexTag).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !deleteIndex.Acknowledged {
		t.Fatal(fmt.Errorf("Delete action not Acknowledged"))
	}
	fmt.Println("Delete index done")
}

// Rollover Index

func TestElasticCreateAlias(t *testing.T) {
	aliasName := ""
	esClient := ElasticInitForTest()
	addAction := elastic.NewAliasAddAction(aliasName).Index(indexTag).IsWriteIndex(true)
	createAlias, err := esClient.Alias().Action(addAction).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !createAlias.Acknowledged {
		t.Fatal(fmt.Errorf("Create alias not acknowledged"))
	}
}

func TestElasticCreateRollover(t *testing.T) {
	esClient := ElasticInitForTest()
	conditions := map[string]interface{}{
		"max_age":  "7d",
		"max_docs": int64(10000),
		"max_size": "5gb",
	}
	resp, err := esClient.RolloverIndex(indexTag).Conditions(conditions).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Acknowledged {
		t.Fatal(fmt.Errorf("Create rollover policy not acknowledged"))
	}
}
