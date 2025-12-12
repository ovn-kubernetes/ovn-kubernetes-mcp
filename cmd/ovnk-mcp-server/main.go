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
	kernelmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/mcp"
	kubernetesmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kubernetes/mcp"
	mustgathermcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/must-gather/mcp"
	ovsmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/ovs/mcp"
	sosreportmcp "github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/sosreport/mcp"
)

type MCPServerConfig struct {
	Mode       string
	Transport  string
	Port       string
	Kubernetes kubernetesmcp.Config
}

// setupLiveCluster sets up the live cluster mode.
func setupLiveCluster(serverCfg *MCPServerConfig, server *mcp.Server) {
	k8sMcpServer, err := kubernetesmcp.NewMCPServer(serverCfg.Kubernetes)
	if err != nil {
		log.Fatalf("Failed to create OVN-K MCP server: %v", err)
	}
	log.Println("Adding Kubernetes tools to OVN-K MCP server")
	k8sMcpServer.AddTools(server)

	ovsServer := ovsmcp.NewMCPServer(k8sMcpServer)
	log.Println("Adding OVS tools to OVN-K MCP server")
	ovsServer.AddTools(server)

	kernelMcpServer := kernelmcp.NewMCPServer(k8sMcpServer)
	log.Println("Adding Kernel tools to OVN-K MCP server")
	kernelMcpServer.AddTools(server)
}

// setupOffline sets up the offline mode.
func setupOffline(server *mcp.Server) {
	sosreportServer := sosreportmcp.NewMCPServer()
	log.Println("Adding sosreport tools to OVN-K MCP server")
	sosreportServer.AddTools(server)

	mustGatherServer, err := mustgathermcp.NewMCPServer()
	if err != nil {
		log.Printf("Failed to create Must Gather MCP server, will not be able to use must gather tools: %v", err)
		return
	}
	log.Println("Adding Must Gather tools to OVN-K MCP server")
	mustGatherServer.AddTools(server)
}

func main() {
	serverCfg := parseFlags()

	ovnkMcpServer := mcp.NewServer(
		&mcp.Implementation{Name: "ovn-kubernetes"},
		&mcp.ServerOptions{HasTools: true},
	)

	// Setup the MCP server based on the mode.
	switch serverCfg.Mode {
	case "live-cluster":
		setupLiveCluster(serverCfg, ovnkMcpServer)
	case "offline":
		setupOffline(ovnkMcpServer)
	case "both":
		setupLiveCluster(serverCfg, ovnkMcpServer)
		setupOffline(ovnkMcpServer)
	default:
		log.Fatalf("Invalid mode: %s. Valid modes are: live-cluster, offline, both", serverCfg.Mode)
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
	flag.StringVar(&cfg.Mode, "mode", "live-cluster", "Mode of debugging: live-cluster or offline or both")
	flag.StringVar(&cfg.Transport, "transport", "stdio", "Transport to use: stdio or http")
	flag.StringVar(&cfg.Port, "port", "8080", "Port to use")
	flag.StringVar(&cfg.Kubernetes.Kubeconfig, "kubeconfig", "", "Path to the kubeconfig file")
	flag.Parse()
	return cfg
}
