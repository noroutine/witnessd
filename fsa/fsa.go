// FSA - finite-state automatae, time bound
package fsa

import (
    "time"
)

type TransitionFunc func(int, int) int
type TimeoutFunc func(int) (<-chan time.Time, func(int) int)
type TerminateFunc func(int) bool

type FSA struct  {
    Result chan int
    state int
    timeout TimeoutFunc 
    input chan int          // input events
    exec TransitionFunc
    end TerminateFunc
    term chan bool
}

var neverTicks chan time.Time = make(chan time.Time, 0)

func NeverTerminates() TerminateFunc {
    return func(s int) bool {
        return false
    }
}

func TerminatesOn(states ...int) TerminateFunc {
    var tss map[int]bool = make(map[int]bool, len(states))
    for _, ts := range states {
        tss[ts] = true
    }

    return func(s int) bool {
        _, ok := tss[s]
        return ok
    }
}

func NeverTimesOut() TimeoutFunc {
    return func(int) (<-chan time.Time, func(int) int) {
        return neverTicks, func(s int) int {
            return s
        }
    }
}

func New(e TransitionFunc, end TerminateFunc, tout TimeoutFunc) (a *FSA) {    
    a = &FSA{
        Result: make(chan int),
        state: 0,
        input: make(chan int),
        exec: e,
        timeout: tout,
        end: end,
        term: make(chan bool),
    }
    go a.run()
    return
}

func (a *FSA) run() {
    for {
        timeTick, tfunc := a.timeout(a.state)
        select {
        case event := <- a.input: a.state = a.exec(a.state, event)
        case <- timeTick: a.state = tfunc(a.state)
        case <- a.term:
            a.Result <- a.state
            close(a.Result)
            close(a.input)
            close(a.term)
            return
        }

        if a.end(a.state) {
            go a.Terminate()
        }
    }
}

func (a *FSA) Send(input int) {
    a.input <- input
}

func (a *FSA) Terminate() {
    a.term <- true
}