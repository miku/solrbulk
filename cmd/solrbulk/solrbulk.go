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
package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"time"

	"github.com/miku/solrbulk"
)

func main() {

	version := flag.Bool("v", false, "prints current program version")
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	memprofile := flag.String("memprofile", "", "write heap profile to file")
	host := flag.String("host", "localhost", "SOLR host (deprecated, use -server)")
	port := flag.Int("port", 8983, "SOLR port (deprecated, use -server)")
	collection := flag.String("collection", "", "SOLR core / collection")
	batchSize := flag.Int("size", 1000, "bulk batch size")
	commitSize := flag.Int("commit", 1000000, "commit after this many docs")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers to use")
	verbose := flag.Bool("verbose", false, "output basic progress")
	gzipped := flag.Bool("z", false, "unzip gz'd file on the fly")
	reset := flag.Bool("reset", false, "remove all docs from index")
	server := flag.String("server", "", "url to SOLR server, including host, port and path to collection")
	optimize := flag.Bool("optimize", false, "optimize index")

	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *version {
		fmt.Println(solrbulk.Version)
		os.Exit(0)
	}

	options := solrbulk.Options{
		Collection: *collection,
		BatchSize:  *batchSize,
		CommitSize: *commitSize,
		Verbose:    *verbose,
	}

	// Assemble a new server option, that behaves like the old one, if -server
	// is not specified.
	var srv string
	if *server != "" {
		srv = *server
	} else {
		if *collection != "" {
			srv = fmt.Sprintf("http://%s:%d/solr/%s", *host, *port, options.Collection)
		} else {
			srv = fmt.Sprintf("http://%s:%d/solr", *host, *port)
		}
	}

	if !strings.HasPrefix(srv, "http") {
		srv = fmt.Sprintf("http://%s", srv)
	}
	options.Server = srv

	if *reset {
		hostpath := fmt.Sprintf("%s/update", options.Server)
		urls := []string{
			fmt.Sprintf("%s?stream.body=<delete><query>*:*</query></delete>", hostpath),
			fmt.Sprintf("%s?stream.body=<commit/>", hostpath),
		}
		for _, url := range urls {
			resp, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("%s %s", resp.Status, url)
		}
		os.Exit(0)
	}

	var file io.Reader = os.Stdin

	if flag.NArg() > 0 {
		f, err := os.Open(flag.Args()[0])
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		file = f
	}

	runtime.GOMAXPROCS(*numWorkers)

	queue := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go solrbulk.Worker(fmt.Sprintf("worker-%d", i), options, queue, &wg)
	}

	commitURL := fmt.Sprintf("%s/update?commit=true", options.Server)

	// final commit
	defer func() {
		resp, err := http.Get(commitURL)
		if err != nil {
			log.Fatal(err)
		}
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
		log.Printf("final commit: %s\n", resp.Status)
	}()

	reader := bufio.NewReader(file)
	if *gzipped {
		zreader, err := gzip.NewReader(file)
		if err != nil {
			log.Fatal(err)
		}
		reader = bufio.NewReader(zreader)
	}

	i := 0
	start := time.Now()

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
			resp, err := http.Get(commitURL)
			if err != nil {
				log.Fatal(err)
			}
			if options.Verbose {
				log.Printf("commit @%d %s\n", i, resp.Status)
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
		pprof.WriteHeapProfile(f)
		f.Close()
	}

	if *verbose {
		elapsed := time.Since(start)
		rate := float64(i) / elapsed.Seconds()
		log.Printf("%d docs in %s at %0.3f docs/s with %d workers\n", i, elapsed, rate, *numWorkers)
	}

	if *optimize {
		hostpath := fmt.Sprintf("%s/update", options.Server)
		url := fmt.Sprintf("%s?stream.body=<optimize/>", hostpath)
		resp, err := http.Get(url)
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
