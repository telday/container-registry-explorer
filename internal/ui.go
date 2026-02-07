package internal

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	explorer "github.com/telday/container-registry-explorer/pkg"
)

type ExplorerApp struct {
	registry string
	app      *tview.Application
}

func NewExplorerApp(registry string) *ExplorerApp {
	app := tuiApp()
	explorer := &ExplorerApp{
		registry: registry,
		app:      app,
	}

	return explorer
}

func (e *ExplorerApp) Run() error {
	if err := e.app.Run(); err != nil {
		return err
	}

	return nil
}

func tuiApp() *tview.Application {
	app := tview.NewApplication()

	imageBox := getImageBox(app)
	tagsBox := getTagsBox()

	flexView := tview.NewFlex().
		AddItem(imageBox, 0, 1, true).
		AddItem(tagsBox, 0, 1, false)

	pages := tview.NewPages()
	pages.AddPage("Main", flexView, true, true)

	tagOptsModal := tview.NewModal().
		AddButtons([]string{"Quit"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Quit" {
				pages.SwitchToPage("Main")
				return
			}
		})

	pages.AddPage("Image Opts", tagOptsModal, true, false)

	// Sets up our basic movement values
	flexView.SetInputCapture(flexViewMovement(app, imageBox, tagsBox))
	imageBox.SetSelectedFunc(updateTagsBoxFunction(tagsBox))
	loadInitialImages(imageBox)

	tagsBox.SetSelectedFunc(func(int, string, string, rune) {
		tagOptsModal.SetText("Options for image: ")
		pages.SwitchToPage("Image Opts")
	})

	return app.SetRoot(pages, true).SetFocus(pages)
}

func flexViewMovement(app *tview.Application, imageBox, tagsBox *tview.List) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			tagsBox.Clear()
			app.SetFocus(imageBox)
			return nil
		} else if event.Key() == tcell.KeyRight && imageBox.HasFocus() {
			// ToDo: This could easily be done by switching between the indexed items for the flexview
			app.SetFocus(tagsBox)
			return nil
		} else if event.Key() == tcell.KeyLeft && tagsBox.HasFocus() {
			app.SetFocus(imageBox)
			return nil
		}
		return event
	}
}

func loadInitialImages(imageBox *tview.List) {
	images := explorer.GetImageNames("localhost")
	for _, image := range images {
		imageBox.AddItem(image, "", 0, nil)
	}
}

func getImageBox(app *tview.Application) *tview.List {
	list := tview.NewList().
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		}).
		ShowSecondaryText(false)

	// These methods return a Box obj so we cannot directly chain them
	// See: https://pkg.go.dev/github.com/rivo/tview#hdr-Type_Hierarchy
	list.
		SetBorder(true).
		SetTitle("Images")

	return list
}

func getTagsBox() *tview.List {
	list := tview.NewList().
		ShowSecondaryText(false)

	list.
		SetBorder(true).
		SetTitle("Tags")

	return list
}

func updateTagsBoxFunction(tagsBox *tview.List) func(int, string, string, rune) {
	return func(_ int, imageName, _ string, _ rune) {
		tagsBox.Clear()

		tags := explorer.GetTags("localhost/" + imageName)
		for _, tag := range tags {
			tagsBox.AddItem(tag, "", 0, nil)
		}
		tagsBox.SetTitle(fmt.Sprintf("Image %s tags", imageName))
	}
}
