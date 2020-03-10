# Exploring Generator Pattern in Go

> DDD, CQRS, Actor Model, State Machine, Saga, Aggregate Root
> 
> Generator Pattern, Yield, Lazy, Channel, CSM, Worker


Imagine a `Domain` event called `Order.Placed`. Each `Order.Placed` contains:

```json
{
    "customer": "CUST-001",
    "merchant": "MRCN-001",
    "payment": "CARD-001",
    "items": [
        {"id": "IT-001", "qty": 1},
        {"id": "IT-002", "qty": 2}
    ],
    "promo": "PROM-001"
}
```

Everytime an `order` is placed, there must be something in the `Backend` service that:
* Getting customer's detail
* Getting merchant's detail
* Getting payment's detail
* Getting promotion's detail
* Getting each product / item detail
* Calculate `sum` of the order and deduct it with promotion
* Make an `order` based on the gathere information, and notify it to `merchant`
* Make an `invoice` based on the gathered information, and notify it to `customer`
* Make a payment through `payment` service based on the `invoice`
* Flag the invoice `state` as `paid`
* Notify merchant that `payment` is done to said `order`
* Notify customer that `payment` success to said `invoice`

That's **Tremendously Ridiculous** amount of work needed to be done in the `Backend` side.
It sounds `DUMB` to just make end `API` endpoint to handle these work, especially if your `Platform` can receive a ludicrous amount of order per minute across multiple region.

## Aggregate Root using Actor

First of all, let's take on the rate of `order` issue.
`Order` can be placed at any time by millions of `customer` at the same time and the system must NEVER lose any of it.
System can defer processing an order by placing it in a `queue` which can be persisted and recovered in case of server restart, crash, or whatever disaster that comes.

Let's design an object called `Actor`

* An `Actor` is an object which `process` messages being sent to it
* So each `Actor` have an inbox which acts as a queue
* The inbox can contain any type of message, but usually 1 actor handle 1 kind of message
* The actor `processor` is a function delegated by its caller
* Each actor can only have 1 type of `processor` but can have multiple instance of it. Let's call it `worker`.
* A `processor` might, or might not produce a result
* A `processor` might, or might not produce an error

```go
package actor

// Processor is the delegate which process a message
// @worker is its assigned worker number (starts from 1) in case we make more than 1 worker
// @actor is the reference to which actor that receives the message
// @message is the current individual message from actor's inbox
type Processor func(worker int, actor *Actor, message interface{}) (interface{}, error)

// Exception handler in case processor produce an error
// @worker is its assigned worker number (starts from 1) in case we make more than 1 worker
// @actor is the reference to which actor that receives the message
// @err is the error that happened after trying to process a message
type Exception func(worker int, actor *Actor, err error)

// Options when initializeing an Actor
type Options struct {
    Worker      int              // number of worker / processor go routine, defaults = 1
    Output      chan interface{} // output channel, on which Actor will sned in after process is done
}

// configure fallbacks to default
func (opt *Options) configure() {
    if opt.Worker <= 0 {
        opt.Worker = 1
    }
}

// Actor ...
type Actor struct {
    inbox     chan interface{}
    outbox    chan interface{}
    process   Processor
    exception Exception
}


// New instance of an Actor with w as number of worker
// @p is the processor function
// @e is the exception handler
func New(p Processor, e Exception, opt *Options) *Actor {
    opt.configure()

    actor := &Actor{
        inbox:     make(chan interface{}, opt.Worker),
        outbox:    opt.Output,
        process:   p,
        exception: e,
    }

    actor.start(0, opt.Worker)
    return actor
}

// start the actor with n number of worker
func (actor *Actor) start(idx, n int) {
    if idx == n {
        return
    }

    // worker number starts from 1
    go actor.work(idx + 1)
    actor.start(idx+1, n)
}
func (actor *Actor) work(w int) {
    for message := range actor.inbox {
        result, err := actor.process(w, actor, message)

        if err != nil && actor.exception != nil {
            actor.exception(w, actor, err)
            continue
        }

        if actor.outbox != nil {
            actor.outbox <- result
            continue
        }
    }
}

// Queue a message to inbox
func (actor *Actor) Queue(messages ...interface{}) {
    for _, message := range messages {
        actor.inbox <- message
    }
}
```

We'll test the actor with simple test case:

