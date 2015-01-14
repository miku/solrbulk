solrbulk
========

> Sometimes you need to index a bunch of documents really, really fast.
  Even with Solr 4.0 and soft commits, if you send one document at a time
  you will be limited by the network. The solution is two-fold: batching
  and multi-threading.

