// Command server is the Shush backend entry point. It serves the ConnectRPC
// handler — gRPC, gRPC-Web, and HTTP/JSON from a single endpoint — over
// net/http on :8080. The PostgreSQL store, auth interceptors, and TLS will be
// added in later slices.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"connectrpc.com/grpcreflect"
	"github.com/Aravind-pk/shush/backend/gen/shush/v1/shushv1connect"
	"github.com/Aravind-pk/shush/backend/internal/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	mux := http.NewServeMux()

	path, handler := shushv1connect.NewShushServiceHandler(server.New())
	mux.Handle(path, handler)

	reflector := grpcreflect.NewStaticReflector(shushv1connect.ShushServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	// Liveness probe — a hand-written route on the same mux, showing how
	// non-RPC HTTP endpoints coexist with the Connect handler.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	// h2c lets us serve HTTP/2 without TLS, which gRPC clients need over plain
	// http://. In production this sits behind a TLS-terminating proxy or gets
	// real certs.
	srv := &http.Server{
		Addr:    *addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Graceful shutdown: on SIGINT/SIGTERM, stop accepting new requests and let
	// in-flight ones finish before exiting.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("shutting down…")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("shush Connect server listening on %s", *addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("serve: %v", err)
	}
}
