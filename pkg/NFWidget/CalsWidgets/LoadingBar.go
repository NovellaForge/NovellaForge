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
	progress    binding.Float
	complete    chan struct{}
	status      binding.String
	timeToSleep time.Duration
	bar         *widget.ProgressBar
	label       *widget.Label
	Box         fyne.CanvasObject
}

func NewLoading(loadingChan chan struct{}, timeToSleep time.Duration, minMax ...float64) *Loading {
	loading := &Loading{
		progress: binding.NewFloat(),
		complete: loadingChan,
		status:   binding.NewString(),
	}
	loading.bar = widget.NewProgressBarWithData(loading.progress)
	loading.label = widget.NewLabelWithData(loading.status)
	loading.label.Alignment = fyne.TextAlignCenter
	loading.label.TextStyle = fyne.TextStyle{Bold: true, Italic: true}
	loading.Box = container.NewVBox(loading.label, loading.bar)
	loading.timeToSleep = timeToSleep
	switch len(minMax) {
	case 1:
		loading.bar.Min = 0
		loading.bar.Max = minMax[0]
	case 2:
		loading.bar.Min = minMax[0]
		loading.bar.Max = minMax[1]
	default:
		loading.bar.Min = 0
		loading.bar.Max = 1
	}
	return loading
}

func (l *Loading) SetProgress(progress float64, status ...string) {
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
	time.Sleep(l.timeToSleep)
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
	l.SetStatus("Complete")
}
func (l *Loading) GetProgress() float64 {
	return l.bar.Value
}
