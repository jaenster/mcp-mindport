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
	configFile     string
	daemonMode     bool
	domain         string
	createDomain   bool
	listDomains    bool
	defaultDomain  string
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
	rootCmd.Flags().StringVar(&domain, "domain", "", "start server in specific domain context (e.g., 'project1' or 'team:backend')")
	rootCmd.Flags().BoolVar(&createDomain, "create-domain", false, "create domain if it doesn't exist (use with --domain)")
	rootCmd.Flags().BoolVar(&listDomains, "list-domains", false, "list all available domains and exit")
	rootCmd.Flags().StringVar(&defaultDomain, "default-domain", "", "set the default domain for the server (overrides config)")
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

	// Override default domain if specified via CLI
	if defaultDomain != "" {
		cfg.Domain.DefaultDomain = defaultDomain
		log.Printf("Default domain set to: %s", defaultDomain)
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

	// Handle domain-specific operations
	if listDomains {
		handleListDomains(mcpServer)
		return
	}

	if domain != "" {
		if err := handleDomainStartup(mcpServer, domain, createDomain); err != nil {
			log.Fatalf("Failed to setup domain: %v", err)
		}
		cfg.Domain.CurrentDomain = domain
		log.Printf("Starting MindPort in domain context: %s", domain)
	} else {
		cfg.Domain.CurrentDomain = cfg.Domain.DefaultDomain
		log.Printf("Starting MindPort in default domain context: %s", cfg.Domain.DefaultDomain)
	}

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

// handleListDomains lists all available domains
func handleListDomains(server *mcp.Server) {
	// This would need access to domain manager, but for now we'll implement a basic version
	fmt.Println("Available domains:")
	fmt.Println("• default - Default domain for general resources")
	fmt.Println("\nNote: Use domain management tools within the MCP server for full domain listing")
}

// handleDomainStartup sets up domain-specific startup
func handleDomainStartup(server *mcp.Server, domainName string, createIfNotExists bool) error {
	// For now, we'll update the config to set the current domain
	// In a full implementation, this would validate the domain exists
	// and possibly create it if requested
	
	if domainName == "" {
		return fmt.Errorf("domain name cannot be empty")
	}

	// Basic validation
	if len(domainName) > 64 {
		return fmt.Errorf("domain name too long (max 64 characters)")
	}

	log.Printf("Domain startup configured for: %s", domainName)
	if createIfNotExists {
		log.Printf("Will create domain '%s' if it doesn't exist", domainName)
	}

	return nil
}