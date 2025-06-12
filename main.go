package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"mcp-mindport/internal/config"
	"mcp-mindport/internal/daemon"
	"mcp-mindport/internal/mcp"
	"mcp-mindport/internal/search"
	"mcp-mindport/internal/storage"

	"github.com/spf13/cobra"
)

var (
	configFile string
	daemonMode bool
)

var rootCmd = &cobra.Command{
	Use:   "mcp-mindport",
	Short: "MindPort - MCP Resource Server with optimized search capabilities",
	Long:  `MindPort: A Model Context Protocol server that provides optimized storage and search for AI systems`,
	Run:   runServer,
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "config file (default is $HOME/.mcp-mindport.yaml)")
	rootCmd.Flags().BoolVarP(&daemonMode, "daemon", "d", false, "run as daemon")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	store, err := storage.NewBadgerStore(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize search engine
	searchEngine, err := search.NewBleveSearch(cfg.Search.IndexPath)
	if err != nil {
		log.Fatalf("Failed to initialize search engine: %v", err)
	}
	defer searchEngine.Close()

	// Create MCP server
	mcpServer := mcp.NewServer(store, searchEngine, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	if daemonMode {
		// Run as daemon
		if err := daemon.Start(ctx, mcpServer, cfg); err != nil {
			log.Fatalf("Failed to start daemon: %v", err)
		}
	} else {
		// Run MCP server directly (stdio mode)
		go func() {
			if err := mcpServer.Start(ctx); err != nil {
				log.Printf("MCP server error: %v", err)
				cancel()
			}
		}()
	}

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down...")
	cancel()
}