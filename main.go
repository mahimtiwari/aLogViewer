package main

import (
	"alogviewer/screens/logviewer"
	"alogviewer/widgets/clickable"
	"fmt"
	"regexp"
	"runtime/debug"

	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/sqweek/dialog"
)

type LogFile struct {
	Path       string
	window     fyne.Window
	viewerBool bool
}

func (lfObj *LogFile) SetPath(path string) {
	valid, _ := regexp.MatchString(`\.log(\.\d+)?$`, path)
	if len(path) < 4 || !valid || lfObj.Path == path {
		log.Println("Selected file is not a .log file or is the same as the current one")
		return

	}

	lfObj.Path = path
	log.Println("Log file path set to:", path)
	lfObj.viewerBool = true
	lfObj.window.SetContent(logviewer.LogViewerScreen(lfObj.Path))

}

func fileAreaClicked(lf *LogFile) {
	file, err := dialog.File().Title("Select a log file").Filter("Log files", "log", "log.*").Load()

	if err != nil {
		log.Println("Error:", err)
		return
	}

	lf.SetPath(file)

}

func fileAreaHovered(mouse *desktop.MouseEvent) {

}

func main() {

	app := app.New()
	window := app.NewWindow("aLogViewer")
	rect := canvas.NewRectangle(color.NRGBA{R: 0, G: 0, B: 0, A: 0})
	label := widget.NewLabel("Select or Drop a log File")
	pageBtnStack := container.NewStack(rect, container.NewCenter(label))
	lf := &LogFile{
		Path:       "",
		window:     window,
		viewerBool: false,
	}

	pageBtn := clickable.NewClickable(
		pageBtnStack,
		func() { fileAreaClicked(lf) },
		fileAreaHovered)

	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open", func() { fileAreaClicked(lf) }),
	)

	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Refresh", func() {
			fmt.Println("Refresh clicked")
		}),
		fyne.NewMenuItem("Free Memory", func() {
			debug.FreeOSMemory()
		}),
	)

	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", func() {
			fmt.Println("About clicked")
		}),
	)

	window.SetMainMenu(fyne.NewMainMenu(
		fileMenu,
		viewMenu,
		helpMenu,
	))

	window.SetContent(pageBtn)

	window.SetOnDropped(func(pos fyne.Position, uris []fyne.URI) {

		if len(uris) > 0 && !lf.viewerBool {
			lf.SetPath(uris[0].Path())
		}
	})

	window.Resize(fyne.NewSize(600, 400))
	window.ShowAndRun()

}
