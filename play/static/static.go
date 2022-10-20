package static

import "embed"

// FS is an [fs.FS] implementation containing all static files needed for serving the Grawkit
// playground.
//
//go:embed *
var FS embed.FS
