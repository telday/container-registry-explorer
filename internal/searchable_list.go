package internal

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// SearchableListConfig holds configuration for the SearchableList primitive.
type SearchableListConfig struct {
	// DebounceInterval is the delay after the last keystroke before filtering
	// is applied. This prevents excessive filtering while the user is still typing.
	DebounceInterval time.Duration
}

// DefaultSearchableListConfig returns the default configuration.
func DefaultSearchableListConfig() SearchableListConfig {
	return SearchableListConfig{
		DebounceInterval: 300 * time.Millisecond,
	}
}

// searchableListItem represents a single item stored in the list.
type searchableListItem struct {
	mainText      string
	secondaryText string
	shortcut      rune
	selected      func()
}

// SearchableList is a composite tview primitive that combines a List with
// a search InputField. Pressing `/` activates the search input; typing
// filters the list with a case-insensitive substring match after a short
// debounce. Enter confirms the search and returns focus to the list.
// Escape exits the input and returns focus to the list.
type SearchableList struct {
	*tview.Flex

	list   *tview.List
	input  *tview.InputField
	config SearchableListConfig
	app    *tview.Application

	// allItems is the full unfiltered set of items.
	allItems []searchableListItem

	// debounceTimer is used to delay filtering until the user pauses typing.
	debounceTimer *time.Timer

	// searching is true when the input field is visible and focused.
	searching bool
}

// NewSearchableList creates a new SearchableList with the given application
// reference and an optional configuration. If no config is provided the
// default is used.
func NewSearchableList(app *tview.Application, config ...SearchableListConfig) *SearchableList {
	cfg := DefaultSearchableListConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	list := tview.NewList().ShowSecondaryText(false)
	input := tview.NewInputField()
	input.SetLabel("/")
	input.SetFieldWidth(0)
	input.SetBorder(true)

	sl := &SearchableList{
		Flex:   tview.NewFlex().SetDirection(tview.FlexRow),
		list:   list,
		input:  input,
		config: cfg,
		app:    app,
	}

	// Start with only the list visible (no search bar).
	sl.rebuildLayout()

	// When the user presses `/` in the list, open the search input.
	sl.list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == '/' {
			sl.openSearch()
			return nil
		}
		return event
	})

	// Input field key handling.
	sl.input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			sl.confirmSearch()
			return nil
		case tcell.KeyEscape:
			sl.confirmSearch()
			return nil
		}
		return event
	})

	// Trigger debounced filtering whenever the input text changes.
	sl.input.SetChangedFunc(func(text string) {
		sl.scheduleFilter(text)
	})

	return sl
}

// openSearch shows the input field and focuses it.
func (sl *SearchableList) openSearch() {
	sl.searching = true
	sl.input.SetText("")
	sl.rebuildLayout()
	sl.app.SetFocus(sl.input)
}

// confirmSearch hides the input (if empty) and returns focus to the list.
func (sl *SearchableList) confirmSearch() {
	sl.searching = false
	text := sl.getSearchText()
	if text == "" {
		// No active filter — restore full list and hide input.
		sl.applyFilter("")
		sl.rebuildLayout()
	}
	sl.app.SetFocus(sl.list)
}

// rebuildLayout reconstructs the flex layout based on whether the search
// input should be visible.
func (sl *SearchableList) rebuildLayout() {
	sl.Flex.Clear()
	if sl.searching || sl.getSearchText() != "" {
		sl.Flex.AddItem(sl.input, 3, 0, sl.searching)
	}
	sl.Flex.AddItem(sl.list, 0, 1, !sl.searching)
}

// getSearchText returns the current search query with the leading `/` stripped.
func (sl *SearchableList) getSearchText() string {
	return sl.input.GetText()
}

// scheduleFilter resets the debounce timer. When it fires, the list is
// filtered by the current search text.
func (sl *SearchableList) scheduleFilter(text string) {
	if sl.debounceTimer != nil {
		sl.debounceTimer.Stop()
	}

	sl.debounceTimer = time.AfterFunc(sl.config.DebounceInterval, func() {
		sl.app.QueueUpdateDraw(func() {
			sl.applyFilter(text)
		})
	})
}

// applyFilter updates the visible list items based on the search query.
func (sl *SearchableList) applyFilter(query string) {
	sl.list.Clear()
	query = strings.ToLower(query)

	for _, item := range sl.allItems {
		if query == "" || strings.Contains(strings.ToLower(item.mainText), query) {
			sl.list.AddItem(item.mainText, item.secondaryText, item.shortcut, item.selected)
		}
	}
}

// --- Public API that mirrors tview.List ---

// AddItem adds an item to the searchable list, storing it in the backing
// slice so it can be restored after filtering.
func (sl *SearchableList) AddItem(mainText, secondaryText string, shortcut rune, selected func()) *SearchableList {
	sl.allItems = append(sl.allItems, searchableListItem{
		mainText:      mainText,
		secondaryText: secondaryText,
		shortcut:      shortcut,
		selected:      selected,
	})
	sl.list.AddItem(mainText, secondaryText, shortcut, selected)
	return sl
}

// Clear removes all items from the list and the backing store, and resets
// the search input.
func (sl *SearchableList) Clear() *SearchableList {
	sl.allItems = nil
	sl.list.Clear()
	sl.input.SetText("")
	sl.searching = false
	sl.rebuildLayout()
	return sl
}

// SetSelectedFunc sets the handler called when the user selects an item.
func (sl *SearchableList) SetSelectedFunc(handler func(int, string, string, rune)) *SearchableList {
	sl.list.SetSelectedFunc(handler)
	return sl
}

// SetChangedFunc sets the handler called when the highlighted item changes.
func (sl *SearchableList) SetChangedFunc(handler func(int, string, string, rune)) *SearchableList {
	sl.list.SetChangedFunc(handler)
	return sl
}

// GetList returns the underlying tview.List for advanced usage.
func (sl *SearchableList) GetList() *tview.List {
	return sl.list
}

// SetTitle sets the border title on the underlying list.
func (sl *SearchableList) SetTitle(title string) *SearchableList {
	sl.list.SetTitle(title)
	return sl
}

// SetBorder sets the border on the underlying list.
func (sl *SearchableList) SetBorder(show bool) *SearchableList {
	sl.list.SetBorder(show)
	return sl
}

// HasFocus returns true if either the list or the input has focus.
func (sl *SearchableList) HasFocus() bool {
	return sl.list.HasFocus() || sl.input.HasFocus()
}

// GetItemCount returns the number of items currently visible in the list.
func (sl *SearchableList) GetItemCount() int {
	return sl.list.GetItemCount()
}
