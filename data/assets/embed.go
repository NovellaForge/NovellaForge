package assets

import "embed"

//go:embed icons/editor.png
var EditorPng embed.FS

// BinaryFS
//
// TODO These embedded Binaries need to be copied to the game folder instead of being unpacked by the editor,
// since it will be the game unpacking them for normal runtime
//
//go:embed binaries
var BinaryFS embed.FS
