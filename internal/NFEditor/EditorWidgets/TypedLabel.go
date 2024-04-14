package EditorWidgets

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type TypedLabelRenderer struct {
	label *TypedLabel
}

func (t TypedLabelRenderer) Destroy() {}

func (t TypedLabelRenderer) Layout(size fyne.Size) {
	t.label.container.Layout.Layout(t.label.container.Objects, size)
}

func (t TypedLabelRenderer) MinSize() fyne.Size {
	return t.label.container.MinSize()
}

func (t TypedLabelRenderer) Objects() []fyne.CanvasObject {
	return t.label.container.Objects
}

func (t TypedLabelRenderer) Refresh() {
	t.label.container.Refresh()
}

type TypedLabel struct {
	labelType ValueType
	label     *widget.Label
	container *fyne.Container
	widget.BaseWidget
}

func (t *TypedLabel) CreateRenderer() fyne.WidgetRenderer {
	return &TypedLabelRenderer{label: t}

}

func NewTypedLabel(t ValueType, text string) *TypedLabel {
	label := &TypedLabel{
		labelType: t,
		label:     widget.NewLabel(text),
		container: container.NewStack(),
	}
	label.ExtendBaseWidget(label)
	label.container.Add(label.label)
	label.SetType(t)
	return label
}

func (t *TypedLabel) SetType(labelType ValueType) {
	t.labelType = labelType
	switch labelType {
	case IntType:
		t.label.Importance = widget.MediumImportance
	case FloatType:
		t.label.Importance = widget.HighImportance
	case BooleanType:
		b, err := strconv.ParseBool(t.Text())
		if err != nil {
			t.label.Importance = widget.WarningImportance
		} else {
			if b {
				t.label.Importance = widget.SuccessImportance
			} else {
				t.label.Importance = widget.DangerImportance
			}
		}
	case StringType:
		t.label.Importance = widget.LowImportance
	default: // ObjectType
		t.label.Text = "Object..."
		t.label.Importance = widget.WarningImportance
	}
}

func (t *TypedLabel) Text() string {
	return t.label.Text
}

func (t *TypedLabel) SetText(val string) {
	t.label.SetText(val)
}
