package duitfiles

import (
	"strings"

	"9fans.net/go/draw"
	"github.com/mjl-/duit"
)

type columnUI struct {
	files *files
	name  string   // File/dir element of this column, not full path. Directories always end with a slash.
	names []string // File/dir names. Dir names always end with a slash.
	field *duit.Field
	list  *duit.List
	duit.Box
}

func newColumnUI(files *files, colIndex int, name string, names []string) (ui *columnUI) {
	ui = &columnUI{
		files: files,
		name:  name,
		names: names,
	}
	dui := ui.files.dui
	colsUI := ui.files.colsUI

	l := make([]*duit.ListValue, len(names))
	for i, name := range names {
		l[i] = &duit.ListValue{Text: name, Value: name}
	}
	ui.list = &duit.List{
		Values: l,
		Changed: func(index int) (e duit.Event) {
			ui.files.clearError()
			if ui.list.Values[index].Selected {
				colsUI.selectName(colIndex, ui.list.Values[index].Value.(string))
			} else {
				colsUI.selectName(colIndex, "")
			}
			return
		},
		Keys: func(k rune, m draw.Mouse) (e duit.Event) {
			ui.files.clearError()
			switch k {
			case '\n':
				sel := ui.list.Selected()
				if len(sel) != 1 {
					return
				}
				index := sel[0]
				e.Consumed = true
				path := colsUI.composePath(colIndex, ui.list.Values[index].Value.(string))
				go func() {
					files.selected <- path
				}()
			case draw.KeyLeft:
				e.Consumed = true
				if colIndex > 0 {
					colsUI.selectName(colIndex-1, "")
				} else {
					colsUI.selectName(colIndex, "")
				}
			case draw.KeyRight:
				sel := ui.list.Selected()
				if len(sel) != 1 {
					return
				}
				index := sel[0]
				elem := ui.list.Values[index].Value.(string)
				if strings.HasSuffix(elem, "/") {
					e.Consumed = true
					colsUI.selectName(colIndex, elem)
					colNames := colsUI.Kids[colIndex+1].UI.(*columnUI).names
					if len(colNames) > 0 {
						colsUI.selectName(colIndex+1, colNames[0])
					}
					dui.Focus(colsUI.Kids[len(colsUI.Kids)-1].UI.(*columnUI).list)
				}
			}
			return
		},
	}
	ui.field = &duit.Field{
		Keys: func(k rune, m draw.Mouse) (e duit.Event) {
			ui.files.clearError()
			switch k {
			case 'f' & 0x1f:
				// Completion. The list only contains values that has a substring match.
				e.Consumed = true
				var s string
				if len(ui.list.Values) == 1 {
					s = ui.list.Values[0].Text
				} else {
					for _, lv := range ui.list.Values {
						if !strings.HasPrefix(lv.Text, ui.field.Text) {
							continue
						}
						if s == "" {
							s = lv.Text
							continue
						}
						for i, c := range []byte(lv.Text) {
							if i >= len(s) || s[i] != c {
								s = s[:i]
								break
							}
						}
					}
				}
				ui.field.Text = s
				ui.field.Cursor1 = 0
				ui.field.Changed(ui.field.Text)
				e.NeedDraw = true
			}
			return
		},
		Changed: func(newValue string) (e duit.Event) {
			ui.files.clearError()
			nl := []*duit.ListValue{}
			exactMatch := false
			for _, name := range names {
				exactMatch = exactMatch || name == newValue
				if strings.Contains(name, newValue) {
					nl = append(nl, &duit.ListValue{Text: name, Value: name})
				}
			}
			ui.list.Values = nl
			if exactMatch {
				colsUI.selectName(colIndex, newValue)
				dui.Focus(colsUI.Kids[len(colsUI.Kids)-1].UI.(*columnUI).field)
			}
			e.NeedDraw = true
			dui.MarkLayout(ui.list)
			return
		},
	}

	ui.Box = duit.Box{
		Kids: duit.NewKids(
			&duit.Box{
				Padding: duit.SpaceXY(6, 4),
				Kids:    duit.NewKids(ui.field),
			},
			duit.NewScroll(ui.list),
		),
	}
	return
}
