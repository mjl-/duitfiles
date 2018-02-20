/*
Package duitfiles provides a file picker for use with duit, the developer ui toolkit.
*/
package duitfiles

import (
	"fmt"
	"image"
	"io/ioutil"
	"os"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

// Favorites on the left, fixed size. Remainder on the right contains one or more listboxes.
// Favorites is populated with some names that point to dirs. Clicking makes the favorite active, and focuses on first column.
// Typing then filters only the matching elements.  We just show text. Names ending in "/" are directories.
// Hitting tab on a directory opens that dir, and moves focus there.
// Hitting "enter" on a file causes it to be plumbed (opened).

type files struct {
	dui *duit.DUI

	selected     chan string
	contentUI    duit.UI // contains either the columns with directories/files, or errorUI with an error message
	colsUI       *columnsUI
	errorLabel   *duit.Label
	errorClear   *duit.Button
	pathLabel    *duit.Label // at top of the window
	selectButton *duit.Button
	favUI        *favoritesUI
	bold         *draw.Font
}

func (f *files) error(err error, msg string) bool {
	if err == nil {
		return false
	}
	f.errorLabel.Text = fmt.Sprintf("%s: %s", msg, err)
	f.dui.MarkLayout(nil)
	f.dui.Focus(f.errorClear)
	return true
}

func (f *files) clearError() {
	if f.errorLabel.Text == "" {
		return
	}
	f.errorLabel.Text = ""
	f.dui.MarkLayout(f.contentUI)
}

func (f *files) listDir(path string) []string {
	l, err := ioutil.ReadDir(path)
	if f.error(err, "readdir") {
		return []string{}
	}
	names := make([]string, len(l))
	for i, fi := range l {
		names[i] = fi.Name()
		if fi.IsDir() {
			names[i] += "/"
		}
	}
	return names
}

// Select creates a new window with the file picker and returns the selected filename or an error.
//
// The window shows favoites on the left. The user's home directory
// and the file system root are always listed. Along with any entries from
// $appdata/duitfiles/favorites (each line is a directory listed in
// the favorites). The +/- button adds/removes a favorite from the
// list.
//
// On the right side, you'll see at least one column with the files
// in the directory, the selected favorite. Typing in the search box
// above the files filters by substring. Hitting ctrl-f completes
// the text, making it the longest common prefix of the files still
// visible. After having typed/completed/clicked an exact match for
// a file, that file is selected. For directories, this opens a new
// columns and warps the pointer to its search box.
//
// The currently selected path is shown at the top.
// Hitting the "select"-button, or return, returns the selected path.
func Select() (filename string, err error) {
	f := &files{
		selected: make(chan string),
	}

	dui, err := duit.NewDUI("files", nil)
	if err != nil {
		return "", fmt.Errorf("new DUI: %s", err)
	}
	f.dui = dui

	favorites, err := loadFavorites()
	if err != nil {
		return "", fmt.Errorf("loading favorites: %s", err)
	}

	f.favUI = newFavoritesUI(f, favorites)
	f.pathLabel = &duit.Label{Text: f.favUI.list.Values[0].Value.(string)}
	f.selectButton = &duit.Button{
		Text: "select",
		Click: func() (e duit.Event) {
			p := f.pathLabel.Text
			go func() {
				f.selected <- p
			}()
			return
		},
	}

	f.errorLabel = &duit.Label{}
	f.errorClear = &duit.Button{
		Text:     "clear",
		Colorset: &dui.Primary,
		Click: func() (e duit.Event) {
			f.clearError()
			return
		},
	}
	errorUI := duit.NewMiddle(duit.SpaceXY(duit.ScrollbarSize, duit.ScrollbarSize), &duit.Box{
		Margin: image.Pt(6, 4),
		Kids:   duit.NewKids(f.errorLabel, f.errorClear),
	})

	f.colsUI = &columnsUI{
		files: f,
		Split: duit.Split{
			Background: dui.Gutter,
			Gutter:     1,
			Split: func(width int) []int {
				widths := make([]int, len(f.colsUI.Kids))
				col := width / len(widths)
				for i := range widths {
					widths[i] = col
				}
				widths[len(widths)-1] = width - col*(len(widths)-1)
				return widths
			},
		},
	}
	f.colsUI.Split.Kids = duit.NewKids(newColumnUI(f, 0, "", f.listDir(f.pathLabel.Text)))

	bold, _ := dui.Display.OpenFont(os.Getenv("fontbold"))
	f.bold = bold

	f.contentUI = &duit.Pick{
		Pick: func(_ image.Point) duit.UI {
			if f.errorLabel.Text != "" {
				return errorUI
			}
			return f.colsUI
		},
	}

	dui.Top.UI = &duit.Split{
		Gutter:     1,
		Background: dui.Gutter,
		Split: func(width int) []int {
			return []int{dui.Scale(200), width - dui.Scale(200)}
		},
		Kids: duit.NewKids(
			f.favUI,
			&duit.Box{
				Height: -1,
				Valign: duit.ValignMiddle,
				Kids: duit.NewKids(
					&duit.Box{
						Padding: duit.Space{Left: duit.ScrollbarSize, Top: 4, Bottom: 4},
						Margin:  image.Pt(6, 4),
						Kids:    duit.NewKids(f.pathLabel, f.selectButton),
					},
					f.contentUI,
				),
			},
		),
	}
	dui.Top.ID = "favorites"
	dui.Render()
	dui.Focus(f.colsUI.Kids[0].UI.(*columnUI).field)

	for {
		select {
		case e := <-dui.Inputs:
			dui.Input(e)

		case xerr, ok := <-dui.Error:
			if !ok {
				return
			}
			dui.Close()
			return "", xerr

		case filename = <-f.selected:
			dui.Close()
			return
		}
	}
}
