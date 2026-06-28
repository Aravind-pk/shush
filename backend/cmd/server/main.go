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

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/Aravind-pk/shush/backend/gen/shush/v1/shushv1connect"
	"github.com/Aravind-pk/shush/backend/internal/auth"
	"github.com/Aravind-pk/shush/backend/internal/server"
	"github.com/Aravind-pk/shush/backend/internal/store"
	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	// Load .env if present (dev convenience). Each path is tried independently
	// so it works whether run from the repo root or from backend/ (via air).
	// godotenv never overrides vars already set in the real environment.
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}
	kek := []byte(os.Getenv("MASTER_KEK"))

	clerkKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkKey == "" {
		log.Fatal("CLERK_SECRET_KEY is required (verifies Clerk session tokens)")
	}
	// Configure the Clerk SDK globally so the auth interceptor's jwt.Verify can
	// fetch and cache Clerk's JWKS to validate session tokens.
	clerk.SetKey(clerkKey)

	// Open the Postgres store (pgx pool). Use a bounded context so a dead DB
	// fails fast at boot instead of hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	st, err := store.New(ctx, dsn, kek)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	mux := http.NewServeMux()

	// Every RPC is guarded by the Clerk auth interceptor.
	path, handler := shushv1connect.NewShushServiceHandler(
		server.New(st),
		connect.WithInterceptors(auth.NewInterceptor()),
	)
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
