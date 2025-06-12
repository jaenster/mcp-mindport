package daemon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"syscall"

	"mcp-mindport/internal/config"
	"mcp-mindport/internal/mcp"

	"github.com/gorilla/websocket"
)

type Daemon struct {
	server   *mcp.Server
	config   *config.Config
	upgrader websocket.Upgrader
}

func NewDaemon(server *mcp.Server, config *config.Config) *Daemon {
	return &Daemon{
		server: server,
		config: config,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now - in production, you'd want to be more restrictive
				return true
			},
		},
	}
}

func Start(ctx context.Context, server *mcp.Server, config *config.Config) error {
	daemon := NewDaemon(server, config)
	
	// Write PID file
	if err := writePidFile(config.Daemon.PidFile); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer removePidFile(config.Daemon.PidFile)

	// Setup logging to file
	if config.Daemon.LogFile != "" {
		logFile, err := os.OpenFile(config.Daemon.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	log.Printf("Starting MCP daemon on %s:%d", config.Server.Host, config.Server.Port)

	// Setup HTTP server with WebSocket support
	http.HandleFunc("/mcp", daemon.handleWebSocket)
	http.HandleFunc("/health", daemon.handleHealth)
	http.HandleFunc("/", daemon.handleRoot)

	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port),
	}

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- httpServer.ListenAndServe()
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		log.Println("Daemon shutting down...")
		return httpServer.Shutdown(context.Background())
	case err := <-errChan:
		return fmt.Errorf("daemon server error: %w", err)
	}
}

func (d *Daemon) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := d.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("New WebSocket connection from %s", r.RemoteAddr)

	// Handle MCP protocol over WebSocket
	// This is a simplified implementation - in production you'd want proper error handling
	for {
		var request mcp.MCPRequest
		if err := conn.ReadJSON(&request); err != nil {
			log.Printf("Failed to read WebSocket message: %v", err)
			break
		}

		// Process the MCP request (this would need to be adapted from the stdio version)
		response := d.processRequest(r.Context(), &request)
		
		if err := conn.WriteJSON(response); err != nil {
			log.Printf("Failed to write WebSocket message: %v", err)
			break
		}
	}
}

func (d *Daemon) processRequest(ctx context.Context, request *mcp.MCPRequest) interface{} {
	// This is a simplified version - you'd need to adapt the MCP server's handleRequest method
	// For now, just return a basic response
	return map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      request.ID,
		"result":  map[string]interface{}{"status": "ok"},
	}
}

func (d *Daemon) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok","service":"mcp-mindport"}`))
}

func (d *Daemon) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <title>MindPort</title>
</head>
<body>
    <h1>MindPort Resource Server</h1>
    <p>Model Context Protocol server with optimized search capabilities.</p>
    <ul>
        <li><a href="/health">Health Check</a></li>
        <li>WebSocket endpoint: <code>/mcp</code></li>
    </ul>
    <p>Use this server with MCP-compatible AI tools for resource storage and retrieval.</p>
</body>
</html>
	`))
}

func writePidFile(pidFile string) error {
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func removePidFile(pidFile string) {
	os.Remove(pidFile)
}

// Signal handling for graceful shutdown
func handleSignals(ctx context.Context, cancel context.CancelFunc) {
	// This would be called from main to handle SIGTERM, SIGINT, etc.
	// Implementation depends on your specific requirements
}

// Daemonize process (Unix-specific)
func Daemonize() error {
	// Fork the process
	if os.Getppid() == 1 {
		// Already a daemon
		return nil
	}

	// Fork and exit parent
	pid, err := syscall.ForkExec(os.Args[0], os.Args, &syscall.ProcAttr{
		Dir:   "/",
		Env:   os.Environ(),
		Files: []uintptr{0, 1, 2}, // stdin, stdout, stderr
		Sys: &syscall.SysProcAttr{
			Setsid: true,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to fork process: %w", err)
	}

	if pid > 0 {
		// Parent process - exit
		os.Exit(0)
	}

	// Child process continues as daemon
	return nil
}