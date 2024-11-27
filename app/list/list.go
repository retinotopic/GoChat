package list

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type List struct {
	*tview.Box
	offset       int // scroll offset
	SelectedFunc func(ListItem)
	Items        ListItems
	Current      ListItem // type Room
}
type ListItem interface {
	GetMainText() string
	GetSecondaryText() string
	GetColor() tcell.Color
	SetMainText(string)
	SetSecondaryText(string)
	SetColor(tcell.Color)
	Next() ListItem
	Prev() ListItem
	IsNil() bool
}
type ListItems interface {
	MoveToFront(ListItem)
	MoveToBack(ListItem)
	GetFront() ListItem
	Remove(ListItem)
	Clear()
	Len() int
}

func (l *List) SetSelectedFunc(handler func(ListItem)) *List {
	l.SelectedFunc = handler
	return l
}

func (l *List) Draw(screen tcell.Screen) {
	l.Box.DrawForSubclass(screen, l)
	x, y, width, height := l.GetInnerRect()
	x = x + 2
	element := l.Items.GetFront()
	for i := 0; i < l.offset && element != nil && !element.IsNil(); i++ {
		element = element.Next()
	}

	row := 0
	for element != nil && !element.IsNil() && row < height {
		mainText := element.GetMainText()
		lines := splitTextIntoLines(mainText, width)
		for lineIndex, line := range lines {
			if row+lineIndex >= height {
				break
			}
			tview.Print(screen, line, x, y+row+lineIndex, width, tview.AlignLeft, element.GetColor())
		}

		if len(element.GetSecondaryText()) > 0 && width > 3 {
			secondaryLines := splitTextIntoLines(element.GetSecondaryText(), width-2)
			startY := row + len(lines)
			for lineIndex, line := range secondaryLines {
				if startY+lineIndex >= height {
					break
				}
				tview.Print(screen, line, x, y+startY+lineIndex,
					width, tview.AlignLeft, tcell.ColorGray)
			}
			row += len(lines) + len(secondaryLines)
		} else {
			row += len(lines)
		}

		if element == l.Current {
			screen.SetContent(x-2, y+row-1, '>', nil, tcell.StyleDefault)
		}
		element = element.Next()
	}
}
func splitTextIntoLines(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	var lines []string
	words := strings.Fields(text)
	currentLine := ""
	for _, word := range words {
		if len(word) > maxWidth {
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
				currentLine = ""
			}

			for i := 0; i < len(word); i += maxWidth {
				end := i + maxWidth
				if end > len(word) {
					end = len(word)
				}
				lines = append(lines, word[i:end])
			}
			continue
		}
		if len(currentLine)+len(word)+1 <= maxWidth {
			if len(currentLine) > 0 {
				currentLine += " "
			}
			currentLine += word
		} else {
			if len(currentLine) > 0 {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if len(currentLine) > 0 {
		lines = append(lines, currentLine)
	}

	return lines
}
func (l *List) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return l.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {

		switch event.Key() {
		case tcell.KeyUp:
			if l.Current.Prev() != nil && !l.Current.Prev().IsNil() {
				if l.Current == l.getFirstVisibleElement() {
					l.offset--
				}
				l.Current = l.Current.Prev()
			}
		case tcell.KeyDown:
			if l.Current.Next() != nil && !l.Current.Next().IsNil() {
				if l.isLastVisibleElement(l.Current) {
					l.offset++
				}
				l.Current = l.Current.Next()
			}
		case tcell.KeyEnter:
			l.SelectedFunc(l.Current)
		}
	})
}
func (l *List) getFirstVisibleElement() ListItem {
	element := l.Items.GetFront()
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
		elementHeight := len(splitTextIntoLines(element.GetMainText(), width))
		if len(element.GetSecondaryText()) > 0 {
			elementHeight += len(splitTextIntoLines(element.GetSecondaryText(), width-2))
		}

		if element == item {
			nextElementHeight := 0
			if element.Next() != nil && !element.Next().IsNil() {
				nextElementHeight = len(splitTextIntoLines(element.Next().GetMainText(), width))
				if len(element.Next().GetSecondaryText()) > 0 {
					nextElementHeight += len(splitTextIntoLines(element.Next().GetSecondaryText(), width-2))
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
