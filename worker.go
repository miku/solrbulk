// Package solrbulk implements bulk SOLR imports.
//
//  Copyright 2015 by Leipzig University Library, http://ub.uni-leipzig.de
//                 by The Finc Authors, http://finc.info
//                 by Martin Czygan, <martin.czygan@uni-leipzig.de>
//
// This file is part of some open source application.
//
// Some open source application is free software: you can redistribute
// it and/or modify it under the terms of the GNU General Public
// License as published by the Free Software Foundation, either
// version 3 of the License, or (at your option) any later version.
//
// Some open source application is distributed in the hope that it will
// be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
// of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
//
// @license GPL-3.0+ <http://spdx.org/licenses/GPL-3.0+>
//
package solrbulk

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Version of application.
const Version = "0.2.7"

// Options holds bulk indexing options.
type Options struct {
	Collection string
	BatchSize  int
	CommitSize int
	Verbose    bool
	Server     string
}

// BulkIndex takes a set of documents as strings and indexes them into SOLR.
func BulkIndex(docs []string, options Options) error {
	link := fmt.Sprintf("%s/update", options.Server)

	var lines []string
	for _, doc := range docs {
		if len(strings.TrimSpace(doc)) == 0 {
			continue
		}
		lines = append(lines, doc)
	}

	body := fmt.Sprintf("[%s]\n", strings.Join(lines, ","))
	resp, err := http.Post(link, "application/json", strings.NewReader(body))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, resp.Body); err != nil {
			return err
		}
		log.Printf("%s: %s", link, buf.String())
		log.Fatal(resp.Status)
	}
	return resp.Body.Close()
}

// Worker will batch index documents from lines channel.
func Worker(id string, options Options, lines chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var docs []string
	i := 0
	for s := range lines {
		docs = append(docs, s)
		i++
		if i%options.BatchSize == 0 {
			msg := make([]string, len(docs))
			if n := copy(msg, docs); n != len(docs) {
				log.Fatalf("%d docs in batch, but only %d copied", len(docs), n)
			}

			if err := BulkIndex(msg, options); err != nil {
				log.Fatal(err)
			}
			if options.Verbose {
				log.Printf("[%s] @%d\n", id, i)
			}
			docs = nil
		}
	}
	if len(docs) == 0 {
		return
	}
	msg := make([]string, len(docs))
	copy(msg, docs)

	if err := BulkIndex(msg, options); err != nil {
		log.Fatal(err)
	}
	if options.Verbose {
		log.Printf("[%s] @%d\n", id, i)
	}
}
