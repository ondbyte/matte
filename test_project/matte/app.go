package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var _ = "importsPlaceHolder"

type ShutDowner func() error
type ServerRunner func() (chan error, ShutDowner)

var (
	DefaultPort = ":8000"

	// all the servers
	ServerRunners = []ServerRunner{}
	// all the server shut downers
	ShutDowners = []ShutDowner{}
)

func RunHttpServer() (chan error, ShutDowner) {
	ctx, cancel := context.WithCancel(context.Background())
	handler := http.NewServeMux()

	var _ = "handlersPlaceHolder"

	addr := DefaultPort
	var _ = "portPlaceHolder"

	server := http.Server{
		Addr:    addr,
		Handler: handler,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	errChan := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
		cancel()
	}()
	return errChan, func() error {
		return server.Shutdown(ctx)
	}
}

func osSignal() chan os.Signal {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM, os.Kill, syscall.SIGINT)
	return signalChannel
}

func main() {
	// place holder should be replaced with `Runners = append(Runners, ...)`
	var _ = "serverRunnersPlaceHolder"

	errChan := make(chan error)
	for _, serverRunner := range ServerRunners {
		serverErr, shutDowner := serverRunner()
		ShutDowners = append(ShutDowners, shutDowner)
		go func() {
			errChan <- <-serverErr
		}()
	}
	shutDown := func() error {
		errS := ""
		for _, shutDown := range ShutDowners {
			err := shutDown()
			if err != nil {
				errS += errS + "\n"
			}
		}
		if errS != "" {
			return fmt.Errorf(errS)
		}
		return nil
	}
	select {
	case err := <-errChan:
		{
			if err != nil {
				panic(err)
			}
		}
	case sig := <-osSignal():
		{
			fmt.Printf("signal: %v", sig)
			fmt.Println("shutting down")
			shutDown()
		}
	}
}
