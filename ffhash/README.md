### ffhash

Fast and fair consistent hash for small clusters (best < 18 nodes, maximum 20)

Hashes 64-bit key to node, where data for that key stored

#### Pros

* fast (75 ns/hash for 2 nodes, 230ns/hash for 10 nodes, 400 ns/hash for 18 nodes)
* close-to-ideal keyspace distribution among cluster nodes (each node gets equal keyspace partition, so no worries about unfair load)
* location for storage and replicas cam be calculated upfront and depend only on the key and cluster size, so client can directly request data from correct node
* no central routing or coordination needed
* addition/removal of new node makes only 1/n-th of the data to be reallocated

#### Cons
* number of buckets is n! where n is amount of nodes, so cluster size is limited by capacity of 64-bit integers - 20 nodes is the maximum for now
* linear growth of hashing time with cluster size
* with 19 and 20 nodes the bucket size is critically low and division errors start to significantly distort the partitioning

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

20 nodes, notice small bucket range and partitioning distortion for nodes 0, 1, 2 and 3

    bucketRange: 7, buckets: 2432902008176640000, buckets/node: 121645100408832000
    hash/ms: 2.240514, ns/hash: 446.3262
    Keyspace distribution
      14 : 499591, deviation: -0.08%
      17 : 499333, deviation: -0.13%
      16 : 499557, deviation: -0.09%
      4 : 499666, deviation: -0.07%
      0 : 462724, deviation: -7.46%
      19 : 499735, deviation: -0.05%
      1 : 462193, deviation: -7.56%
      18 : 499294, deviation: -0.14%
      15 : 499992, deviation: -0.00%
      7 : 500537, deviation: 0.11%
      10 : 499102, deviation: -0.18%
      5 : 500229, deviation: 0.05%
      12 : 499516, deviation: -0.10%
      3 : 538012, deviation: 7.60%
      8 : 499442, deviation: -0.11%
      2 : 538406, deviation: 7.68%
      6 : 501391, deviation: 0.28%
      11 : 499474, deviation: -0.11%
      9 : 500612, deviation: 0.12%
      13 : 501194, deviation: 0.24%
