package solrbulk

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
)

// Application Version
const Version = "0.1.1"

// Options represents bulk indexing options
type Options struct {
	Host       string
	Port       int
	BatchSize  int
	CommitSize int
	Verbose    bool
}

// BulkIndex takes a set of documents as strings and indexes them into elasticsearch
func BulkIndex(docs []string, options Options) error {
	url := fmt.Sprintf("http://%s:%d/solr/update", options.Host, options.Port)
	var lines []string
	for _, doc := range docs {
		if len(strings.TrimSpace(doc)) == 0 {
			continue
		}
		lines = append(lines, doc)
	}
	body := fmt.Sprintf("[%s]\n", strings.Join(lines, ","))
	response, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		log.Fatal(response.Status)
	}
	// > Caller should close resp.Body when done reading from it.
	// Results in a resource leak otherwise.
	response.Body.Close()
	return nil
}

// Worker will batch index documents that come in on the lines channel
func Worker(id string, options Options, lines chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var docs []string
	counter := 0
	for s := range lines {
		docs = append(docs, s)
		counter++
		if counter%options.BatchSize == 0 {
			err := BulkIndex(docs, options)
			if err != nil {
				log.Fatal(err)
			}
			if options.Verbose {
				log.Printf("[%s] @%d\n", id, counter)
			}
			docs = docs[:0]
		}
	}
	err := BulkIndex(docs, options)
	if err != nil {
		log.Fatal(err)
	}
	if options.Verbose {
		log.Printf("[%s] @%d\n", id, counter)
	}
}