```go
func Test_Actor(t *testing.T) {
    word= [...]string{"One", "Two", "Three"}
    actor := New(func(w int, actor *Actor, message interface{}) (interface{}, error) {
        result := words[w-1]

        fmt.Println("worker", w,
            "receive", message,
            "processed as", result,
            "send to?", actor.outbox)

        return result, nil
    }, func(w int, actor *Actor, err error) {
        fmt.Println(err)
    }, &Options{Worker: 3})

    wg := sync.WaitGroup{}
    for i := 0; i <= 10; i++ {
        wg.Add(1)
        go func(i int) {
            idx := i % 3
            word := words[idx]
            actor.Queue(word)
            wg.Done()
        }(i)
    }

    wg.Wait()
}
```

* The test code creates an `actor` with 3 `worker`
* Each `worker` do the same type of job, which only returns word based on its worker number
* We loop 10 times and `queue` a word into the `actor`
* Lastly, we wait for the `wait group` to be done

```bash
robin.bastian$ go test -timeout=10s -run "^(Test_Actor)$"
worker 2 receive One processed as Two send to? <nil>
worker 2 receive One processed as Two send to? <nil>
worker 3 receive Two processed as Three send to? <nil>
worker 3 receive One processed as Three send to? <nil>
worker 1 receive Two processed as One send to? <nil>
worker 1 receive Three processed as One send to? <nil>
worker 1 receive Three processed as One send to? <nil>
worker 1 receive Two processed as One send to? <nil>
worker 1 receive One processed as One send to? <nil>
worker 2 receive Three processed as Two send to? <nil>
worker 3 receive Two processed as Three send to? <nil>
PASS
ok      github.com/bastianrob/go-experiences/generator/actor    0.006s
```

From the test results, we can see:

* Exactly 3 workers are spawned with one type of processor
* Order of message processed is **NOT** guaranteed
* Each process returns a `result` but is not being sent to `actor's` outbox

## Stopping the Actor

Next, we'll need a mechanism to `Stop` an `actor`. When an actor is stopped, we collect all pending messages and return it to the caller.

First of all, we add an exit mechanism to the `Actor` by adding:

```go
type Actor struct {
    ...
    // exit mechanism
    exit       chan struct{}
    workgroup  *sync.WaitGroup // worker waitgroup
    inboxgroup *sync.WaitGroup // inbox waitgroup
}

func New(p Processor, e Exception, opt *Options) *Actor {
    opt.configure()

    actor := &Actor{
        ...
        exit:       make(chan struct{}),
        workgroup:  &sync.WaitGroup{},
        inboxgroup: &sync.WaitGroup{},
    }

    actor.start(0, opt.Worker)
    return actor
}
```

* `exit` is an exit channel which will be used to signal each `worker` to stop processing a message
* `workgroup` is a wait group that counts how many worker is still active
* `inboxgroup` is a wait group that counts how many messages is still in the inbox

Next, we implement all of those `exit` mechanism:

* `inboxgroup` is added everytime a message is queued into `inbox`
* `workgroup` is added everytime a `worker` is spawned
* Each of it tries to listen to both `inbox` and `exit` channel by using `select case` statement
* `inboxgroup` is decreased everytime a message is done processed by a `worker`
* `workgroup` is decreased everytime a worker quits when it receives signal from `exit` channel

```go
func (actor *Actor) Queue(messages ...interface{}) {
    // add length of message to inbox wait group
    actor.inboxgroup.Add(len(messages))
    for _, message := range messages {
        actor.inbox <- message
    }
}

func (actor *Actor) work(w int) {
    actor.workgroup.Add(1)       // worker group is added
    defer actor.workgroup.Done() // defer flagging worker group as done

    for {
        select {
        case message := <-actor.inbox: // waits for message to come from inbox
            result, err := actor.process(w, actor, message)

            if err != nil && actor.exception != nil {
                actor.exception(w, actor, err)
                actor.inboxgroup.Done() // flag 1 message as done
                continue
            }

            if actor.outbox != nil {
                actor.outbox <- result
                actor.inboxgroup.Done() // flag 1 message as done
                continue
            }

            // flag 1 message as done
            actor.inboxgroup.Done()
        case <-actor.exit: // listen on exit signal
            return
        }
    }
}
```

Lastly, we implements the `Stop` method to:

* Close the `exit` channel so it is broadcasted to all `workers`
* Wait for all `workers` to finish exiting
* Gather the pending messages that's still lingering in `inbox`
* Decrease the `inboxgroup` counter everytime we finished collecting a message
* Wait for all pending messages to be collected
* Close the `inbox` channel
* And report all the pending messages to caller

