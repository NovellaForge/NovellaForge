package EditorWidgets

import (
	"fyne.io/fyne/v2/widget"
	"strconv"
)

type TypedLabel struct {
	labelType ValueType
	*widget.Label
}

func NewTypedLabel(t ValueType, text string) *TypedLabel {
	label := &TypedLabel{
		labelType: t,
		Label:     widget.NewLabel(text),
	}
	label.SetType(t)
	return label
}

func (t *TypedLabel) SetType(labelType ValueType) {
	t.labelType = labelType
	switch labelType {
	case IntType:
		t.Importance = widget.MediumImportance
	case FloatType:
		t.Importance = widget.HighImportance
	case BooleanType:
		b, err := strconv.ParseBool(t.Text)
		if err != nil {
			t.Importance = widget.WarningImportance
		} else {
			if b {
				t.Importance = widget.SuccessImportance
			} else {
				t.Importance = widget.DangerImportance
			}
		}
	case StringType:
		t.Importance = widget.LowImportance
	default: // ObjectType
		t.Text = "Object..."
		t.Importance = widget.WarningImportance
	}
}
