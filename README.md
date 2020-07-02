# solrbulk

Motivation:

> Sometimes you need to index a bunch of documents really, really fast.
  Even with Solr 4.0 and soft commits, if you send one document at a time
  you will be limited by the network. The solution is two-fold: batching
  and multi-threading. http://lucidworks.com/blog/high-throughput-indexing-in-solr/

solrbulk expects as input a file with [line-delimited JSON](https://en.wikipedia.org/wiki/JSON_Streaming#Line-delimited_JSON). Each line represents a single document. solrbulk takes care of reformatting the documents into the bulk JSON format, that [SOLR understands](https://cwiki.apache.org/confluence/display/solr/Uploading+Data+with+Index+Handlers#UploadingDatawithIndexHandlers-JSONFormattedIndexUpdates).

solrbulk will send documents in batches and in parallel. The number of documents per batch can be set via `-size`, the number of workers with `-w`.

[![Project Status: Active – The project has reached a stable, usable state and is being actively developed.](https://www.repostatus.org/badges/latest/active.svg)](https://www.repostatus.org/#active)

This project has been developed for [Project finc](https://finc.info) at [Leipzig University Library](https://ub.uni-leipzig.de).

## Installation

Installation via Go tools.

    $ go get github.com/miku/solrbulk/cmd/...

There are also DEB, RPM and
[arch](https://github.com/miku/solrbulk/blob/master/arch/PKGBUILD) packages
available at
[https://github.com/miku/solrbulk/releases/](https://github.com/miku/solrbulk/releases/).

## Usage

Flags.

    $ solrbulk
    Usage of solrbulk:
      -commit int
            commit after this many docs (default 1000000)
      -cpuprofile string
            write cpu profile to file
      -memprofile string
            write heap profile to file
      -no-final-commit
            omit final commit (at end of file or stdin)
      -optimize
            optimize index
      -purge
            remove documents from index before indexing (use purge-query to selectively clean)
      -purge-pause duration
            insert a short pause after purge (default 2s)
      -purge-query string
            query to use, when purging (default "*:*")
      -server string
            url to SOLR server, including host, port and path to collection
      -size int
            bulk batch size (default 1000)
      -update-request-handler-name string
            where solr.UpdateRequestHandler is mounted on the server, https://is.gd/s0eirv (default "/update")
      -v    prints current program version
      -verbose
            output basic progress
      -w int
            number of workers to use (default 4)
      -z    unzip gz'd file on the fly

## Example

Given a [newline delimited JSON](http://jsonlines.org/) file:

    $ cat file.ldj
    {"id": "1", "state": "Alaska"}
    {"id": "2", "state": "California"}
    {"id": "3", "state": "Oregon"}
    ...

    $ solrbulk -verbose -server https://192.168.1.222:8085/collection1 file.ldj

The server parameter contains host, port and path up to, but excluding the
default [*update*
route](https://lucene.apache.org/solr/guide/6_6/uploading-data-with-index-handlers.html)
for search (since 0.3.4, this can be adjusted via
`-update-request-handler-name` flag).

For example, if you usually update via `https://192.168.1.222:8085/solr/biblio/update` the server parameter would be:

    $ solrbulk -server https://192.168.1.222:8085/solr/biblio file.ldj


## Some performance observations

* Having as many workers as core is generally a good idea. However the returns seem to diminish fast with more cores.
* Disable `autoCommit`, `autoSoftCommit` and the transaction log in `solrconfig.xml`.
* Use some high number for `-commit`. solrbulk will issue a final commit request at the end of the processing anyway.
* For some use cases, the bulk indexing approach is about twice as fast as a standard request to `/solr/update`.
* On machines with more cores, try to increase [maxIndexingThreads](https://cwiki.apache.org/confluence/display/solr/IndexConfig+in+SolrConfig).

## Elasticsearch?

Try [esbulk](https://github.com/miku/esbulk).
