package ssh

import (
	"fmt"
	"regexp"

	"github.com/AndreZiviani/fzf-wrapper/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// removes only the color customization at the
// beging of the string, if exists
// does NOT remove other customizations
func removeLineColor(list *tview.List, id int) {
	currentText, currentSecondary := list.GetItemText(id)
	re := regexp.MustCompile(`^\[[a-zA-Z0-9:-]+\]`)
	tmp := re.ReplaceAllString(currentText, "${1}")
	list.SetItemText(id, tmp, currentSecondary)

}
func boldItem(list *tview.List, id int) {
	if list.GetItemCount() == 0 {
		return
	}
	currentText, currentSecondary := list.GetItemText(id)
	list.SetItemText(id, fmt.Sprintf("[::b]%s", currentText), currentSecondary)
}

func NewTui() *Tui {
	t := Tui{
		app: tview.NewApplication(),
		resourceDetails: tview.NewTextView().
			SetDynamicColors(true).
			SetRegions(true),
		resourceList: tview.NewList().
			ShowSecondaryText(false).
			SetSelectedBackgroundColor(tcell.ColorDarkSlateGray).
			SetSelectedTextColor(tcell.ColorWhite).
			SetMainTextColor(tcell.ColorDarkGray).
			SetWrapAround(true),
		input: tview.NewInputField().
			SetLabel(">: "),
		flex: tview.NewFlex(),
		fzf:  fzfwrapper.NewWrapper(fzfwrapper.WithSortBy(fzfwrapper.ByScore, fzfwrapper.ByPosition, fzfwrapper.ByLength)),
	}

	t.app.EnableMouse(true)
	t.resourceDetails.SetBorder(true)
	t.resourceList.SetBorder(true)

	/*
		t.resourceList.SetChangedFunc(func(id int, text string, secondary string, shortcut rune) {
			t.resourceDetails.SetText(
				fmt.Sprintf("%s\n", secondary),
			)
		})
	*/
	t.resourceList.SetChangedFunc(t.resourceListFunc)

	/*
		t.input.SetChangedFunc(func(text string) {
			if text == "" {
				t.resourceList.Clear()
				last := len(t.instances) - 1
				for k, v := range t.instances {
					t.instanceIdx[last-k] = k
					t.resourceList.InsertItem(
						-t.resourceList.GetItemCount()-1,
						v.PrintName(),
						v.PrintDetails(),
						0, nil,
					)
				}
				return
			}

			t.fzf.SetPattern(text)
			results, _ := t.fzf.Fuzzy()

			t.resourceList.Clear()
			t.resourceDetails.Clear()

			last := len(results) - 1
			for k, v := range results {
				t.instanceIdx[last-k] = int(v.Item.Index())
				i := t.instances[v.Item.Index()]
				t.resourceList.InsertItem(
					-t.resourceList.GetItemCount()-1,
					tview.TranslateANSI(
						i.PrintName(),
					),
					tview.TranslateANSI(v.HighlightResult()),
					0, nil,
				)
			}

			t.resourceList.SetCurrentItem(-1)
			t.resourceList.SetOffset(0, 0)
			boldItem(t.resourceList, t.resourceList.GetCurrentItem())
		})
	*/
	t.input.SetChangedFunc(t.inputFunc)

	t.flex.SetDirection(tview.FlexRow).
		// Horizontal view, textView
		AddItem(tview.NewFlex().
			// Vertical view, options | details
			AddItem(t.resourceList, 0, 1, false).
			AddItem(t.resourceDetails, 0, 1, false),
			0, 1, false).
		// Horizontal view, input field
		AddItem(t.input, 1, 1, true)

	t.setCaptureEvents()
	return &t
}

func (t *Tui) setCaptureEvents() {
	// Capture key events to perform custom actions
	// Configure TAB key to cycle between windows
	// Configure Up/Down key in input screen to scroll the list
	t.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		k := event.Key()
		where := t.app.GetFocus()
		switch k {
		case tcell.KeyEnter:
			current := t.resourceList.GetCurrentItem() // current index selected from list
			instanceIdx := t.instanceIdx[current]      // offset of instances list
			t.selected = &t.instances[instanceIdx]
			t.app.Stop()
			return nil
		case tcell.KeyTab:
			switch where {
			case t.resourceDetails:
				// next window
				t.app.SetFocus(t.input)
				return nil
			case t.resourceList:
				// next window
				t.app.SetFocus(t.resourceDetails)
				return nil
			case t.input:
				// next window
				t.app.SetFocus(t.resourceList)
				return nil
			}
		case tcell.KeyBacktab:
			switch where {
			case t.resourceDetails:
				// previous window
				t.app.SetFocus(t.resourceList)
				return nil
			case t.resourceList:
				// previous window
				t.app.SetFocus(t.input)
				return nil
			case t.input:
				// previous window
				t.app.SetFocus(t.resourceDetails)
				return nil
			}
		case tcell.KeyUp:
			switch where {
			case t.input, t.resourceList:
				// list up
				current := t.resourceList.GetCurrentItem()
				previous := current - 1
				if previous < 0 {
					previous = t.resourceList.GetItemCount() - 1
				}
				removeLineColor(t.resourceList, current)
				boldItem(t.resourceList, previous)
				t.resourceList.SetCurrentItem(previous)
				return nil
			}
		case tcell.KeyDown:
			switch where {
			case t.input, t.resourceList:
				// list down
				current := t.resourceList.GetCurrentItem()
				next := (current + 1) % t.resourceList.GetItemCount()
				removeLineColor(t.resourceList, current)
				boldItem(t.resourceList, next)
				t.resourceList.SetCurrentItem(next)
				return nil
			}
		}
		return event
	})
}
