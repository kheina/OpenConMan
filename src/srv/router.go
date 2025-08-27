package srv

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"

	"github.com/kheina/openconman/src/srv/api"
	"github.com/kheina/openconman/src/srv/middleware/cors"
	"github.com/kheina/openconman/src/srv/ui"
)

type Router struct {
	addr       string
	gaddr      string
	srvCtx     context.Context
	killCtx    context.Context
	shutdownCh chan struct{}
	logger     hclog.Logger

	// cancelling the Serve context triggers a graceful shutdown
	SrvCancel context.CancelFunc
	// cancelling the Kill contexted triggers a forced shutdown
	KillCancel context.CancelFunc

	listeners    []net.Listener
	grpcServer   *grpc.Server
	httpServer   *http.Server
	grpcListener net.Listener
	httpListener net.Listener
}

// MakeShutdownCh returns a channel that can be used for shutdown
// notifications for commands. This channel will send a message for every
// SIGINT or SIGTERM received.
func MakeShutdownCh() chan struct{} {
	resultCh := make(chan struct{})

	shutdownCh := make(chan os.Signal, 4)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		for {
			<-shutdownCh
			resultCh <- struct{}{}
		}
	}()
	return resultCh
}

func MakeWaitGroupCh(wg interface{ Wait() }) chan struct{} {
	ch := make(chan struct{})
	go func() {
		wg.Wait()
		ch <- struct{}{}
	}()
	return ch
}

// GracefulShutdown attempts to gracefully shutdown servers and listeners by
// declining new connections and resolving active connections. blocks until
// complete.
func (r *Router) GracefulShutdown() {
	r.SrvCancel()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		r.grpcServer.GracefulStop()
	}()
	go func() {
		defer wg.Done()
		_ = r.httpServer.Shutdown(context.Background())
	}()
	wg.Wait()
}

// ForceShutdown blocks until all listeners are servers are stopped. it should
// close all connections immediately.
func (r *Router) ForceShutdown() error {
	const op = "srv.(Router).ForceShutdown"
	r.KillCancel()
	var wg sync.WaitGroup
	wg.Add(4)
	// grpc stop is a blocking connection, so do it in a goroutine
	go func() {
		defer wg.Done()
		_ = r.httpListener.Close()
	}()
	go func() {
		defer wg.Done()
		_ = r.grpcListener.Close()
	}()
	go func() {
		defer wg.Done()
		r.grpcServer.Stop()
	}()
	go func() {
		defer wg.Done()
		_ = r.httpServer.Close()
	}()
	wg.Wait()
	return nil
}

func (r *Router) Serve() error {
	const op = "srv.(Router).Serve"
	switch {
	case r.addr == "":
		return fmt.Errorf("%s: router missing host address", op)
	case r.srvCtx == nil:
		return fmt.Errorf("%s: router missing serve context", op)
	case r.SrvCancel == nil:
		return fmt.Errorf("%s: router missing serve cancel", op)
	}

	ctx, kctx := r.srvCtx, r.killCtx
	r.grpcServer = grpc.NewServer()
	if err := api.RegisterGrpcServices(ctx, r.grpcServer, r.logger); err != nil {
		return fmt.Errorf("%s: failed to register gRPC services: %w", op, err)
	}
	wg := new(sync.WaitGroup)
	wg.Add(2) // number of listeners, grpc + http

	// start the grpc server for the api
	r.logger.Info(fmt.Sprintf("%s: serving gRPC on %s", op, r.gaddr))
	var err error
	r.grpcListener, err = net.Listen("tcp", r.gaddr)
	go func() {
		defer wg.Done()
		if err := r.grpcServer.Serve(r.grpcListener); err != nil {
			r.logger.Info(fmt.Sprintf("%s: gRPC server shutdown: %s", op, err.Error()))
		}
	}()

	h, err := api.Handler(ctx, r.gaddr, r.logger)
	if err != nil {
		return fmt.Errorf("%s: failed to retrieve api handler: %w", op, err)
	}

	// api handler needs to be wrapped in cors middleware
	cors := cors.New(r.logger)
	h = cors.WrapHandler(h)
	mux := http.NewServeMux()

	// all api handlers should be added on a versioned path
	mux.Handle("/v1/", h)

	// now fetch the ui handler
	if h, err = ui.Handler(); err != nil {
		return fmt.Errorf("%s: failed to retrieve ui handler: %w", op, err)
	}
	// use root pattern to match all non-api requests
	mux.Handle("/", h)

	r.httpServer = &http.Server{
		Handler:           cleanhttp.PrintablePathCheckHandler(mux, nil),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		IdleTimeout:       30 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	r.httpListener, err = net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("%s: failed to listen on address: %w", op, err)
	}

	r.logger.Info(fmt.Sprintf("%s: listening on %s", op, r.addr))
	go func() {
		defer wg.Done()
		if err := r.httpServer.Serve(r.httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			r.logger.Warn(fmt.Sprintf("%s: server shutdown: %s", op, err.Error()))
		}
	}()

serve:
	for {
		select {
		case <-ctx.Done():
			r.logger.Info(fmt.Sprintf("%s: shutdown signal received", op))
			break serve
		case <-r.shutdownCh:
			r.logger.Info(fmt.Sprintf("%s: interrupt signal received", op))
			break serve
		}
	}

	go r.GracefulShutdown()
	// go r.grpcServer.Stop()
	// _ = r.httpServer.Close()
	// _ = r.httpListener.Close()
	// _ = r.grpcListener.Close()
	wgCh := MakeWaitGroupCh(wg)

shutdown:
	for {
		select {
		case <-wgCh:
			r.logger.Info(fmt.Sprintf("%s: graceful shutdown finished", op))
			return nil
		case <-kctx.Done():
			r.logger.Error(fmt.Sprintf("%s: kill signal received", op))
			go r.ForceShutdown()
			break shutdown
		case <-r.shutdownCh:
			r.logger.Error(fmt.Sprintf("%s: second interrupt received", op))
			go r.ForceShutdown()
			break shutdown
		}
	}

	for {
		select {
		case <-wgCh:
			r.logger.Info(fmt.Sprintf("%s: force shutdown finished", op))
			return nil
		case <-r.shutdownCh:
			r.logger.Error(fmt.Sprintf("%s: third interrupt received - force exiting", op))
			os.Exit(130)
			return nil
		}
	}
}

func NewRouter(host string, port, grpcPort uint, logLevel hclog.Level) (*Router, error) {
	var logLock sync.Mutex
	logger := hclog.New(&hclog.LoggerOptions{
		Output: os.Stdout,
		Level:  logLevel,
		// JSONFormat: true,
		Mutex: &logLock,
	})

	kctx, kcancel := context.WithCancel(context.Background())
	ctx, cancel := context.WithCancel(kctx)
	return &Router{
		addr:       net.JoinHostPort(host, strconv.FormatUint(uint64(port), 10)),
		gaddr:      net.JoinHostPort(host, strconv.FormatUint(uint64(grpcPort), 10)),
		srvCtx:     ctx,
		SrvCancel:  cancel,
		killCtx:    kctx,
		KillCancel: kcancel,
		shutdownCh: MakeShutdownCh(),
		logger:     logger,
	}, nil
}
