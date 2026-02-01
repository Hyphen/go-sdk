// Package hyphen provides a Go SDK for the Hyphen platform services including
// Toggle (feature flags), NetInfo (geo information), and Link (URL shortening).
package hyphen

import (
	"github.com/Hyphen/go-sdk/pkg/env"
	"github.com/Hyphen/go-sdk/pkg/link"
	"github.com/Hyphen/go-sdk/pkg/netinfo"
	"github.com/Hyphen/go-sdk/pkg/toggle"
)

// ClientOptions contains configuration for creating a unified Hyphen client
type ClientOptions struct {
	// Common authentication
	APIKey       string // API key for Link and NetInfo services
	PublicAPIKey string // Public API key for Toggle service

	// Toggle options
	ApplicationID       string          // Application ID for Toggle
	Environment         string          // Environment for Toggle (defaults to "development")
	DefaultContext      *toggle.Context // Default context for Toggle evaluations
	HorizonURLs         []string        // Custom Horizon URLs for Toggle
	DefaultTargetingKey string          // Default targeting key for Toggle

	// NetInfo options
	NetInfoBaseURI string // Base URI for NetInfo service

	// Link options
	OrganizationID string   // Organization ID for Link service
	LinkURIs       []string // Custom URIs for Link service
}

// Re-export options types for convenience
type (
	ToggleOptions  = toggle.ToggleOptions
	NetInfoOptions = netinfo.NetInfoOptions
	LinkOptions    = link.LinkOptions
)

// Re-export main types for convenience
type (
	// Toggle types
	Toggle            = toggle.Toggle
	ToggleContext     = toggle.Context
	ToggleUser        = toggle.User
	ToggleCustomAttrs = toggle.CustomAttributes

	// NetInfo types
	NetInfo = netinfo.NetInfo
	IPInfo  = netinfo.IPInfo

	// Link types
	Link                   = link.Link
	ShortCodeResponse      = link.ShortCodeResponse
	CreateShortCodeOptions = link.CreateShortCodeOptions
	UpdateShortCodeOptions = link.UpdateShortCodeOptions
	CreateQRCodeOptions    = link.CreateQRCodeOptions
	QRCodeResponse         = link.QRCodeResponse
	GetShortCodesResponse  = link.GetShortCodesResponse
	GetQRCodesResponse     = link.GetQRCodesResponse
	GetCodeStatsResponse   = link.GetCodeStatsResponse

	// EnvOptions for environment variable loading
	EnvOptions = env.EnvOptions
)

// Client is the main Hyphen SDK client that provides access to all services
type Client struct {
	Toggle  *toggle.Toggle
	NetInfo *netinfo.NetInfo
	Link    *link.Link
	options *ClientOptions
}

// New creates a new Hyphen client with all services configured
func New(opts ClientOptions) (*Client, error) {
	client := &Client{
		options: &opts,
	}

	// Initialize Toggle service if public API key provided
	if opts.PublicAPIKey != "" && opts.ApplicationID != "" {
		t, err := NewToggle(opts.PublicAPIKey, opts.ApplicationID, ToggleOptions{
			Environment:         opts.Environment,
			DefaultContext:      opts.DefaultContext,
			HorizonURLs:         opts.HorizonURLs,
			DefaultTargetingKey: opts.DefaultTargetingKey,
		})
		if err != nil {
			return nil, err
		}
		client.Toggle = t
	}

	// Initialize NetInfo service if API key provided
	if opts.APIKey != "" {
		n, err := NewNetInfo(opts.APIKey, NetInfoOptions{
			BaseURI: opts.NetInfoBaseURI,
		})
		if err == nil {
			client.NetInfo = n
		}
	}

	// Initialize Link service if API key and organization ID provided
	if opts.APIKey != "" && opts.OrganizationID != "" {
		l, err := NewLink(opts.APIKey, opts.OrganizationID, LinkOptions{
			URIs: opts.LinkURIs,
		})
		if err == nil {
			client.Link = l
		}
	}

	return client, nil
}

// NewToggle creates a new Toggle client
func NewToggle(publicAPIKey, applicationID string, opts ToggleOptions) (*toggle.Toggle, error) {
	return toggle.New(publicAPIKey, applicationID, opts)
}

// NewNetInfo creates a new NetInfo client
func NewNetInfo(apiKey string, opts NetInfoOptions) (*netinfo.NetInfo, error) {
	return netinfo.New(apiKey, opts)
}

// NewLink creates a new Link client
func NewLink(apiKey, organizationID string, opts LinkOptions) (*link.Link, error) {
	return link.New(apiKey, organizationID, opts)
}

// LoadEnv loads environment variables from .env files
var LoadEnv = env.LoadEnv

// Re-export constants
const (
	QRSizeSmall  = link.QRSizeSmall
	QRSizeMedium = link.QRSizeMedium
	QRSizeLarge  = link.QRSizeLarge
)
