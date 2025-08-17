package clickable

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type Clickable struct {
	widget.BaseWidget
	Content fyne.CanvasObject
	OnClick func()
	OnHover func(mouse *desktop.MouseEvent)
}

func (c *Clickable) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.Content)
}

func (c *Clickable) MouseMoved(mouse *desktop.MouseEvent) {
	if c.OnHover != nil {
		c.OnHover(mouse)
	}
}

func (c *Clickable) Tapped(_ *fyne.PointEvent) {
	if c.OnClick != nil {
		c.OnClick()

	}
}

func NewClickable(content fyne.CanvasObject, onClick func(), onHover func(mouse *desktop.MouseEvent)) *Clickable {
	c := &Clickable{
		Content: content,
		OnClick: onClick,
		OnHover: onHover,
	}
	c.ExtendBaseWidget(c)
	return c
}
