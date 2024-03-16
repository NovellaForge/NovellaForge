package CalsWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"log"
	"time"
)

type Loading struct {
	progress binding.Float
	complete chan struct{}
	status   binding.String
	label    *widget.Label
	bar      *widget.ProgressBar
	box      *fyne.Container
	min      float64
	max      float64
	widget.BaseWidget
}

type loadingRenderer struct {
	objects []fyne.CanvasObject
	loading *Loading
}

func (l loadingRenderer) Destroy() {}

func (l loadingRenderer) Layout(size fyne.Size) {
	l.loading.box = container.NewVBox()
	l.loading.label = widget.NewLabelWithData(l.loading.BindStatus())
	l.loading.label.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
	l.loading.label.Alignment = fyne.TextAlignCenter
	l.loading.bar = widget.NewProgressBarWithData(l.loading.BindProgress())
	l.loading.bar.Min = l.loading.min
	l.loading.bar.Max = l.loading.max
	l.loading.box.Add(l.loading.label)
	l.loading.box.Add(l.loading.bar)
	l.loading.box.Resize(size)
	l.Refresh()
}

func (l loadingRenderer) MinSize() fyne.Size {
	return l.loading.box.MinSize()
}

func (l loadingRenderer) Objects() []fyne.CanvasObject {
	return l.objects
}

func (l loadingRenderer) Refresh() {
	l.loading.box.Refresh()
}

func NewLoading(loadingChan chan struct{}, minMax ...float64) *Loading {
	loading := &Loading{
		progress: binding.NewFloat(),
		complete: loadingChan,
		status:   binding.NewString(),
		label:    widget.NewLabel(""),
		bar:      widget.NewProgressBar(),
		box:      container.NewVBox(),
	}
	switch len(minMax) {
	case 1:
		loading.min = 0
		loading.max = minMax[0]
	case 2:
		loading.min = minMax[0]
		loading.max = minMax[1]
	default:
		loading.min = 0
		loading.max = 1
	}
	loading.ExtendBaseWidget(loading)
	return loading
}

func (l *Loading) SetProgress(progress float64, timeToSleep time.Duration, status ...string) {
	err := l.progress.Set(progress)
	if err != nil {
		log.Println(err)
	}
	if len(status) > 0 {
		err = l.status.Set(status[0])
		if err != nil {
			log.Println(err)
		}
	}
	time.Sleep(timeToSleep)
}
func (l *Loading) BindProgress() binding.Float {
	return l.progress
}
func (l *Loading) BindStatus() binding.String {
	return l.status
}
func (l *Loading) SetStatus(status string) {
	err := l.status.Set(status)
	if err != nil {
		log.Println(err)
	} else {
		log.Println(status)
	}
}
func (l *Loading) Complete() {
	l.complete <- struct{}{}
	close(l.complete)
	l.SetStatus("Complete")
}

func (l *Loading) CreateRenderer() fyne.WidgetRenderer {
	return &loadingRenderer{loading: l, objects: []fyne.CanvasObject{l.box, l.label, l.bar}}
}
