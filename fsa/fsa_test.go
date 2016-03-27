package fsa

import (
    "testing"
    "flag"
    "time"
    "os"
)

func TestMain(m *testing.M) {
    flag.Parse()
    os.Exit(m.Run())
}

func TestTrivial(t *testing.T) {
    trivial := New(func(state, input int) int {
        return input
    }, NeverTerminates(), NeverTimesOut())

    go func() {
        trivial.Send(4)
        trivial.Terminate()
    }()

    select {
    case result := <- trivial.Result:
        if result != 4 {
            t.Error("wrong state")
        }
    case <- time.After(10 * time.Millisecond): t.Error("deadlock")
    }
}

func TestTermination(t *testing.T) {
    trivial := New(func(state, input int) int {
        return input
    }, TerminatesOn(4, 5), NeverTimesOut())

    go func() {
        trivial.Send(5)
    }()

    select {
    case result := <- trivial.Result:
        if !(result == 4 || result == 5) {
            t.Error("wrong state")
        }
    case <- time.After(10 * time.Millisecond): t.Error("deadlock")
    }
}

func TestTerminateOn(t *testing.T) {
    tf := TerminatesOn(1, 3, 5, 7)

    if ! (tf(1) && tf(3) && tf(5) && tf(7) && !tf(0) && !tf(2) && !tf(4) && !tf(6) && !tf(8)) {
        t.Error("terminates on wrong state")
    }
}