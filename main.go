package main

import (
	"alogviewer/widgets/clickable"
	"fmt"
	"image/color"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

type LogFile struct {
	Path    string
	Content []byte
	lbabel  *widget.Label
}

func (lfObj *LogFile) SetPath(path string) {
	lfObj.Path = path
	log.Println("Log file path set to:", path)
	lfObj.lbabel.SetText(fmt.Sprintf("Selected: %s", path))
}

func (lfObj *LogFile) SetContent(content []byte) {
	lfObj.Content = content
}

func main() {
	app := app.New()

	window := app.NewWindow("aLogViewer")
	rect := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	label := widget.NewLabel("Select or Drop a log File")
	pageBtnStack := container.NewStack(rect, container.NewCenter(label))

	lf := &LogFile{
		Path:    "",
		Content: nil,
		lbabel:  label,
	}

	pageBtn := clickable.NewClickable(pageBtnStack, func() {
		fmt.Println("Button clicked")
		file, err := dialog.File().Filter("Log files", "log").Load()

		if err != nil {
			log.Println("Error:", err)
			return
		}
		if file != "" {
			content, err := os.ReadFile(file)

			lf.SetPath(file)
			lf.SetContent(content)

			if err != nil {
				log.Println("Error:", err)

			}

		}

	})

	window.SetContent(pageBtn)

	window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {
		if len(uris) > 0 {
			lf.SetPath(uris[0].Path())
		}
	})

	window.Resize(fyne.NewSize(600, 400))
	window.ShowAndRun()

}
