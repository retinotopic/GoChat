package list

import (
	"strings"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Content struct {
	MainText      string
	SecondaryText string
}

func NewList(items ListItems, option func(ListItem)) *List {
	return &List{
		Option:      option,
		Items:       items,
		Box:         tview.NewBox().SetBorder(true),
		lines:       make([]string, 0, 1000),
		selectedBuf: make([]Content, 0, 100),
		Current:     nil,
	}
}

type List struct {
	*tview.Box
	lines       []string
	offset      int // scroll offset
	selectedBuf []Content
	Option      func(ListItem)
	Items       ListItems
	Current     ListItem // type Room
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

	l.selectedBuf = l.selectedBuf[:0]
	for front != nil && front.IsNil() {
		cnt := Content{MainText: front.GetMainText(), SecondaryText: front.GetSecondaryText()}
		l.selectedBuf = append(l.selectedBuf, cnt)
		front = front.Next()
	}
	return l.selectedBuf
}
func (l *List) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()
	if l.Current == nil && l.Items.Len() > 0 {
		l.Current = l.Items.GetBack()
	}
	element := l.Items.GetBack()
	for i := 0; i < l.offset && element != nil && !element.IsNil(); i++ {
		element = element.Next()
	}
	row := 0
	for element != nil && !element.IsNil() && row < height {

		mainText := element.GetMainText()
		lines := l.splitTextIntoLines(mainText, width)
		for lineIndex, line := range lines {
			if row+lineIndex >= height {
				break
			}
			tview.Print(screen, line, x+2, y+row+lineIndex, width, tview.AlignLeft, element.GetColor(0))

			if element == l.Current {
				screen.SetContent(x, y+row+lineIndex, '|', nil, tcell.StyleDefault)
			}
		}

		if len(element.GetSecondaryText()) > 0 && width > 3 {
			secondaryLines := l.splitTextIntoLines(element.GetSecondaryText(), width-2)
			startY := row + len(lines)
			for lineIndex, line := range secondaryLines {
				if startY+lineIndex >= height {
					break
				}
				tview.Print(screen, line, x+2, y+startY+lineIndex,
					width, tview.AlignLeft, element.GetColor(1))
				if element == l.Current {
					screen.SetContent(x, y+startY+lineIndex, '|', nil, tcell.StyleDefault)
				}
			}
			row += len(lines) + len(secondaryLines)
		} else {
			row += len(lines)
		}
		element = element.Next()
	}
}
func (l *List) splitTextIntoLines(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	l.lines = l.lines[:0]
	words := strings.Fields(text)
	currentLine := ""
	for _, word := range words {
		wordWidth := utf8.RuneCountInString(word)

		if wordWidth > maxWidth {
			if len(currentLine) > 0 {
				l.lines = append(l.lines, currentLine)
				currentLine = ""
			}

			runes := []rune(word)
			for i := 0; i < len(runes); i += maxWidth {
				end := i + maxWidth
				if end > len(runes) {
					end = len(runes)
				}
				l.lines = append(l.lines, string(runes[i:end]))
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
				l.lines = append(l.lines, currentLine)
			}
			currentLine = word
		}
	}

	if len(currentLine) > 0 {
		l.lines = append(l.lines, currentLine)
	}

	return l.lines
}
func (l *List) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

		switch event.Key() {
		case tcell.KeyUp:
			if l.Current != nil && !l.Current.IsNil() && l.Current.Prev() != nil && !l.Current.Prev().IsNil() {
				if l.Current == l.getFirstVisibleElement() {
					l.offset--
				}
				l.Current = l.Current.Prev()
			}
		case tcell.KeyDown:
			if l.Current != nil && !l.Current.IsNil() && l.Current.Next() != nil && !l.Current.Next().IsNil() {
				if l.isLastVisibleElement(l.Current) {
					l.offset++
				}
				l.Current = l.Current.Next()
			}
		case tcell.KeyEnter:
			if l.Option != nil {
				l.Option(l.Current)
			}
		}
		return
	})
}
func (l *List) getFirstVisibleElement() ListItem {
	element := l.Items.GetBack()
	for i := 0; i < l.offset && element != nil && !element.IsNil(); i++ {
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
