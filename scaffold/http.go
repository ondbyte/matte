package main

import (
	"context"
	"errors"
	"net"
	"net/http"
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
