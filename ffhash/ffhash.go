package ffhash

/**

ffhash.go

Fast and fair consistent hash for small clusters (best < 18 nodes, maximum 20)
Hashes 64-bit key to node, where data for that key stored

*/

func Sum64(key uint64, n uint64) uint64 {
    var totalBuckets uint64 = Fact(n)
    bucketRange := 0xFFFFFFFFFFFFFFFF / totalBuckets
    bucket := key / bucketRange
    return fairfast(bucket, n, 0, totalBuckets)
}

func fairfast(bucket uint64, n uint64, rangeStart uint64, rangeEnd uint64) uint64 {
    var i uint64
    var f, l uint64 = 0, 1

    for i = 2; i <= n; i++ {

        pivot := rangeStart + (rangeEnd - rangeStart) / i
        size := pivot - rangeStart        
        if bucket < pivot {
            // exclude first
            f, l = l, i
            rangeEnd = pivot

        } else {
            // exclude last
            f, l = f, i

            // quickly find the range for the node
            var cut uint64
            for cut = rangeStart; bucket >= cut + size; cut = cut + size { 
            }

            rangeStart, rangeEnd = cut, cut + size
        }
    }

    if l == n {
        return f
    } else {
        return l
    }
}

func Fact(n uint64) uint64 {
    switch n {
        case 1:  return 1
        case 2:  return 2
        case 3:  return 6
        case 4:  return 24
        case 5:  return 120
        case 6:  return 720
        case 7:  return 5040
        case 8:  return 40320
        case 9:  return 362880
        case 10: return 3628800
        case 11: return 39916800
        case 12: return 479001600
        case 13: return 6227020800
        case 14: return 87178291200
        case 15: return 1307674368000
        case 16: return 20922789888000
        case 17: return 355687428096000
        case 18: return 6402373705728000
        case 19: return 121645100408832000
        case 20: return 2432902008176640000
        // for 21 it's already bigger than uint64
        default: panic("uint64 overflow")
    }
}