solrbulk
========

WIP!

> Sometimes you need to index a bunch of documents really, really fast.
  Even with Solr 4.0 and soft commits, if you send one document at a time
  you will be limited by the network. The solution is two-fold: batching
  and multi-threading.


Usage
-----

    $ solrbulk
    Usage: ./solrbulk [OPTIONS] JSON
      -commit=10000: commit after this many docs
      -cpuprofile="": write cpu profile to file
      -host="localhost": elasticsearch host
      -memprofile="": write heap profile to file
      -port=8983: SOLR port
      -size=1000: bulk batch size
      -v=false: prints current program version
      -verbose=false: output basic progress
      -w=4: number of workers to use
      -z=false: unzip gz'd file on the fly

----

Hit [SOLR-6626](https://issues.apache.org/jira/browse/SOLR-6626),
seems to be [fixed](https://svn.apache.org/viewvc?view=revision&revision=1646389) in 5.0.

> NPE in FieldMutatingUpdateProcessor when indexing a doc with null field value
