package CalsWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"image/color"
)

type MultiImage struct {
	widget.BaseWidget
	widget.DisableableWidget
	sizing     NFWidget.Sizing
	padding    NFWidget.Padding
	images     []*canvas.Image
	stack      *fyne.Container
	background color.Color
	Index      int
	Loop       bool
	OnTap      func()
	OnHover    func()
	WhileHover func()
	EndHover   func()
}

func (m *MultiImage) SetLoop(Loop bool) {
	m.Loop = Loop
}

func (m *MultiImage) SetOnTap(OnTap func()) {
	m.OnTap = OnTap
}

func (m *MultiImage) SetOnHover(OnHover func()) {
	m.OnHover = OnHover
}

func (m *MultiImage) SetWhileHover(WhileHover func()) {
	m.WhileHover = WhileHover
}

func (m *MultiImage) SetEndHover(EndHover func()) {
	m.EndHover = EndHover
}

func (m *MultiImage) MouseIn(event *desktop.MouseEvent) {
	if m.OnHover != nil {
		m.OnHover()
	}
}

func (m *MultiImage) MouseMoved(event *desktop.MouseEvent) {
	if m.WhileHover != nil {
		m.WhileHover()
	}
}

func (m *MultiImage) MouseOut() {
	if m.EndHover != nil {
		m.EndHover()
	}
}

func (m *MultiImage) Tapped(event *fyne.PointEvent) {
	if m.OnTap != nil {
		m.OnTap()
	}
}

func NewMultiImage(images []*canvas.Image) *MultiImage {
	for i, image := range images {
		if i == 0 {
			image.Show()
		} else {
			image.Hide()
		}
		image.ScaleMode = canvas.ImageScaleFastest
		image.FillMode = canvas.ImageFillContain
	}
	multi := &MultiImage{
		images:     images,
		Index:      0,
		Loop:       true,
		sizing:     NFWidget.NewSizing(100, 100, 1000, 1000, true, true),
		padding:    NFWidget.NewPadding(0, 0, 0, 0),
		background: color.Transparent,
	}
	multi.ExtendBaseWidget(multi)
	return multi
}

func (m *MultiImage) SetBackground(background color.Color) {
	m.background = background
}

func (m *MultiImage) Images() []*canvas.Image {
	return m.images
}

func (m *MultiImage) SetImages(images []*canvas.Image) {
	m.stack.RemoveAll()
	for _, image := range images {
		image.ScaleMode = canvas.ImageScaleFastest
		image.FillMode = canvas.ImageFillContain
		m.stack.Add(image)
	}
	m.images = images
	m.Refresh()
}

func (m *MultiImage) AddImage(image *canvas.Image) {
	image.ScaleMode = canvas.ImageScaleFastest
	image.FillMode = canvas.ImageFillContain
	m.images = append(m.images, image)
	m.stack.Add(image)
	m.Refresh()
}

func (m *MultiImage) InsertImage(index int, image *canvas.Image) {
	image.ScaleMode = canvas.ImageScaleFastest
	image.FillMode = canvas.ImageFillContain
	if index < 0 || index >= len(m.images) {
		m.AddImage(image)
		return
	}
	m.images = append(m.images[:index], append([]*canvas.Image{image}, m.images[index:]...)...)
	m.stack.RemoveAll()
	for _, i := range m.images {
		m.stack.Add(i)
	}
	m.Refresh()
}

func (m *MultiImage) RemoveImage(index int) {
	if index < 0 || index >= len(m.images) {
		return
	}
	m.images = append(m.images[:index], m.images[index+1:]...)
	m.stack.RemoveAll()
	for _, i := range m.images {
		m.stack.Add(i)
	}
	m.Refresh()
}

// RemoveImageByImage Will remove all images that match the file of image pointer unless first is true
func (m *MultiImage) RemoveImageByImage(image *canvas.Image, first ...bool) {
	onlyFirst := false
	for _, f := range first {
		onlyFirst = onlyFirst || f
	}
	for i, img := range m.images {
		if img.File == image.File {
			m.RemoveImage(i)
			if onlyFirst {
				return
			}
		}
	}
}

func (m *MultiImage) ClearImages() {
	m.images = nil
	m.stack.RemoveAll()
	m.Refresh()
}

func (m *MultiImage) NextIndex() {
	m.SetIndex(m.Index + 1)
}

func (m *MultiImage) PrevIndex() {
	m.SetIndex(m.Index - 1)
}

func (m *MultiImage) SetIndex(Index int, Refresh ...bool) {
	refresh := false
	for _, r := range Refresh {
		refresh = refresh || r
	}
	for Index < 0 {
		if m.Loop {
			Index = len(m.images) + Index
		} else {
			Index = 0
		}
	}
	if m.images == nil || len(m.images) == 0 {
		m.Index = Index
		return
	}
	if Index >= len(m.images) {
		Index = len(m.images) - 1
		if m.Loop {
			Index = Index % len(m.images)
			if Index != 0 {
				Index -= 1
			}
		}
	}
	if Index == m.Index && !refresh {
		return
	}
	m.Index = Index
	for i, image := range m.images {
		if i == m.Index {
			image.Show()
		} else if !image.Hidden {
			image.Hide()
		}
	}
}

func (m *MultiImage) CreateRenderer() fyne.WidgetRenderer {
	stack := container.NewStack()
	for _, image := range m.images {
		stack.Add(image)
	}
	m.stack = stack
	return &multiImageRenderer{objects: []fyne.CanvasObject{stack}, multi: m}
}

type multiImageRenderer struct {
	objects []fyne.CanvasObject
	multi   *MultiImage
}

func (r *multiImageRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *multiImageRenderer) Refresh() {
	r.multi.SetIndex(r.multi.Index, true)
}

func (r *multiImageRenderer) BackgroundColor() color.Color {
	return r.multi.background
}

func (r *multiImageRenderer) Destroy() {

}

func (r *multiImageRenderer) MinSize() fyne.Size {
	return fyne.NewSize(r.multi.sizing.MinWidth, r.multi.sizing.MinHeight)
}

func (r *multiImageRenderer) Layout(size fyne.Size) {
	m := r.multi

	if m.sizing.FitWidth {
		size.Width = size.Width - m.padding.Horizontal()
	} else {
		size.Width = min(size.Width-m.padding.Horizontal(), m.sizing.MaxWidth-m.padding.Horizontal())
	}
	if m.sizing.FitHeight {
		size.Height = size.Height - m.padding.Vertical()
	} else {
		size.Height = min(size.Height-m.padding.Vertical(), m.sizing.MaxHeight-m.padding.Vertical())
	}

	m.stack.Resize(size)
	m.stack.Move(fyne.NewPos(m.padding.Left, m.padding.Top))
}
