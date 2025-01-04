package assets

import "embed"

// EditorPng is an embed.FS that contains the editor icon
//
//go:embed icons/editor.png
var EditorPng embed.FS

// TypesFS is an embed.FS that contains the type files for editor parsing
//
//go:embed layouts widgets functions
var TypesFS embed.FS

// LegalFS is an embedded file system containing the "legal" directory and its contents.
//
//go:embed legal
var LegalFS embed.FS
