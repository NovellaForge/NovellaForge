package assets

import "embed"

// EditorPng is an embed.FS that contains the editor icon
//
//go:embed icons/editor.png
var EditorPng embed.FS

// BinaryFS is an embed.FS that contains the binary files
//
//go:embed binaries
var BinaryFS embed.FS
