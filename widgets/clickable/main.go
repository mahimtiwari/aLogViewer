package clickable

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Clickable struct {
	widget.BaseWidget
	Content fyne.CanvasObject
	OnClick func()
	win     fyne.Window
}

func (c *Clickable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.Content)
}

func (c *Clickable) Tapped(_ *fyne.PointEvent) {
	if c.OnClick != nil {
		c.OnClick()

	}
}

func NewClickable(content fyne.CanvasObject, onClick func()) *Clickable {
	c := &Clickable{
		Content: content,
		OnClick: onClick,
	}
	c.ExtendBaseWidget(c)
	return c
}