```go
// Stop actor from processing any message
func (actor *Actor) Stop() (pendings []interface{}) {
    // stop all worker from processing any inbox
    close(actor.exit)
    actor.workgroup.Wait()

    // gather pending messages inside inbox and flag it as done
    go func() {
        for message := range actor.inbox {
            pendings = append(pendings, message)
            actor.inboxgroup.Done()
        }
    }()

    // wait for pending messages gathering to be completed and close the inbox channel
    actor.inboxgroup.Wait()
    close(actor.inbox)

    // return gathered pending messages
    return pendings
}
```

And now it's time to test the stop mechanism:

```go
func Test_ActorStop(t *testing.T) {
    mux := sync.Mutex{}
    var processed []interface{}
    actor := New(func(w int, actor *Actor, message interface{}) (interface{}, error) {
        mux.Lock()
        processed = append(processed, message)
        mux.Unlock()

        return nil, nil
    }, func(w int, actor *Actor, err error) {
        fmt.Println(err)
    }, &Options{Worker: 5})

    expected := 0
    for i := 1; i <= 100; i++ {
        go actor.Queue(i)
        expected = expected + i
    }

    pendings := actor.Stop()
    combined := append(processed, pendings...)

    sum := 0
    for _, e := range combined {
        sum = sum + e.(int)
    }
    if sum != expected {
        t.Error("Sum of 1-100 must be", expected, "but received", sum)
    }

    fmt.Println("PEND", pendings)
    fmt.Println("PROC", processed)
}
```

* The test code creates an `Actor` with 5 `workers`
* We feed the `actor` with number from 1 to 100
* So the `expected` sum of `1 + 2 + ... + 100` should be `5050`
* In the `processor` function, we collect all processed number into a variable called `var processed []interface{}`
* We use `mutex` to guard it because it will be accessed by 5 `workers` working from different go routine
* We then collect all pending numbers by calling `actor.Stop` and store it to a variable called `pendings`
* We combine both `processed` and `pendings` into slice called `combined` and `sum` all numbers inside it
* We compare the `sum` of all combined numbers with the `expected` number

```bash
robin.bastian$ go test -timeout=10s -run "^(Test_ActorStop)$"
PEND [7 21 8 22 10 23 63 24 57 25 26 27 28 29 30 31 73 64 65 66 67 68 58 59 60 61 62 71 70 79 74 75 76 77 69 78 72 82 80 81 84 83 85 86 93 87 88 89 90 91 92 96 94 95 98 97 99 100]
PROC [6 11 3 4 5 43 13 33 14 34 15 35 16 36 17 18 37 38 39 40 41 42 56 44 45 46 47 48 49 50 51 52 53 54 55 19 9 12 1 32 20 2]
PASS
ok      github.com/bastianrob/go-experiences/generator/actor    0.006s
```

## Directing Actors

Individually, an `actor` can do its job fine enough. But it will be better if we can chain `actors` together to build a pipeline processing!
So let's think of what we already have:

* Each actor have an `inbox` and `outbox`
* While inbox is `required` an `outbox` is optional
* Now we want to chain output from one `actor` as input to another `actor`
* In other words: we can assign one `actor's` outbox with another `actor's` inbox

```go
package actor

// Direct inbox of a target actor, as source actor's outbox
func Direct(actors ...*Actor) {
    var source *Actor
    for _, target := range actors {
        if source == nil {
            source = target
            continue
        }

        source.outbox = target.inbox
        source = target
    }
}
```

Again, we'll write a test to ensure the behaviour is within expectation

```go
func Test_ActorDirected(t *testing.T) {
    errPrinter := func(w int, actor *Actor, err error) {
        fmt.Println("worker", w, "-", "err:", err)
    }

    opt := &Options{
        Worker: 3,
    }

    bale := New(func(w int, actor *Actor, in interface{}) (interface{}, error) {
        return in, nil
    }, errPrinter, opt)
    bane := New(func(w int, actor *Actor, in interface{}) (interface{}, error) {
        switch {
        case in == "I AM VENGEANCE":
            return "I AM INEVITABLE", nil
        case in == "I AM THE NIGHT":
            return "I AM BANE", nil
        case in == "I'M BATMAN":
            return "I WILL BREAK YOU", nil
        default:
            return nil, errors.New("WHATEVER YOU SAY")
        }
    }, errPrinter, opt)
    printer := New(func(w int, actor *Actor, in interface{}) (interface{}, error) {
        return nil, nil
    }, errPrinter, opt)

    bale.name = "Bale"
    bane.name = "Bane"
    printer.name = "Printer"

    Direct(bale, bane, printer)
    bale.Queue("I AM VENGEANCE", "I AM THE NIGHT", "I'M BATMAN", "HEY HO!")
}
```

Result shows:

```bash
panic: sync: negative WaitGroup counter
```

Uh oh, turns out we can't