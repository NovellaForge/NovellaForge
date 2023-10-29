package CalsWidgets

import "github.com/NovellaForge/NovellaForge/pkg/NFWidget"

// You can import this package to register all the widgets stored in this package,
// if your ide tells you it is unused add a _ prefix to the import
func init() {
	NFWidget.Register("Dialog", DialogHandler)

}
