package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	logger "github.com/grackleclub/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

const serviceName = "abolish"

var (
	log     *slog.Logger
	version = buildVersion() // VCS commit (or dev) from build info
	tag     = dev            // release tag, injected via -ldflags "-X main.tag=<tag>"
	port    = "8686"
)

//go:embed index.html
var index []byte

func main() {
	ctx := context.Background()
	if p := os.Getenv("ABOLISH_PORT"); p != "" {
		port = p
	}

	var err error
	log, err = logger.New(slog.HandlerOptions{})
	if err != nil {
		panic(fmt.Sprintf("create slog handler: %v", err))
	}
	log = log.With(
		"service.version", version,
		"service.tag", tag,
		"service.name", serviceName,
	)

	shutdown, err := tracing(ctx)
	if err != nil {
		panic(fmt.Sprintf("setup tracing: %v", err))
	}
	defer shutdown(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)

	log.Info("starting server", "port", port)
	handler := otelhttp.NewHandler(logMW(rateMW(mux)), "server")
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Error("server failed", "error", err)
		panic(fmt.Sprintf("server failed: %v", err))
	}
}

// rootHandler serves the embedded index page.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _, err := w.Write(index); err != nil {
		log.Error("write index", "error", err)
	}
}
