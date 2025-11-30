// Package main implements a simple webserver with an embedded Tailscale client.
// This allows the server to be accessible via a Tailscale network (tailnet).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tailscale.com/tsnet"
)

func main() {
	hostname := flag.String("hostname", "hello-tailscale", "Tailscale hostname for this service")
	port := flag.Int("port", 80, "Port to listen on")
	stateDir := flag.String("state-dir", "", "Directory to store Tailscale state (defaults to OS-specific location)")
	flag.Parse()

	// Create a new tsnet server
	srv := &tsnet.Server{
		Hostname: *hostname,
	}

	if *stateDir != "" {
		srv.Dir = *stateDir
	}

	// Start the Tailscale server
	log.Printf("Starting Tailscale server with hostname: %s", *hostname)
	ln, err := srv.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to start Tailscale listener: %v", err)
	}
	defer ln.Close()

	// Get local client to retrieve the Tailscale IP
	lc, err := srv.LocalClient()
	if err != nil {
		log.Fatalf("Failed to get local client: %v", err)
	}

	// Wait for Tailscale to be ready and get the IP
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	status, err := lc.StatusWithoutPeers(ctx)
	if err != nil {
		log.Printf("Warning: Could not get Tailscale status: %v", err)
	} else {
		log.Printf("Tailscale IP: %v", status.TailscaleIPs)
	}

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloHandler)
	mux.HandleFunc("/health", healthHandler)

	// Create HTTP server
	httpServer := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start the HTTP server in a goroutine
	go func() {
		log.Printf("HTTP server listening on Tailscale network at port %d", *port)
		if err := httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	if err := srv.Close(); err != nil {
		log.Printf("Tailscale server close error: %v", err)
	}

	log.Println("Server shutdown complete")
}

// helloHandler serves the "Hello World" page
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Hello Tailscale</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }
        .container {
            text-align: center;
            padding: 2rem;
        }
        h1 {
            font-size: 3rem;
            margin-bottom: 1rem;
        }
        p {
            font-size: 1.2rem;
            opacity: 0.9;
        }
        .tailscale-badge {
            margin-top: 2rem;
            padding: 0.5rem 1rem;
            background: rgba(255,255,255,0.2);
            border-radius: 20px;
            font-size: 0.9rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🌐 Hello, World!</h1>
        <p>Welcome to the Tailscale-powered webserver</p>
        <div class="tailscale-badge">Connected via Tailscale</div>
    </div>
</body>
</html>`)
}

// healthHandler provides a simple health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
}
