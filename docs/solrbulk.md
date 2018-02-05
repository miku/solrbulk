SOLRBULK 1 "JANUAR 2015" "Leipzig University Library" "Manuals"
=================================================================

NAME
----

solrbulk - send documents to SOLR in bulk and in parallel.

SYNOPSIS
--------

`solrbulk` [`-server` *URL*, `-size` *N*, `-w` *N*, `-commit` *N*, `-z`] *file*

DESCRIPTION
-----------

solrbulk takes as input a newline delimited JSON file and indexes all documents
into SOLR running on a given server address. The documents are batched and
indexed in parallel to achieve a high indexing throughput.

OPTIONS
-------

`-collection` *string*
  Collection or core name. Deprecated, use `-server`.

`-commit` *N*
  Issue a commit every N documents

`-cpuprofile` *filename*
  Write cpu profile to given filename.

`-host` *string*
  SOLR hostname. Deprecated, use `-server`.

`-memprofile` *filename*
  Write memory profile to given filename.

`-optimize`
  Issue an optimize after final commit.

`-port` *N*
  SOLR port. Deprecated, use `-server`.

`-purge`
  Remove documents from index before indexing (use purge-query to selectively clean). No questions asked.

`-purge-query` *query*
  Query to use, when purging (default "*:*").

`-reset`
  Remove all documents from the given index. No questions asked.

`-server` *URL*
  SOLR URL including host, port and core, like http://localhost:8983/solr/biblio. Currently, the `update` route is implied and fixed.

`-size` *N*
  Batch size.

`-v`
  Program version.

`-verbose`
  Show progress.

`-w` *N*
  Number of workers.

`-z`
  Decompress gzip input file on the fly.

EXAMPLES
--------

Index a file:

  `solrbulk -verbose -server 110.81.131.200:8080/solr/biblio file.ldj`

Index from compressed standard input:

  `cat file.ldj.gz | solrbulk -z -server 110.81.131.200:8080/solr/biblio`

Index a file, but remove all docs from the index before indexing:

  `solrbulk -purge 110.81.131.200:8080/solr/biblio file.ldj`

BUGS
----

Please report bugs to https://github.com/miku/solrbulk/issues.

AUTHOR
------

Martin Czygan <martin.czygan@uni-leipzig.de>

SEE ALSO
--------

[FINC](https://finc.info), [AMSL](http://amsl.technology/), [esbulk](https://github.com/miku/esbulk)
