package ui

import "embed"

// Files is the embedded file system for static files and templates.
// To use this, uncomment the embed directive below and ensure you have files in templates/ and static/.
//
// Uncomment the line below when you have template/static files to embed:
// //go:embed "templates/*" "static/*"
var Files embed.FS
