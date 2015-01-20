package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/miku/solrbulk"
)

func main() {
	defaultPath, _ := exec.LookPath("solrbulk")

	version := flag.Bool("v", false, "prints current program version")
	verbose := flag.Bool("verbose", false, "output basic progress")
	host := flag.String("host", "localhost", "SOLR host")
	port := flag.Int("port", 8983, "SOLR port")
	path := flag.String("path", defaultPath, "path to solrbulk")
	header := flag.Bool("header", false, "output header row")
	limit := flag.Int("limit", 10000, "number of docs in sample file")

	flag.Parse()

	if *version {
		fmt.Println(solrbulk.Version)
		os.Exit(0)
	}

	if *path == "" {
		log.Fatal("solrbulk required")
	}

	if flag.NArg() < 1 {
		log.Fatal("need a suitable sample file")
	}

	filename := flag.Arg(0)

	workers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 24, 32, 48, 64, 128, 256, 512, 1024}
	bsizes := []int{1, 10, 50, 100, 500, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 12000, 14000, 16000, 18000, 20000, 30000, 40000, 50000}
	csizes := []int{1000, 5000, 10000, 20000, 50000, 100000, 200000, 500000, 1000000, 5000000, 10000000, 20000000}

	if *header {
		fmt.Printf("w\tsize\tcommit\telapsed\n")
	}

	for _, w := range workers {
		for _, size := range bsizes {
			if size > *limit {
				continue
			}
			for _, commit := range csizes {
				if commit > *limit {
					continue
				}
				cmd := exec.Command(*path, "-host", *host, "-port", strconv.Itoa(*port), "-verbose", "-reset")
				if *verbose {
					log.Println(strings.Join(cmd.Args, " "))
				}
				err := cmd.Run()
				if err != nil {
					log.Fatal(err)
				}

				start := time.Now()
				cmd = exec.Command(*path, "-host", *host, "-port", strconv.Itoa(*port), "-verbose", "-w", strconv.Itoa(w), "-size", strconv.Itoa(size), "-commit", strconv.Itoa(commit), filename)
				if *verbose {
					log.Println(strings.Join(cmd.Args, " "))
				}
				err = cmd.Run()
				if err != nil {
					log.Println(err)
					fmt.Printf("%d\t%d\t%d\tFAILED\n", w, size, commit)
					continue
				}

				elapsed := time.Since(start)
				if *verbose {
					log.Printf("indexing took %s", elapsed)
				}
				fmt.Printf("%d\t%d\t%d\t%0.3f\n", w, size, commit, elapsed.Seconds())
			}
		}
	}
}
