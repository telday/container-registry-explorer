package internal

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	explorer "github.com/telday/container-registry-explorer/pkg"
)

type ExplorerApp struct {
	registry      string
	app           *tview.Application
	pages         *tview.Pages
	selectedImage string
	selectedTag   string
	statusText    *tview.TextView
}

func NewExplorerApp(registry string) *ExplorerApp {
	explorer := &ExplorerApp{
		registry: registry,
	}
	explorer.setupTuiApp()

	return explorer
}

func (e *ExplorerApp) Run() error {
	if err := e.app.Run(); err != nil {
		return err
	}

	return nil
}

func (e *ExplorerApp) setupTuiApp() {
	app := tview.NewApplication()

	imageBox := getImageBox(app)
	tagsBox := getTagsBox(app)

	// Status bar at bottom
	e.statusText = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	e.statusText.SetBorder(true)

	mainContent := tview.NewFlex().
		AddItem(imageBox, 0, 1, true).
		AddItem(tagsBox, 0, 1, false)

	flexView := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(mainContent, 0, 1, true).
		AddItem(e.statusText, 3, 0, false)

	pages := tview.NewPages()
	pages.AddPage("Main", flexView, true, true)

	tagOptsModal := tview.NewModal().
		AddButtons([]string{"Pull Image", "Copy SHA", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			switch buttonLabel {
			case "Pull Image":
				e.pullSelectedImage()
			case "Copy SHA":
				e.copyImageSHA()
			case "Cancel":
				pages.SwitchToPage("Main")
			}
		})

	pages.AddPage("Image Opts", tagOptsModal, true, false)

	// Sets up our basic movement values
	mainContent.SetInputCapture(flexViewMovement(app, imageBox, tagsBox))
	imageBox.SetSelectedFunc(e.updateTagsBoxFunction(tagsBox))
	e.loadInitialImages(imageBox)

	tagsBox.SetSelectedFunc(func(_ int, tagName string, _ string, _ rune) {
		// Store the selected image and tag
		e.selectedTag = tagName
		fullRef := e.getFullImageRef()
		tagOptsModal.SetText(fmt.Sprintf("Options for:\n%s", fullRef))
		pages.SwitchToPage("Image Opts")
	})

	e.pages = pages
	e.app = app.SetRoot(pages, true).SetFocus(pages)
}

func (e *ExplorerApp) getFullImageRef() string {
	return fmt.Sprintf("%s/%s:%s", e.registry, e.selectedImage, e.selectedTag)
}

func (e *ExplorerApp) pullSelectedImage() {
	imageRef := e.getFullImageRef()
	e.setStatus(fmt.Sprintf("[yellow]Pulling %s...[white]", imageRef))
	e.pages.SwitchToPage("Main")

	go func() {
		outputChan, errChan := explorer.PullImageAsync(imageRef)

		select {
		case err := <-errChan:
			if err != nil {
				e.app.QueueUpdateDraw(func() {
					e.setStatus(fmt.Sprintf("[red]Failed to pull %s: %v[white]", imageRef, err))
				})
				return
			}
		case output := <-outputChan:
			e.app.QueueUpdateDraw(func() {
				if output != "" {
					e.setStatus(fmt.Sprintf("[green]Successfully pulled %s[white]", imageRef))
				}
			})
		}
	}()
}

func (e *ExplorerApp) copyImageSHA() {
	imageRef := e.getFullImageRef()
	e.setStatus(fmt.Sprintf("[yellow]Getting SHA for %s...[white]", imageRef))
	e.pages.SwitchToPage("Main")

	go func() {
		digest, err := explorer.GetImageDigest(imageRef)
		e.app.QueueUpdateDraw(func() {
			if err != nil {
				e.setStatus(fmt.Sprintf("[red]Failed to get SHA: %v[white]", err))
				return
			}

			if err := clipboard.WriteAll(digest); err != nil {
				e.setStatus(fmt.Sprintf("[red]Failed to copy to clipboard: %v[white]", err))
				return
			}

			e.setStatus(fmt.Sprintf("[green]Copied SHA to clipboard: %s[white]", digest))
		})
	}()
}

func (e *ExplorerApp) setStatus(msg string) {
	e.statusText.SetText(msg)
}

func flexViewMovement(app *tview.Application, imageBox, tagsBox *SearchableList) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			tagsBox.Clear()
			app.SetFocus(imageBox)
			return nil
		} else if event.Key() == tcell.KeyRight && imageBox.HasFocus() {
			// ToDo: This could easily be done by switching between the indexed items for the flexview
			app.SetFocus(tagsBox.GetList())
			return nil
		} else if event.Key() == tcell.KeyLeft && tagsBox.HasFocus() {
			app.SetFocus(imageBox.GetList())
			return nil
		}
		return event
	}
}

func (e *ExplorerApp) loadInitialImages(imageBox *SearchableList) {
	images := explorer.GetImageNames(e.registry)
	for _, image := range images {
		imageBox.AddItem(image, "", 0, nil)
	}
}

func getImageBox(app *tview.Application) *SearchableList {
	sl := NewSearchableList(app)
	sl.AddItem("Quit", "Press to exit", 'q', func() {
		app.Stop()
	})

	// These methods return a Box obj so we cannot directly chain them
	// See: https://pkg.go.dev/github.com/rivo/tview#hdr-Type_Hierarchy
	sl.SetBorder(true)
	sl.SetTitle("Images")

	return sl
}

func getTagsBox(app *tview.Application) *SearchableList {
	sl := NewSearchableList(app)

	sl.SetBorder(true)
	sl.SetTitle("Tags")

	return sl
}

func (e *ExplorerApp) updateTagsBoxFunction(tagsBox *SearchableList) func(int, string, string, rune) {
	return func(_ int, imageName, _ string, _ rune) {
		if imageName == "Quit" {
			return
		}
		tagsBox.Clear()

		// Store the selected image name
		e.selectedImage = imageName

		tags := explorer.GetTags(strings.Join([]string{e.registry, imageName}, "/"))
		for _, tag := range tags {
			tagsBox.AddItem(tag, "", 0, nil)
		}
		tagsBox.SetTitle(fmt.Sprintf("Image %s tags", imageName))
	}
}
