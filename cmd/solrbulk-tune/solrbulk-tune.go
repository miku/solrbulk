package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/miku/solrbulk"
)

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32784)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

func main() {
	defaultPath, _ := exec.LookPath("solrbulk")

	version := flag.Bool("v", false, "prints current program version")
	verbose := flag.Bool("verbose", false, "output basic progress")
	host := flag.String("host", "localhost", "SOLR host")
	port := flag.Int("port", 8983, "SOLR port")
	path := flag.String("path", defaultPath, "path to solrbulk")
	header := flag.Bool("header", false, "output header row")
	maxRetry := flag.Int("retry", 25, "retry count for index cleanup")

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

	ff, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	reader := bufio.NewReader(ff)
	lines, err := lineCounter(reader)
	if err != nil {
		log.Fatal(err)
	}
	err := ff.Close()
	if err != nil {
		log.Fatal(err)
	}

	workers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 24, 32, 48, 64, 128, 256, 512, 1024}
	bsizes := []int{1, 10, 50, 100, 500, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 12000, 14000, 16000, 18000, 20000, 30000, 40000, 50000}
	csizes := []int{1000, 5000, 10000, 20000, 50000, 100000, 200000, 500000, 1000000, 5000000, 10000000, 20000000}

	if *header {
		fmt.Printf("w\tsize\tcommit\telapsed\n")
	}

	for _, w := range workers {
		for _, size := range bsizes {
			if size > lines {
				continue
			}
			for _, commit := range csizes {
				if commit > lines {
					continue
				}
				retry := 0
				var cmd *exec.Cmd
				for {
					cmd = exec.Command(*path, "-host", *host, "-port", strconv.Itoa(*port), "-verbose", "-reset")
					if *verbose {
						log.Println(strings.Join(cmd.Args, " "))
					}
					err := cmd.Run()
					if err == nil {
						break
					} else {
						if retry == *maxRetry {
							log.Fatal(err)
						} else {
							log.Println(err)
						}
					}
					retry++
					time.Sleep(5 * time.Second)
					log.Printf("retry [%d]...", retry)
				}

				start := time.Now()
				cmd = exec.Command(*path, "-host", *host, "-port", strconv.Itoa(*port), "-verbose", "-w", strconv.Itoa(w), "-size", strconv.Itoa(size), "-commit", strconv.Itoa(commit), filename)
				if *verbose {
					log.Println(strings.Join(cmd.Args, " "))
				}

				err := cmd.Run()
				if err != nil {
					log.Println(err)
					elapsed := time.Since(start)
					fmt.Printf("%d\t%d\t%d\tFAILED AFTER %0.3f\n", w, size, commit, elapsed.Seconds())
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
