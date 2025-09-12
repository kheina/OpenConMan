package srv

import (
	"context"
	"crypto/tls"
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
	"google.golang.org/grpc/credentials"

	"github.com/kheina/openconman/src/srv/api"
	"github.com/kheina/openconman/src/srv/certs"
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

	insecure bool
	cert     []byte
	key      []byte

	// cancelling the Serve context triggers a graceful shutdown
	SrvCancel context.CancelFunc
	// cancelling the Kill context triggers a forced shutdown
	KillCancel context.CancelFunc

	grpcServer   *grpc.Server
	grpcListener net.Listener
	httpServer   *http.Server
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

// MakeWaitGroupCh creates a channel that returns a message on waitgroup completion
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

// Serve runs the server and serves the conman server with the provided details
// from NewRouter
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

	var err error
	if r.grpcListener, err = net.Listen("tcp", r.gaddr); err != nil {
		return fmt.Errorf("%s: failed to init grpc listener: %w", op, err)
	}
	if r.httpListener, err = net.Listen("tcp", r.addr); err != nil {
		return fmt.Errorf("%s: failed to listen on address: %w", op, err)
	}

	var conf *tls.Config
	grpco := []grpc.ServerOption{}
	if !r.insecure {
		switch {
		case r.cert == nil && r.key != nil:
			fallthrough
		case r.cert != nil && r.key == nil:
			return fmt.Errorf("%s: tls cert or key provided, but not both", op)
		case r.cert != nil && r.key != nil:
			// valid, load the certs below
		default:
			r.cert, r.key, err = certs.GenerateCertificate()
			if err != nil {
				return fmt.Errorf("%s: failed to generate tls cert: %w", op, err)
			}
			r.logger.Debug("TLS certificate generated")
		}

		cert, err := tls.X509KeyPair(r.cert, r.key)
		if err != nil {
			return fmt.Errorf("%s: failed to parse tls cert and key: %w", op, err)
		}
		r.logger.Debug("TLS certificate loaded")

		conf = &tls.Config{
			ServerName:   "conman", // this may need to be updated to a dns name or something
			NextProtos:   []string{"h2", "http/1.1"},
			ClientAuth:   tls.RequestClientCert,
			Certificates: []tls.Certificate{cert},
		}

		// don't convert the grpc listener to tls, as it's handled internally by grpc.Server
		r.httpListener = tls.NewListener(r.httpListener, conf)
		grpco = append(grpco, grpc.Creds(credentials.NewTLS(conf)))
	}

	ctx := r.srvCtx
	r.grpcServer = grpc.NewServer(grpco...)
	h, err := api.Handler(ctx, r.grpcServer, r.gaddr, r.logger, conf)
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
		TLSConfig:         conf,
		ErrorLog:          r.logger.StandardLogger(&hclog.StandardLoggerOptions{ForceLevel: hclog.Debug}),
	}

	return r.serve()
}

func (r Router) serve() error {
	const op = "srv.(Router).serve"
	ctx, kctx := r.srvCtx, r.killCtx
	wg := new(sync.WaitGroup)
	wg.Add(2) // number of listeners, grpc + http

	// start the grpc server for the api
	r.logger.Info(fmt.Sprintf("%s: serving gRPC on %s", op, r.gaddr))
	go func() {
		defer wg.Done()
		if err := r.grpcServer.Serve(r.grpcListener); err != nil {
			r.logger.Info(fmt.Sprintf("%s: gRPC server shutdown: %s", op, err.Error()))
		}
	}()

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

func NewRouter(host string, port, grpcPort uint, workDir string, logLevel hclog.Level, cert, key []byte, insecure bool) (*Router, error) {
	const op = "srv.NewRouter"
	var logLock sync.Mutex
	logger := hclog.New(&hclog.LoggerOptions{
		Output: os.Stdout,
		Level:  logLevel,
		// JSONFormat: true,
		Mutex: &logLock,
	})
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			return nil, fmt.Errorf("%s: received error while changing working dir: %w", op, err)
		}
	}
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
		cert:       cert,
		key:        key,
		insecure:   insecure,
	}, nil
}
