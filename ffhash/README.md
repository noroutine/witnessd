### ffhash

Fast and fair consistent hash for small clusters (maximum 20 nodes)

#### Pros

* fast (from 400 ns/hash for 20 nodes to 75 ns/hash for 2 nodes)
* close-to-ideal keyspace distribution among cluster nodes (each node gets equal keyspace partition)
* nodes for storage and replicas are calculated and depends only on the key and cluster size (client can directly request data from correct node)
* no central routing or coordination needed
* addition/removal of new node makes only 1/n-th of the data to be reallocated

#### Cons
* number of buckets is n! where n is amount of nodes, so cluster size is limited by capacity of 64-bit integers - 20 nodes is the maximum for now
* complexity of O(n), where n is cluster size

### Stats

Stats for 1000000 random keys on Core i5 3.4GHz 

2 nodes

    bucketRange: 9223372036854775807, buckets: 2, buckets/node: 1
    hash/ms: 13.767771, ns/hash: 72.6334
    Keyspace distribution
      1 : 5000626, deviation: 0.01%
      0 : 4999374, deviation: -0.01%


5 nodes

    bucketRange: 153722867280912930, buckets: 120, buckets/node: 24
    hash/ms: 8.030123, ns/hash: 124.5311
    Keyspace distribution
      1 : 1999196, deviation: -0.04%
      3 : 2000712, deviation: 0.04%
      0 : 1999237, deviation: -0.04%
      4 : 2000709, deviation: 0.04%
      2 : 2000146, deviation: 0.01%


10 nodes

    hash/ms: 4.289224, ns/hash: 233.1424
    Keyspace distribution
      0 : 999327, deviation: -0.07%
      6 : 999946, deviation: -0.01%
      9 : 1001914, deviation: 0.19%
      8 : 998524, deviation: -0.15%
      2 : 999659, deviation: -0.03%
      7 : 1000240, deviation: 0.02%
      5 : 1000845, deviation: 0.08%
      1 : 999628, deviation: -0.04%
      4 : 1001472, deviation: 0.15%
      3 : 998445, deviation: -0.16%

19 nodes

    bucketRange: 151, buckets: 121645100408832000, buckets/node: 6402373705728000
    hash/ms: 2.363837, ns/hash: 423.0411
    Keyspace distribution
      0 : 523954, deviation: -0.45%
      12 : 525333, deviation: -0.19%
      11 : 526442, deviation: 0.02%
      4 : 531324, deviation: 0.95%
      1 : 524644, deviation: -0.32%
      6 : 527485, deviation: 0.22%
      3 : 522909, deviation: -0.65%
      7 : 526982, deviation: 0.13%
      18 : 526715, deviation: 0.08%
      9 : 527589, deviation: 0.24%
      5 : 529507, deviation: 0.61%
      13 : 526286, deviation: -0.01%
      10 : 527307, deviation: 0.19%
      15 : 525449, deviation: -0.16%
      16 : 525807, deviation: -0.10%
      17 : 525308, deviation: -0.19%
      2 : 524951, deviation: -0.26%
      8 : 525372, deviation: -0.18%
      14 : 526636, deviation: 0.06%
