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

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	externaltoolsmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/external_tools/mcp"
	kubernetesmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/mcp"
)

type MCPServerConfig struct {
	Mode       string
	Transport  string
	Port       string
	Kubernetes kubernetesmcp.Config
}

func main() {
	serverCfg := parseFlags()

	ovnkMcpServer := mcp.NewServer(
		&mcp.Implementation{Name: "ovn-kubernetes"},
		&mcp.ServerOptions{HasTools: true},
	)

	if serverCfg.Mode == "live-cluster" {
		k8sMcpServer, err := kubernetesmcp.NewMCPServer(serverCfg.Kubernetes)
		if err != nil {
			log.Fatalf("Failed to create OVN-K MCP server: %v", err)
		}
		log.Println("Adding Kubernetes tools to OVN-K MCP server")
		k8sMcpServer.AddTools(ovnkMcpServer)
		externalToolsServer, err := externaltoolsmcp.NewMCPServer(k8sMcpServer)
		if err != nil {
			log.Fatalf("Failed to create external tools MCP server: %v", err)
		}
		log.Println("Adding external tools to OVN-K MCP server")
		externalToolsServer.AddTools(ovnkMcpServer)
	}

	// Create a context that can be cancelled to shutdown the server.
	ctx, cancel := context.WithCancel(context.Background())

	// Create a channel to receive signals to shutdown the server.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Start a goroutine to handle signals to shutdown the server.
	var server *http.Server
	go func() {
		// Wait for a signal to shutdown the server.
		<-signalChan

		log.Printf("Shutting down server")

		// Cancel the context to shutdown the server.
		defer cancel()

		// Shutdown the http server if it is running.
		if server != nil {
			// Shutdown the http server.
			if err := server.Shutdown(ctx); err != nil {
				log.Printf("Failed to shutdown server: %v", err)
			}
		}
	}()

	switch serverCfg.Transport {
	case "stdio":
		t := &mcp.LoggingTransport{Transport: &mcp.StdioTransport{}, Writer: os.Stderr}
		if err := ovnkMcpServer.Run(ctx, t); err != nil && err != context.Canceled {
			log.Printf("Server failed: %v", err)
		}
	case "http":
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return ovnkMcpServer
		}, nil)
		log.Printf("Listening on localhost:%s", serverCfg.Port)
		server = &http.Server{
			Addr:    fmt.Sprintf("localhost:%s", serverCfg.Port),
			Handler: handler,
		}
		if err := server.ListenAndServe(); err != nil {
			log.Printf("HTTP server failed: %v", err)
		}
	default:
		log.Fatalf("Invalid transport: %s", serverCfg.Transport)
	}
}

func parseFlags() *MCPServerConfig {
	cfg := &MCPServerConfig{}
	flag.StringVar(&cfg.Mode, "mode", "live-cluster", "Mode of debugging: live-cluster or offline")
	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport to use: stdio or http")
	flag.StringVar(&cfg.Port, "port", "8080", "Port to use")
	flag.StringVar(&cfg.Kubernetes.Kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	flag.Parse()
	return cfg
}
