package list

import (
	"log"
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Content struct {
	MainText      string
	SecondaryText string
}

func NewList(items ListItems, option func(ListItem), title string, log *log.Logger) *List {
	box := tview.NewBox().SetBorder(true)
	box.SetTitle(title)
	box.SetBackgroundColor(tcell.ColorDarkSlateGray)
	box.SetBorderColor(tcell.ColorSlateGray)
	box.SetFocusFunc(func() {
		box.SetBorderColor(tcell.ColorWhite)
	})
	box.SetBlurFunc(func() {
		box.SetBorderColor(tcell.ColorSlateGray)
	})
	return &List{
		Option:      option,
		Items:       items,
		Box:         box,
		Lines:       make([]string, 0, 1000),
		SelectedBuf: make([]Content, 0, 100),
		Current:     nil,
		Style:       tcell.Style{}.Foreground(tcell.ColorAquaMarine).Background(tcell.ColorMidnightBlue),
		Logger:      log,
	}
}

type List struct {
	*tview.Box
	Logger      *log.Logger
	Lines       []string
	Offset      int // scroll offset
	SelectedBuf []Content
	Style       tcell.Style
	Option      func(ListItem)
	Items       ListItems
	Current     ListItem
}

type ListItem interface {
	GetParent() ListItems
	GetMainText() string
	GetSecondaryText() string
	GetColor(int) tcell.Color
	SetMainText(string, uint8) /* can be: default string override,
	concatenate string via bytes.Buffer,
	delete last (len(string)) bytes */
	SetSecondaryText(string)
	SetColor(tcell.Color, int)
	Next() ListItem
	Prev() ListItem
	IsNil() bool
}

type ListItems interface {
	NewItem([2]tcell.Color, string, string) ListItem
	MoveToFront(ListItem)
	MoveToBack(ListItem)
	GetFront() ListItem
	GetBack() ListItem
	Remove(ListItem)
	Clear()
	Len() int
}

func (l *List) GetSelected() []Content {
	front := l.Items.GetBack()

	l.SelectedBuf = l.SelectedBuf[:0]
	for front != nil && !front.IsNil() {
		cnt := Content{MainText: front.GetMainText(), SecondaryText: front.GetSecondaryText()}
		if front.GetColor(0) == tcell.ColorRed {
			l.SelectedBuf = append(l.SelectedBuf, cnt)
		}
		front = front.Next()
	}
	return l.SelectedBuf
}

func (l *List) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()

	element := l.Items.GetBack()
	for i := 0; i < l.Offset && element != nil && !element.IsNil(); i++ {
		element = element.Next()
	}
	row := 0
	for element != nil && !element.IsNil() && row < height {
		mainText := element.GetMainText()
		Lines := l.splitTextIntoLines(mainText, width)
		for lineIndex, line := range Lines {
			if row+lineIndex >= height {
				break
			}
			tview.Print(screen, line, x, y+row+lineIndex, width, tview.AlignLeft, element.GetColor(0))
			if element == l.Current {
				for i, r := range []rune(line) {
					screen.SetContent(x+i, y+row+lineIndex, r, nil, l.Style)
				}
			}
		}

		if len(element.GetSecondaryText()) > 0 && width > 3 {
			secondaryLines := l.splitTextIntoLines(element.GetSecondaryText(), width-2)
			startY := row + len(Lines)
			for lineIndex, line := range secondaryLines {
				if startY+lineIndex >= height {
					break
				}
				tview.Print(screen, line, x, y+startY+lineIndex,
					width, tview.AlignLeft, element.GetColor(1))
				if element == l.Current {
					for i, r := range []rune(line) {
						screen.SetContent(x+i, y+startY+lineIndex, r, nil, l.Style)
					}
				}
			}
			row += len(Lines) + len(secondaryLines)
		} else {
			row += len(Lines)
		}
		element = element.Next()
	}
}

func (l *List) splitTextIntoLines(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	if l.Current == nil && l.Items.Len() > 0 {
		l.Current = l.Items.GetBack()
	}
	l.Lines = l.Lines[:0]
	words := strings.Fields(text)
	currentLine := ""
	for _, word := range words {
		wordWidth := utf8.RuneCountInString(word)

		if wordWidth > maxWidth {
			if len(currentLine) > 0 {
				l.Lines = append(l.Lines, currentLine)
				currentLine = ""
			}

			runes := []rune(word)
			for i := 0; i < len(runes); i += maxWidth {
				end := i + maxWidth
				if end > len(runes) {
					end = len(runes)
				}
				l.Lines = append(l.Lines, string(runes[i:end]))
			}
			continue
		}
		currentLineWidth := utf8.RuneCountInString(currentLine)
		separatorWidth := 0
		if len(currentLine) > 0 {
			separatorWidth = 1
		}
		if currentLineWidth+separatorWidth+wordWidth <= maxWidth {
			if len(currentLine) > 0 {
				currentLine += " "
			}
			currentLine += word
		} else {
			if len(currentLine) > 0 {
				l.Lines = append(l.Lines, currentLine)
			}
			currentLine = word
		}
	}

	if len(currentLine) > 0 {
		l.Lines = append(l.Lines, currentLine)
	}

	return l.Lines
}
func (l *List) InputHandlerRaw(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	if l.Current == nil && l.Items.Len() > 0 {
		l.Current = l.Items.GetBack()
	}
	if l.Logger != nil {
		i1 := int16(event.Key())
		if i1 == 256 { // rune
			l.Logger.Println("Run:", event.Rune())
		} else { // key
			l.Logger.Println("Key:", tcell.Key(i1))
		}
	}
	switch event.Key() {
	case tcell.KeyUp:
		if l.Current != nil && !l.Current.IsNil() && l.Current.Prev() != nil && !l.Current.Prev().IsNil() {
			if l.Current == l.getFirstVisibleElement() {
				l.Offset--
			}
			l.Current = l.Current.Prev()
		}
	case tcell.KeyDown:
		if l.Current != nil && !l.Current.IsNil() && l.Current.Next() != nil && !l.Current.Next().IsNil() {
			if l.isLastVisibleElement(l.Current) {
				l.Offset++
			}
			l.Current = l.Current.Next()
		}
	case tcell.KeyEnter:
		if l.Option != nil {
			l.Option(l.Current)
		}
	}
	return
}

func (l *List) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.InputHandlerRaw

}
func (l *List) getFirstVisibleElement() ListItem {
	element := l.Items.GetBack()
	for i := 0; i < l.Offset && element != nil && !element.IsNil(); i++ {
		element = element.Next()
	}
	return element
}

func (l *List) isLastVisibleElement(item ListItem) bool {
	_, _, width, height := l.GetInnerRect()

	element := l.getFirstVisibleElement()
	currentHeight := 0

	for element != nil && !element.IsNil() {
		elementHeight := len(l.splitTextIntoLines(element.GetMainText(), width))
		if len(element.GetSecondaryText()) > 0 {
			elementHeight += len(l.splitTextIntoLines(element.GetSecondaryText(), width-2))
		}

		if element == item {
			nextElementHeight := 0
			if element.Next() != nil && !element.Next().IsNil() {
				nextElementHeight = len(l.splitTextIntoLines(element.Next().GetMainText(), width))
				if len(element.Next().GetSecondaryText()) > 0 {
					nextElementHeight += len(l.splitTextIntoLines(element.Next().GetSecondaryText(), width-2))
				}
			}
			return currentHeight+elementHeight+nextElementHeight > height
		}

		currentHeight += elementHeight
		if currentHeight > height {
			return false
		}

		element = element.Next()
	}

	return false
}
