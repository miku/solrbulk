// solrbulk sends documents to a SOLR server.
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

package main

import (
	"bufio"
	b64 "encoding/base64"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	gzip "github.com/klauspost/compress/gzip"
	"github.com/miku/solrbulk"
	"github.com/sethgrid/pester"
	log "github.com/sirupsen/logrus"
)

var (
	version                  = flag.Bool("v", false, "prints current program version")
	cpuprofile               = flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile               = flag.String("memprofile", "", "write heap profile to file")
	batchSize                = flag.Int("size", 1000, "bulk batch size")
	commitSize               = flag.Int("commit", 1000000, "commit after this many docs")
	numWorkers               = flag.Int("w", runtime.NumCPU(), "number of workers to use")
	verbose                  = flag.Bool("verbose", false, "output basic progress")
	gzipped                  = flag.Bool("z", false, "unzip gz'd file on the fly")
	server                   = flag.String("server", "", "url to SOLR server, including host, port and path to collection, e.g. http://localhost:8983/solr/biblio")
	optimize                 = flag.Bool("optimize", false, "optimize index")
	purge                    = flag.Bool("purge", false, "remove documents from index before indexing (use purge-query to selectively clean)")
	purgeQuery               = flag.String("purge-query", "*:*", "query to use, when purging")
	purgePause               = flag.Duration("purge-pause", 2*time.Second, "insert a short pause after purge")
	updateRequestHandlerName = flag.String("update-request-handler-name", "/update", "where solr.UpdateRequestHandler is mounted on the server, https://is.gd/s0eirv")
	noFinalCommit            = flag.Bool("no-final-commit", false, "omit final commit")
	basicAuth                = flag.String("auth", "", "username:password pair for basic auth")
)

func newGetRequest(url string, options solrbulk.Options) (*http.Request, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if options.BasicAuth != "" {
		req.Header.Add("Authorization", "Basic "+b64.StdEncoding.EncodeToString([]byte(options.BasicAuth)))
	}
	return req, nil
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal(err)
		}
		defer pprof.StopCPUProfile()
	}
	if *version {
		fmt.Println(solrbulk.Version)
		os.Exit(0)
	}
	options := solrbulk.Options{
		BatchSize:                *batchSize,
		CommitSize:               *commitSize,
		Verbose:                  *verbose,
		UpdateRequestHandlerName: *updateRequestHandlerName,
		Server:                   *server,
		BasicAuth:                *basicAuth,
	}
	if !strings.HasPrefix(options.Server, "http") {
		options.Server = fmt.Sprintf("http://%s", options.Server)
	}
	if *purge {
		var (
			hostpath = fmt.Sprintf("%s%s", options.Server, options.UpdateRequestHandlerName)
			urls     = []string{
				fmt.Sprintf("%s?stream.body=<delete><query>%s</query></delete>", hostpath, *purgeQuery),
				fmt.Sprintf("%s?stream.body=<commit/>", hostpath),
			}
		)
		for _, url := range urls {
			req, err := newGetRequest(url, options)
			if err != nil {
				log.Fatal(err)
			}

			resp, err := pester.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%s %s", resp.Status, url)
		}
		time.Sleep(*purgePause)
	}
	var file io.Reader = os.Stdin
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		file = f
	}
	var (
		queue     = make(chan string)
		wg        sync.WaitGroup
		commitURL = fmt.Sprintf("%s%s?commit=true", options.Server, options.UpdateRequestHandlerName)
		reader    = bufio.NewReader(file)
	)
	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go solrbulk.Worker(fmt.Sprintf("worker-%d", i), options, queue, &wg)
	}
	if !*noFinalCommit {
		defer func() {
			resp, err := pester.Get(commitURL)
			if err != nil {
				log.Fatal(err)
			}
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
			log.Printf("final commit: %s", resp.Status)
		}()
	}
	if *gzipped {
		zreader, err := gzip.NewReader(reader)
		if err != nil {
			log.Fatal(err)
		}
		reader = bufio.NewReader(zreader)
	}
	var (
		i     = 0
		start = time.Now()
	)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		line = strings.TrimSpace(line)
		queue <- line
		i++
		if i%options.CommitSize == 0 {
			req, err := newGetRequest(commitURL, options)
			if err != nil {
				log.Fatal(err)
			}

			resp, err := pester.Do(req)
			if err != nil {
				log.Fatal(err)
			}
			if options.Verbose {
				log.Printf("commit @%d %s", i, resp.Status)
			}
			if err := resp.Body.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}
	close(queue)
	wg.Wait()
	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal(err)
		}
		f.Close()
	}
	if *verbose {
		elapsed := time.Since(start)
		rate := float64(i) / elapsed.Seconds()
		log.Printf("%d docs in %s at %0.3f docs/s with %d workers", i, elapsed, rate, *numWorkers)
	}
	if *optimize {
		hostpath := fmt.Sprintf("%s%s", options.Server, options.UpdateRequestHandlerName)
		url := fmt.Sprintf("%s?stream.body=<optimize/>", hostpath)

		req, err := newGetRequest(url, options)
		if err != nil {
			log.Fatal(err)
		}

		resp, err := pester.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s %s", resp.Status, url)
		elapsed := time.Since(start)
		if *verbose {
			log.Printf("indexed and optimized in %s", elapsed)
		}
	}
}
