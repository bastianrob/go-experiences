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

    go func(w int) {
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
    }(idx + 1) // worker number starts from 1

    actor.start(idx+1, n)
}

// Queue a message to inbox
func (actor *Actor) Queue(messages ...interface{}) {
    for _, message := range messages {
        actor.inbox <- message
    }
}
```

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

