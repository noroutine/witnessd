package cluster

import "errors"

/*

Checks if three given hash values are in clockwise order on the hash ring

On the ring, there could be following situations

Order           Clockwise

a, b, c         true
a, c, b         false
b, a, c         false
b, c, a         true
c, a, b         true
c, b, a         false

if any of two are equal it's not possible to determine the result and error is returned
 */
func Clockwise(a, b, c[]byte) (bool, error) {
    bc := CompareHashes(b, c)
    if bc == 0 {
        return false, errors.New("b == c")
    }

    if bc < 0 {
        ba := CompareHashes(b, a)
        ac := CompareHashes(a, c)

        if ba == 0 {
            return false, errors.New("a == b")
        }

        if ac == 0 {
            return false, errors.New("a == c")
        }

        return ! (ba < 0 && ac < 0), nil
    } else {
        ca := CompareHashes(c, a)
        if ca == 0 {
            return false, errors.New("a == c")
        }

        ab := CompareHashes(a, b)
        if ab == 0 {
            return false, errors.New("a == b")
        }

        return ca < 0 && ab < 0, nil
    }
}

// 1   if a > b
// 0   if a == b
// -1  if a < b
func CompareHashes(a, b []byte) int {
    var i int
    for i = hash_byte_len - 1; i > 0 && a[i] == b[i]; i-- {
    }

    switch {
    case a[i] > b[i]: return 1
    case a[i] < b[i]: return  -1
    }

    return 0
}


