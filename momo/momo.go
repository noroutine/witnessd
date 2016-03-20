package momo

import (
    //"fmt"
)

func MomoHash32(key int, n int) int {
    totalBuckets := Fact(n)
    keySpace := 1 << 32
    bucketRange := keySpace / totalBuckets
    bucket := key / bucketRange
    // fmt.Printf("keySpace: %v, bucketRange: %v, buckets: %v, bucket: %v\n", keySpace, bucketRange, totalBuckets, bucket)

    return Momo(bucket, n)
}

func Momo(key int, n int) int {
    f, l := 0, 1

    rangeStart := 0
    rangeEnd := Fact(n)

    for i := 2; i <= n; i++ {

        pivot := rangeStart + (rangeEnd - rangeStart) / i
        size := pivot - rangeStart

        // fmt.Printf("size: %v, pivot: %v, [ %v, %v ]\n", size, pivot, rangeStart, rangeEnd)
        
        if size < 0 {
            return -1
        }

        if key < pivot {
            // exclude first
            // fmt.Println("exclude first", f)
            f, l = l, i
            rangeEnd = pivot

        } else {
            // exclude last
            // fmt.Println("exclude last", l)
            f, l = f, i

            // quickly find the range for the node
            var cut int
            for cut = rangeStart; key >= cut + size; cut = cut + size { }

            rangeStart, rangeEnd = cut, cut + size
        }

        // fmt.Printf("[%v, %v, %v] [%v, %v, %v, %v]\n", i, f, l, rangeStart, rangeEnd, pivot, size)
    }

    if l == n {
        return f
    } else {
        return l
    }
}

func Fact(n int) int {
    if n == 1 {
        return 1
    }
    return Fact(n - 1) * n
}