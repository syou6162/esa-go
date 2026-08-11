// Package esa provides esa.io REST API wire mechanics.
//
// NewClient creates a concrete client and accepts options for HTTP transport
// and base URL configuration. Small role interfaces are available for
// consumers that need only selected read, write, body-update, image-upload,
// or team-name capabilities. ErrNotFound identifies missing posts from
// category searches and HTTP 404 responses. Error messages redact secrets.
package esa
