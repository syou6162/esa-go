// Package esa provides esa.io REST API wire mechanics.
//
// NewClient creates a concrete client and accepts options for HTTP transport
// and base URL configuration. Small role interfaces are available for
// consumers that need only selected read, write, body-update, image-upload,
// or team-name capabilities. ErrNotFound identifies missing posts from
// category searches and HTTP 404 responses. Error messages redact secrets.
//
// The revision endpoints (RevisionReader, RevisionRollbacker) wrap esa.io's
// beta Revision API, which is not part of the official API reference; see the
// Revision type for primary sources.
package esa
