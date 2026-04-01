package mcp

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/kernel/types"
	"github.com/ovn-kubernetes/ovn-kubernetes-mcp/pkg/utils"
)

// GetNFT MCP handler for nftables operations.
// GetNFT retrieves nftables configuration from a Kubernetes node.
func (s *MCPServer) GetNFT(ctx context.Context, req *mcp.CallToolRequest, in types.ListNFTParams) (*mcp.CallToolResult, types.Result, error) {
	err := s.utilityExists(ctx, req, in.Node, "nft")
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting nft data: failed to verify nft utility availability in configured image: %w", err)
	}

	if err := validateNFTCommand(in.Command); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting nft data: %w", err)
	}
	if err := utils.ValidateSafeString(in.AddressFamilies, "address families", true, utils.ShellMetaCharactersTypeDefault); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting nft data: %w", err)
	}
	if err := validateNFTAddressFamily(in.AddressFamilies); err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting nft data: %w", err)
	}

	cmd := utils.NewCommand("nft")
	cmd.Add(strings.Fields(in.Command)...)
	cmd.AddIf(in.AddressFamilies != "", in.AddressFamilies)

	stdout, err := s.executeCommand(ctx, req, in.Node, cmd.Build())
	if err != nil {
		return nil, types.Result{}, fmt.Errorf("error while getting nft data: %w", err)
	}
	stdout = strings.Join(in.HeadTailParams.Apply(strings.Split(stdout, "\n"), defaultMaxOutputLines), "\n")
	return nil, types.Result{Data: stdout}, nil
}

// validateNFTCommand validates nftables command. This is to put a limitation on the use of the tool.
func validateNFTCommand(command string) error {
	if _, err := strconv.Atoi(command); err == nil {
		return fmt.Errorf("invalid nft command: %s", command)
	}
	validCommand := map[string]bool{
		"list ruleset":    true,
		"list tables":     true,
		"list chains":     true,
		"list sets":       true,
		"list maps":       true,
		"list flowtables": true,
	}

	if !validCommand[strings.TrimSpace(command)] {
		return fmt.Errorf("invalid nft command: %s", command)
	}
	return nil
}

// validateNFTAddressFamily validates nftables address family.
func validateNFTAddressFamily(addrFamily string) error {
	if addrFamily == "" {
		return nil
	}

	if _, err := strconv.Atoi(addrFamily); err == nil {
		return fmt.Errorf("invalid nft address family: %s", addrFamily)
	}
	validaddrFamily := map[string]bool{
		"ip":     true,
		"ip6":    true,
		"inet":   true,
		"arp":    true,
		"bridge": true,
		"netdev": true,
	}

	if !validaddrFamily[strings.TrimSpace(addrFamily)] {
		return fmt.Errorf("invalid nft address family: %s", addrFamily)
	}
	return nil
}
