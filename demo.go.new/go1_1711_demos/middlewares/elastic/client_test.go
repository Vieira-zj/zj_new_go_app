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

const (
	indexBaseName   = "index-zz-test-tab-000001"
	indexPattern    = "index-zz-test-tab-*"
	indexWriteAlias = "index-zz-test-tab-write"
)

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

	isExist, err := esClient.IndexExists(indexBaseName).Do(context.Background())
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
		"number_of_replicas": 0,
		"max_result_window": 100000,
		"refresh_interval": "30s"
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
	createIndex, err := esClient.CreateIndex(indexBaseName).Body(mapping).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !createIndex.Acknowledged {
		t.Fatal(fmt.Errorf("Create index not acknowledged"))
	}

	res, err := esClient.ClusterHealth().Index(indexBaseName).Do(context.Background())
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

func TestElasticAddDocs(t *testing.T) {
	// 	往 index 中添加 document 时，注意：
	// 1. 生成 random unique id
	// 2. json 对象数据本身不需要 id 字段
	esClient := ElasticInitForTest()
	record1 := indexMapping{User: "foo", Message: "Take Five", Identify: 0}
	put1, err := esClient.Index().Index(indexPattern).Id("1").BodyJson(record1).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

	record2 := `{"user": "bar", "message": "It's a Raggy Waltz"}`
	put2, err := esClient.Index().Index(indexPattern).Id("2").BodyString(record2).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Indexed data %s to index %s, type %s\n", put2.Id, put2.Index, put2.Type)
}

func TestElasticQueryByIndexID(t *testing.T) {
	esClient := ElasticInitForTest()
	get, err := esClient.Get().Index(indexPattern).Id("1").Do(context.Background())
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

func TestElasticSearchByTerm(t *testing.T) {
	esClient := ElasticInitForTest()
	_, err := esClient.Refresh().Index(indexPattern).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Term and Match query
	// https://www.elastic.co/guide/en/elasticsearch/reference/7.0/query-dsl-term-query.html
	//
	// By default, Elasticsearch changes the values of text fields as part of analysis.
	// This can make finding exact matches for text field values difficult. To search text field values, use the match query instead.
	//
	// matchQuery := elastic.NewMatchQuery("user", "bar-foo")
	//

	// search with a term query
	termQuery := elastic.NewTermQuery("user", "foo")
	searchResult, err := esClient.Search().
		Index(indexPattern). // search in index pattern "index-zz-test-tab-*"
		Query(termQuery).    // specify the query
		Sort("user", true).  // sort by "user" field, ascending
		From(0).Size(15).    // take documents 0-9 (default size 10)
		Pretty(true).Do(ctx) // pretty print request and response JSON
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
	deleteIndex, err := esClient.DeleteIndex(indexBaseName).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !deleteIndex.Acknowledged {
		t.Fatal(fmt.Errorf("Delete action not Acknowledged"))
	}
	fmt.Println("Delete index done")
}

/*
Rollover Index

Kibana:
1. create lifecycle policy: zz-test-logs-policy
  - phase: hot, warm, cold
  - hot phase: days, size, number of docs
2. create index template: index-zz-test-tab-tmpl
  - define settings, mappings
  - NOTE: do not define alias here
3. bind lifecycle (rollover) policy with index template
  - select newly created template
  - set alias for rollover index: index-zz-test-tab-write
    - it add settings: "rollover_alias": "index-zz-test-tab-write"
4. create index search pattern: index-zz-test-tab-*

Golang:
1. create index and write alias:
 - index: index-zz-test-tab-000001
 - alias: index-zz-test-tab-write
2. upload data to index by write alias: index-zz-test-tab-write
3. search doc by index pattern: index-zz-test-tab-*
*/

func TestElasticAddEmptyIndexWithAlias(t *testing.T) {
	esClient := ElasticInitForTest()
	resp, err := esClient.CreateIndex(indexBaseName).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Acknowledged {
		t.Fatal(fmt.Errorf("Create index not acknowledged"))
	}

	addAction := elastic.NewAliasAddAction(indexWriteAlias).Index(indexBaseName).IsWriteIndex(true)
	if err := addAction.Validate(); err != nil {
		t.Fatal(err)
	}

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
	resp, err := esClient.RolloverIndex(indexWriteAlias).Conditions(conditions).Do(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// here, results is not acknowledged, and we create rollover index on kibana instead
	if !resp.Acknowledged {
		t.Fatal(fmt.Errorf("Create rollover policy not acknowledged"))
	}
}

/*
ES dsl bool query:
{
  "query": {
    "bool": {
      "must": [
        {
          "term": {"shape": "round"}
        },
        {
          "bool": {
            "should": [
              {"term": {"color": "red"}},
              {"term": {"color": "blue"}}
            ]
          }
        }
      ]
    }
  }
}
*/

func TestElasticDslBoolQuery(t *testing.T) {
	// bool query
	// term query for keyword field
	matches := []elastic.Query{
		elastic.NewTermQuery("env", "test"),
		elastic.NewTermQuery("cid", "cn"),
	}
	query := elastic.NewBoolQuery().Must(matches...)

	// OR sub bool query
	// match query for full match of text field
	shouldMatches := make([]elastic.Query, 0, 2)
	for _, app := range []string{"app-sum-v0.1-beta", "app-min-v0.0-20221215"} {
		shouldMatches = append(shouldMatches, elastic.NewMatchQuery("app_name", app))
	}
	appsQuery := elastic.NewBoolQuery().Should(shouldMatches...)
	query.Must(appsQuery)

	s, err := query.Source()
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("dsl bool query:\n%s\n", b)
}
