package NFVideo

import (
	"fyne.io/fyne/v2"
	xWidget "fyne.io/x/fyne/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFFS"
	"path/filepath"
)

func NewGifPlayer(path string, config NFFS.Configuration) (*xWidget.AnimatedGif, error) {
	//Check if the file exists
	_, err := NFFS.Stat(path, config)
	if err != nil {
		return &xWidget.AnimatedGif{}, err
	}

	gifBytes, err := NFFS.ReadFile(path, config)
	if err != nil {
		return &xWidget.AnimatedGif{}, err
	}

	resource := fyne.NewStaticResource(filepath.Base(path), gifBytes)
	gif, err := xWidget.NewAnimatedGifFromResource(resource)
	if err != nil {
		return &xWidget.AnimatedGif{}, err
	}

	return gif, nil
}
