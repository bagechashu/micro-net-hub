package internal

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
)

// RadiusAuthenticator handles RADIUS authentication
type RadiusAuthenticator struct {
	servers []RadiusServer
	mu      sync.RWMutex
}

// RadiusServer represents a single RADIUS server
type RadiusServer struct {
	Addr   string
	Secret string
}

// NewRadiusAuthenticator creates a new RADIUS authenticator
func NewRadiusAuthenticator(cfg RadiusConfig) (*RadiusAuthenticator, error) {
	authenticator := &RadiusAuthenticator{
		servers: make([]RadiusServer, 0),
	}

	// Parse radiusclient.conf to get servers file path
	serversFile := cfg.ServersFile
	if cfg.ConfigFile != "" {
		parsed, err := parseRadiusConfig(cfg.ConfigFile)
		if err != nil {
			log.Printf("[radius] 警告: 解析 radiusclient.conf 失败: %v", err)
		}
		if parsed != "" {
			serversFile = parsed
		}
	}

	// Parse servers file
	if serversFile == "" {
		return nil, fmt.Errorf("no RADIUS servers file configured")
	}

	servers, err := parseRadiusServers(serversFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RADIUS servers: %w", err)
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("no RADIUS servers configured")
	}

	authenticator.servers = servers
	log.Printf("[radius] 已配置 %d 个 RADIUS 服务器", len(servers))
	return authenticator, nil
}

// Authenticate attempts to authenticate a user against RADIUS servers
func (ra *RadiusAuthenticator) Authenticate(username, password string) (bool, error) {
	if username == "" || password == "" {
		return false, fmt.Errorf("username and password required")
	}

	ra.mu.RLock()
	servers := ra.servers
	ra.mu.RUnlock()

	if len(servers) == 0 {
		return false, fmt.Errorf("no RADIUS servers available")
	}

	var lastErr error
	// Try each server until one succeeds
	for _, server := range servers {
		ok, err := ra.authenticateWithServer(username, password, server)
		if err != nil {
			log.Printf("[radius] 服务器 %s 认证错误: %v", server.Addr, err)
			lastErr = err
			continue
		}
		if ok {
			log.Printf("[radius] 用户 %s 认证成功 (%s)", username, server.Addr)
			return true, nil
		}
	}

	log.Printf("[radius] 用户 %s 认证失败: 所有服务器都被拒绝", username)
	if lastErr != nil {
		return false, fmt.Errorf("authentication failed: %w", lastErr)
	}
	return false, nil
}

// authenticateWithServer performs authentication against a single RADIUS server
// using the radius.Exchange method with proper RFC 2865 attribute handling
func (ra *RadiusAuthenticator) authenticateWithServer(username, password string, server RadiusServer) (bool, error) {
	// Create RADIUS Access-Request packet
	// First parameter: packet type (CodeAccessRequest)
	// Second parameter: shared secret (bytes)
	packet := radius.New(radius.CodeAccessRequest, []byte(server.Secret))

	// Set User-Name attribute (attribute 1)
	// SetString automatically handles string encoding according to RFC 2865
	// rfc2865.UserName_SetString is the recommended method
	if err := rfc2865.UserName_SetString(packet, username); err != nil {
		return false, fmt.Errorf("failed to set username attribute: %w", err)
	}

	// Set User-Password attribute (attribute 2)
	// SetString automatically encrypts using RFC 2865 User-Password encryption
	// Encryption requires: packet.Secret and packet.Authenticator
	// New() automatically generates a random Authenticator
	if err := rfc2865.UserPassword_SetString(packet, password); err != nil {
		return false, fmt.Errorf("failed to set password attribute: %w", err)
	}

	// Set NAS-Identifier attribute (attribute 32)
	// This is optional but recommended to identify the NAS device
	if err := rfc2865.NASIdentifier_SetString(packet, "ocserv-nft"); err != nil {
		log.Printf("[radius] 警告: 无法设置 NAS-Identifier: %v", err)
	}

	// Create context with timeout
	// Always use context to handle timeouts properly
	// Use 3 second timeout for RADIUS requests to handle network latency
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Send RADIUS Access-Request and receive response
	// radius.Exchange handles UDP connection, packet encoding/decoding
	// and automatic response parsing
	response, err := radius.Exchange(ctx, packet, server.Addr)
	if err != nil {
		return false, fmt.Errorf("RADIUS request failed: %w", err)
	}

	// Check response code
	// - CodeAccessAccept (2): Authentication successful
	// - CodeAccessReject (3): Authentication failed
	// - CodeAccessChallenge (11): Server needs more information
	switch response.Code {
	case radius.CodeAccessAccept:
		return true, nil
	case radius.CodeAccessReject:
		return false, nil
	case radius.CodeAccessChallenge:
		return false, fmt.Errorf("RADIUS server requested challenge (not supported)")
	default:
		return false, fmt.Errorf("unexpected RADIUS response code: %v", response.Code)
	}
}

// parseRadiusConfig parses radiusclient.conf to extract servers file path
func parseRadiusConfig(configFile string) (string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return "", fmt.Errorf("failed to open config file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("[radius] 关闭配置文件出错: %v", cerr)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Look for "servers" directive
		if strings.HasPrefix(line, "servers") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading config file: %w", err)
	}

	return "", fmt.Errorf("no servers directive found in config")
}

// parseRadiusServers parses the servers file to extract server addresses and secrets
// Format: <ip_address> <secret>
func parseRadiusServers(serversFile string) ([]RadiusServer, error) {
	file, err := os.Open(serversFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open servers file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("[radius] 关闭服务器文件出错: %v", cerr)
		}
	}()

	var servers []RadiusServer
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse server line: <address> <secret>
		parts := strings.Fields(line)
		if len(parts) < 2 {
			log.Printf("[radius] 警告: 无效的服务器行: %s", line)
			continue
		}

		addr := parts[0]
		secret := parts[1]

		// Add default RADIUS port if not specified
		if !strings.Contains(addr, ":") {
			addr = net.JoinHostPort(addr, "1812")
		}

		servers = append(servers, RadiusServer{
			Addr:   addr,
			Secret: secret,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading servers file: %w", err)
	}

	return servers, nil
}
