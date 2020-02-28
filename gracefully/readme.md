# Shutdown Horror Story

* Imagine you have a `backend service` which process a `queue` of `operations`.
* Around 20 message comes to this `queue` every second.
* Each `operation` takes around 3 seconds to complete.
* And have 20 parallel `workers` to process each message in the `queue`.

> And then there is light

Or no light!

Your engineering team want to deploy a `hotfix` to a critical bug in the `worker` processing code and they need it now.
Otherwise, the data processed by currenly live `worker` might produce wrong result.

You begin to sweat, usually deployment starts at dawn when traffic is low / no traffic at all.
You haven't got time, resource, and possibly proficiency, to properly implements blue-green nor canary delpoyment.
The currently running instance, is all you have and it needs to be restarted, introducing downtime, and possibly corrupting/losing data that is being processed right now in the `worker`.

Your `CD` pipeline, `Mr. Travis` (or jenkins, actions, circle, whatever) starts his workflow.

* He build the `backend service` for you, push it to your VM instance, and then, the time comes...
* He sends `SIGTERM` signal to your `backend service`.
* He waits, and waits for you `backend service` to stop whatever it is doing, so he can close shops.
* But the signal never came...

`Mr. Travis` have to forcefully kill your `backend service` by sending `SIGKILL` signal so `OS'` terminator will hunt, and kill your service without dignity.

* Your `backend service` died while still handling `N amount` of messages
* Your engineering team must then do an autopsy to the killed service by examining `log`.
* Hopeful that they can find whatever messagse haven't done processing, so they can re-queue them manually.

## Gracefully Shutdown your Service

Welcome to the light, young padawan.
Here, you will learn how to gracefully execute order 66...

### The Concept

So, when OS (unix system), process, or people wants to shut your application down, they can send various types of `signal`, which then can be interrupted by our `app` to prepare shutthing down process.

There are lots of [`signal` types](https://golang.org/pkg/os/signal/#hdr-Types_of_signals), but we'll only focus on 2 types of `signal`

* `SIGINT` or Signal Interrupt. Typically sent when a user press `CTRL+C` to exit app
* `SIGTERM` or Signal Terminate. Typically sent by an app to kill another app. Most likely from administrative tools.

Both `signal` can be used to politely ask app to terminate their process, cleaning up any hanging operations.
`OS` will wait for 30s for app to shutdown, otherwise, it will send `SIGKILL`. Which is another type of `signal` that can't be intercepted and will forcefully shutdown app.

### Basic Implementation in `Go`

Using go channel we can make our program wait for a signal

```go
wait := make(chan bool)
before := time.Now()
//this will be executed asynchronously
go func() {
    time.Sleep(3 * time.Second)
    wait <- false
}()

<-wait
fmt.Println("I am done")
fmt.Println(int(time.Since(before).Seconds()), "sec")
```

```log
bastianrob$ go run main.go
I am done
3 sec
```

And then in go, we can listen to `OS' signal` by using

```go
sig := make(chan os.Signal)
signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
```

Notice that `sig` is a channel that waits from `SIGINT` and `SIGTERM` signal.
We'll use this to setup our `teardown` process

```go
package gracefully

// Serve HTTP gracefuly
func Serve(listenAndServe func() error, teardown func(context.Context) error) error {
    term := make(chan os.Signal) // OS termination signal
    fail := make(chan error)     // Teardown failure signal

    go func() {
        signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)
        <-term // waits for termination signal

        // context with 30s timeout
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        // all teardown process must complete within 30 seconds
        fail <- teardown(ctx)
    }()

    // listenAndServe blocks our code from exit, but will produce ErrServerClosed when stopped
    if err := listenAndServe(); err != nil && err != http.ErrServerClosed {
        return err
    }

    // after server gracefully stopped, code proceeds here and waits for any error produced by teardown() process @ line 26
    return <-fail
}
```

And in the `main.go`:

```go
func main() {
    server := &http.Server{
        Addr: ":8080",
        // Handler: your API handler
    }

    if err := gracefully.Serve(server.ListenAndServe, func(ctx context.Context) error {
        if err := server.Shutdown(ctx); err != nil {
            return err
        }

        // unplug from message broker
        // unplug from service mesh
        // remove temporary files
        // wait for all pending queue/topic processor to finish
        // etc, yada-yada

        return nil
    }); err != nil {
        log.Fatalln("ERR:", err)
    }
}
```
