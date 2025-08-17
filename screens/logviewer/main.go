package logviewer

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func LogViewerScreen(path string) fyne.CanvasObject {
	content, _ := os.ReadFile(path)

	return container.NewVBox(
		widget.NewLabel("Logs of "+path),
		widget.NewLabel(string(content)),
	)
}
