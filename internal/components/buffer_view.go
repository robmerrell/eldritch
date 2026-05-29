package components

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wrap"
	"github.com/robmerrell/eldritch/internal/buffer"
	"github.com/robmerrell/eldritch/internal/state"
	"github.com/robmerrell/eldritch/internal/themes"
)

type BufferView struct {
	buffer   *buffer.Buffer
	theme    *themes.Theme
	modeline *Modeline
	width    int
	height   int
}

func (b *BufferView) Init() tea.Cmd {
	return nil
}

func (b *BufferView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height

	case state.MsgModeKeyPress:
		if msg.Mode == state.InputModeNormal {
			switch msg.PressMsg.String() {
			case "d":
				b.buffer.Delete()
			case "h":
				b.buffer.ShiftSelectionsBackward(1, true)
			case "H":
				b.buffer.ShiftSelectionsBackward(1, false)
			case "j":
				b.buffer.ShiftSelectionsDown(1, true)
			case "J":
				b.buffer.ShiftSelectionsDown(1, false)
			case "k":
				b.buffer.ShiftSelectionsUp(1, true)
			case "K":
				b.buffer.ShiftSelectionsUp(1, false)
			case "l":
				b.buffer.ShiftSelectionsForward(1, true)
			case "L":
				b.buffer.ShiftSelectionsForward(1, false)
			case "x":
				b.buffer.SelectLine()
			case "w":
				b.buffer.Insert([]rune("It's a dangerous\n"))
				b.buffer.Insert([]rune("business, Frodo,\n"))
				b.buffer.Insert([]rune("going out your door\n"))
			}
		} else if msg.Mode == state.InputModeInsert {
			switch msg.PressMsg.String() {
			case "enter":
				b.buffer.Insert([]rune("\n"))

			case "backspace":

			default:
				if msg.PressMsg.Text != "" {
					b.buffer.Insert([]rune(msg.PressMsg.Text))
				}
			}
		}
	}

	b.buffer.LogSelections()

	_, cmd := b.modeline.Update(msg)
	return b, cmd
}

func (b *BufferView) View() tea.View {
	// if we don't have a screen width/height yet don't render anything
	if b.height < 1 || b.width < 1 {
		return tea.NewView("")
	}

	// just hardcode this for now 3 nums + space
	lineNumWidth := 4

	contentHeight := b.height - 1
	contentWidth := b.width - lineNumWidth

	lineNumStyle := lipgloss.NewStyle().
		Foreground(b.theme.ModelineFg).
		Background(b.theme.ModelineBg).
		Width(lineNumWidth).
		Height(contentHeight)

	contentStyle := lipgloss.NewStyle().
		Foreground(b.theme.Fg).
		Background(b.theme.Bg).
		Width(contentWidth).
		Height(contentHeight)

	startLine := 0

	var contents strings.Builder
	var lineNums strings.Builder

	defaultStyle := lipgloss.NewStyle().
		Foreground(b.theme.Fg).
		Background(b.theme.Bg).Render

	selectionHeadStyle := lipgloss.NewStyle().
		Foreground(b.theme.SelectionHeadFg).
		Background(b.theme.SelectionHeadBg).Render

	selectionTailStyle := lipgloss.NewStyle().
		Background(b.theme.SelectionTailBg).Render

	for line := range b.buffer.Contents.Lines(startLine, startLine+contentHeight) {
		var strLine strings.Builder

		for i, rn := range line.Runes {
			// render a space with newlines so we can show selection
			switch b.buffer.OffsetAttribute(line.StartOffset + i) {
			case "selection_tail":
				if rn == '\n' {
					strLine.WriteString(selectionTailStyle(" "))
					strLine.WriteString(defaultStyle("\n"))
				} else {
					strLine.WriteString(selectionTailStyle(string(rn)))
				}

			case "selection_head":
				if rn == '\n' {
					strLine.WriteString(selectionHeadStyle(" "))
					strLine.WriteString(defaultStyle("\n"))
				} else {
					strLine.WriteString(selectionHeadStyle(string(rn)))
				}

			default:
				strLine.WriteString(defaultStyle(string(rn)))
			}
		}

		lineWriter := wrap.NewWriter(contentWidth)
		lineWriter.Write([]byte(strLine.String()))

		contents.WriteString(lineWriter.String())
	}

	layout := lipgloss.JoinVertical(
		lipgloss.Left,
		b.modeline.View().Content,
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			lineNumStyle.Render(lineNums.String()),
			contentStyle.Render(contents.String()),
		),
	)

	return tea.NewView(layout)
}

func NewBufferView(buffer *buffer.Buffer, theme *themes.Theme) *BufferView {
	modeline := NewModeline(theme)

	return &BufferView{
		buffer:   buffer,
		theme:    theme,
		modeline: modeline,
	}
}
