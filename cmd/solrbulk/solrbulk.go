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
	host := flag.String("host", "localhost", "SOLR host")
	port := flag.Int("port", 8983, "SOLR port")
	batchSize := flag.Int("size", 10000, "bulk batch size")
	commitSize := flag.Int("commit", 5000000, "commit after this many docs")
	numWorkers := flag.Int("w", runtime.NumCPU(), "number of workers to use")
	verbose := flag.Bool("verbose", false, "output basic progress")
	gzipped := flag.Bool("z", false, "unzip gz'd file on the fly")
	reset := flag.Bool("reset", false, "remove all docs from index")

	var PrintUsage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] FILE\n", os.Args[0])
		flag.PrintDefaults()
	}

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
		Host:       *host,
		Port:       *port,
		BatchSize:  *batchSize,
		CommitSize: *commitSize,
		Verbose:    *verbose,
	}

	if *reset {
		urls := []string{
			fmt.Sprintf("http://%s:%d/solr/update?stream.body=<delete><query>*:*</query></delete>", options.Host, options.Port),
			fmt.Sprintf("http://%s:%d/solr/update?stream.body=<commit/>", options.Host, options.Port),
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

	if flag.NArg() < 1 {
		PrintUsage()
		os.Exit(1)
	}

	file, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	runtime.GOMAXPROCS(*numWorkers)

	queue := make(chan string)
	var wg sync.WaitGroup

	for i := 0; i < *numWorkers; i++ {
		wg.Add(1)
		go solrbulk.Worker(fmt.Sprintf("worker-%d", i), options, queue, &wg)
	}

	commitUrl := fmt.Sprintf("http://%s:%d/solr/update?commit=true", *host, *port)

	defer func() {
		resp, err := http.Get(commitUrl)
		if err != nil {
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

	counter := 0
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
		counter += 1

		if counter%options.CommitSize == 0 {
			resp, err := http.Get(commitUrl)
			if err != nil {
				log.Fatal(err)
			}
			if options.Verbose {
				log.Printf("commit @%d %s\n", counter, resp.Status)
			}
		}
	}

	close(queue)
	wg.Wait()
	elapsed := time.Since(start)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.WriteHeapProfile(f)
		f.Close()
	}

	if *verbose {
		rate := float64(counter) / elapsed.Seconds()
		log.Printf("%d docs in %s at %0.3f docs/s with %d workers\n", counter, elapsed, rate, *numWorkers)
	}
}
