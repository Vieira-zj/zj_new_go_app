package elastic

import (
	"log"
	"sync"

	"github.com/olivere/elastic/v7"
)

var (
	esClient       *elastic.Client
	NewElasticOnce sync.Once
)

func NewElastic(hosts []string, username, pwd string) *elastic.Client {
	NewElasticOnce.Do(func() {
		client, err := elastic.NewClient(
			elastic.SetURL(hosts...),
			elastic.SetBasicAuth(username, pwd),
			elastic.SetHealthcheck(false),
			elastic.SetMaxRetries(5),
		)
		if err != nil {
			log.Fatalln(err)
		}
		esClient = client
	})
	return esClient
}
