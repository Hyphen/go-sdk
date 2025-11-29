// Package hyphen provides a Go SDK for the Hyphen platform services including
// Toggle (feature flags), NetInfo (geo information), and Link (URL shortening).
package hyphen

import (
	"github.com/Hyphen/hyphen-go-sdk/pkg/env"
	"github.com/Hyphen/hyphen-go-sdk/pkg/link"
	"github.com/Hyphen/hyphen-go-sdk/pkg/netinfo"
	"github.com/Hyphen/hyphen-go-sdk/pkg/toggle"
)

// Options contains all configuration options for Hyphen services
type Options struct {
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

// Option is a functional option for configuring Hyphen services
type Option func(*Options)

// WithAPIKey sets the API key (used by Link and NetInfo services)
func WithAPIKey(key string) Option {
	return func(o *Options) {
		o.APIKey = key
	}
}

// WithPublicAPIKey sets the public API key (used by Toggle service)
func WithPublicAPIKey(key string) Option {
	return func(o *Options) {
		o.PublicAPIKey = key
	}
}

// WithApplicationID sets the application ID (used by Toggle service)
func WithApplicationID(id string) Option {
	return func(o *Options) {
		o.ApplicationID = id
	}
}

// WithEnvironment sets the environment (used by Toggle service)
func WithEnvironment(env string) Option {
	return func(o *Options) {
		o.Environment = env
	}
}

// WithOrganizationID sets the organization ID (used by Link service)
func WithOrganizationID(id string) Option {
	return func(o *Options) {
		o.OrganizationID = id
	}
}

// WithDefaultContext sets the default context for Toggle evaluations
func WithDefaultContext(ctx *toggle.Context) Option {
	return func(o *Options) {
		o.DefaultContext = ctx
	}
}

// WithHorizonURLs sets custom Horizon URLs for the Toggle service
func WithHorizonURLs(urls []string) Option {
	return func(o *Options) {
		o.HorizonURLs = urls
	}
}

// WithDefaultTargetingKey sets the default targeting key for Toggle
func WithDefaultTargetingKey(key string) Option {
	return func(o *Options) {
		o.DefaultTargetingKey = key
	}
}

// WithNetInfoBaseURI sets the base URI for the NetInfo service
func WithNetInfoBaseURI(uri string) Option {
	return func(o *Options) {
		o.NetInfoBaseURI = uri
	}
}

// WithLinkURIs sets custom URIs for the Link service
func WithLinkURIs(uris []string) Option {
	return func(o *Options) {
		o.LinkURIs = uris
	}
}

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
	options *Options
}

// New creates a new Hyphen client with all services configured
func New(options ...Option) (*Client, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	client := &Client{
		options: opts,
	}

	// Initialize Toggle service if public API key provided
	if opts.PublicAPIKey != "" {
		t, err := NewToggle(options...)
		if err != nil {
			return nil, err
		}
		client.Toggle = t
	}

	// Initialize NetInfo service if API key provided
	if opts.APIKey != "" {
		n, err := NewNetInfo(options...)
		if err == nil {
			client.NetInfo = n
		}
	}

	// Initialize Link service if API key provided
	if opts.APIKey != "" {
		l, err := NewLink(options...)
		if err == nil {
			client.Link = l
		}
	}

	return client, nil
}

// NewToggle creates a new Toggle client
func NewToggle(options ...Option) (*toggle.Toggle, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	toggleOpts := []toggle.Option{}

	if opts.PublicAPIKey != "" {
		toggleOpts = append(toggleOpts, toggle.WithPublicAPIKey(opts.PublicAPIKey))
	}
	if opts.ApplicationID != "" {
		toggleOpts = append(toggleOpts, toggle.WithApplicationID(opts.ApplicationID))
	}
	if opts.Environment != "" {
		toggleOpts = append(toggleOpts, toggle.WithEnvironment(opts.Environment))
	}
	if opts.DefaultContext != nil {
		toggleOpts = append(toggleOpts, toggle.WithDefaultContext(opts.DefaultContext))
	}
	if len(opts.HorizonURLs) > 0 {
		toggleOpts = append(toggleOpts, toggle.WithHorizonURLs(opts.HorizonURLs))
	}
	if opts.DefaultTargetingKey != "" {
		toggleOpts = append(toggleOpts, toggle.WithDefaultTargetingKey(opts.DefaultTargetingKey))
	}

	return toggle.New(toggleOpts...)
}

// NewNetInfo creates a new NetInfo client
func NewNetInfo(options ...Option) (*netinfo.NetInfo, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	netinfoOpts := []netinfo.Option{}

	if opts.APIKey != "" {
		netinfoOpts = append(netinfoOpts, netinfo.WithAPIKey(opts.APIKey))
	}
	if opts.NetInfoBaseURI != "" {
		netinfoOpts = append(netinfoOpts, netinfo.WithBaseURI(opts.NetInfoBaseURI))
	}

	return netinfo.New(netinfoOpts...)
}

// NewLink creates a new Link client
func NewLink(options ...Option) (*link.Link, error) {
	opts := &Options{}
	for _, opt := range options {
		opt(opts)
	}

	linkOpts := []link.Option{}

	if opts.APIKey != "" {
		linkOpts = append(linkOpts, link.WithAPIKey(opts.APIKey))
	}
	if opts.OrganizationID != "" {
		linkOpts = append(linkOpts, link.WithOrganizationID(opts.OrganizationID))
	}
	if len(opts.LinkURIs) > 0 {
		linkOpts = append(linkOpts, link.WithURIs(opts.LinkURIs))
	}

	return link.New(linkOpts...)
}

// LoadEnv loads environment variables from .env files
var LoadEnv = env.LoadEnv

// Re-export constants
const (
	QRSizeSmall  = link.QRSizeSmall
	QRSizeMedium = link.QRSizeMedium
	QRSizeLarge  = link.QRSizeLarge
)
