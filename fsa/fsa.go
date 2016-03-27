// FSA - finite-state automatae
package fsa

type TransitionFunc func(int, int) int
type TimeoutFunc func(int) chan int
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

var neverWritten chan int = make(chan int, 0)

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
    return func(int) chan int {
        return neverWritten
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
        select {
        case event := <- a.input: a.state = a.exec(a.state, event)
        case tstate := <- a.timeout(a.state): a.state = tstate
        case <- a.term:
            a.Result <- a.state
            close(a.Result)
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