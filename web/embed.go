// Package web provides the embedded production build of the React frontend.
// Build the frontend first: cd web && npm run build
package web

import "embed"

// DistFS contains the compiled React application from web/dist/.
// The `all:` prefix ensures dotfiles (if any) are also included.
//
//go:embed all:dist
var DistFS embed.FS
