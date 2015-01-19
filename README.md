solrbulk
========

This software is functional, yet work-in-progress. Use with care.

Motivation:

> Sometimes you need to index a bunch of documents really, really fast.
  Even with Solr 4.0 and soft commits, if you send one document at a time
  you will be limited by the network. The solution is two-fold: batching
  and multi-threading. http://lucidworks.com/blog/high-throughput-indexing-in-solr/

solrbulk expects as input a file with [line-delimited JSON](http://en.wikipedia.org/wiki/Line_Delimited_JSON). Each line
represents a single document. solrbulk takes care of reformatting the documents
into the bulk JSON format, that [SOLR understands](https://wiki.apache.org/solr/UpdateJSON).

solrbulk will send documents in batches and in parallel. The number of documents
per batch can be set via `-size`, the number of workers with `-w`.

Usage
-----

    $ solrbulk
    Usage: solrbulk [OPTIONS] FILE
      -commit=5000000: commit after this many docs
      -cpuprofile="": write cpu profile to file
      -host="localhost": elasticsearch host
      -memprofile="": write heap profile to file
      -port=8983: SOLR port
      -reset=false: remove all docs from index
      -size=10000: bulk batch size
      -v=false: prints current program version
      -verbose=false: output basic progress
      -w=4: number of workers to use
      -z=false: unzip gz'd file on the fly

Example
-------

    $ cat file.ldj
    {"id": "1", "state": "Alaska"}
    {"id": "2", "state": "California"}
    {"id": "3", "state": "Oregon"}
    ...

    $ solrbulk -verbose -host 192.168.1.222 -port 8085 file.ldj

Some performance observations
-----------------------------

* Having as many workers as core is generally a good idea. However the returns seem to diminish with more cores.
* Disable `autoCommit` and `autoSoftCommit` in `solrconfig.xml`.
* Use some high number for `-commit`. solrbulk will issue a final commit request at the end of the processing anyway.
* For some use cases, the bulk indexing approach is about twice as fast as a standard request to `/solr/update`.

----

Hit [SOLR-6626](https://issues.apache.org/jira/browse/SOLR-6626),
seems to be [fixed](https://svn.apache.org/viewvc?view=revision&revision=1646389) in 5.0.

> NPE in FieldMutatingUpdateProcessor when indexing a doc with null field value
