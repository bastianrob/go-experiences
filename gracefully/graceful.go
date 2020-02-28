package gracefully

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

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
