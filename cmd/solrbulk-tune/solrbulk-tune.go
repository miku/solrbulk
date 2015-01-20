package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"time"
)

func main() {
	defaultPath, _ := exec.LookPath("solrbulk")

	host := flag.String("host", "localhost", "SOLR host")
	port := flag.Int("port", 8983, "SOLR port")
	path := flag.String("path", defaultPath, "path to solrbulk")

	flag.Parse()

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

	var cmd *exec.Cmd
	for _, w := range workers {
		for _, size := range bsizes {
			for _, commit := range csizes {
				s := fmt.Sprintf("%s -host %s -port %d -verbose -reset", *path, *host, *port)
				log.Println(s)
				cmd = exec.Command(s)
				err := cmd.Run()
				if err != nil {
					log.Fatal(err)
				}

				start := time.Now()
				s = fmt.Sprintf("%s -host %s -port %d -verbose -w %d -size %d -commit %d %s", *path, *host, *port, w, size, commit, filename)
				log.Println(s)
				cmd = exec.Command(s)
				err = cmd.Run()
				if err != nil {
					log.Fatal(err)
				}
				elapsed := time.Since(start)
				log.Printf("Indexing took %s", elapsed)
			}
		}
	}
}
