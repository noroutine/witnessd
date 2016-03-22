package cluster

import (
    "testing"
    "flag"
    "os"
)

func TestMain(m *testing.M) {
    flag.Parse()
    os.Exit(m.Run())
}

func TestMarshalling(t *testing.T) {
    m := Message{
        Version:    1,
        OP:         0xFF,
        Args:       make([]byte, 8, 8),
        Reserved:   make([]byte, 8, 8),
        Length:     1,
        Load:       []byte { 1 },
    }

    m1, err := Unmarshall(Marshall(&m))

    if err != nil {
        t.Fail()        
    }

    if m1.Version != m.Version {
        t.Fail()
    }

    if m1.OP != m.OP {
        t.Fail()
    }

    if m1.Length != m.Length {
        t.Fail()
    }

    if m1.Load[0] != m.Load[0] {
        t.Fail()
    }
}
