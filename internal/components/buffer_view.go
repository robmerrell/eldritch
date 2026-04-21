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
				// b.buffer.Delete()
			case "h":
				b.buffer.ShiftSelections(buffer.SelectionDirectionLeft, 1)
			case "j":
				b.buffer.ShiftSelections(buffer.SelectionDirectionDown, 1)
			case "k":
				b.buffer.ShiftSelections(buffer.SelectionDirectionUp, 1)
			case "l":
				b.buffer.ShiftSelections(buffer.SelectionDirectionRight, 1)
			}
		} else if msg.Mode == state.InputModeInsert {
			b.buffer.Insert([]rune(msg.PressMsg.String())[0])
		}
	}

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
		Foreground(b.theme.Fg).
		Background(b.theme.Bg).
		Width(lineNumWidth).
		Height(contentHeight)

	contentStyle := lipgloss.NewStyle().
		Foreground(b.theme.ModelineInputModeFg).
		Background(b.theme.ModelineInputModeBg).
		Width(contentWidth).
		Height(contentHeight)

	selectionHeadInlineStyle := lipgloss.NewStyle().
		Foreground(b.theme.ModelineInputModeBg).
		Background(b.theme.ModelineInputModeFg).Render
	startLine := 0

	var contents strings.Builder
	var lineNums strings.Builder
	renderableContents := b.buffer.ContentsForRendering(startLine, startLine+contentHeight)

	// add a space on empty lines to make rendering a cursor easier and line ends
	for i := range renderableContents {
		if len(renderableContents[i]) == 0 {
			renderableContents[i] = []rune(" ")
		}

		renderableContents[i] = append(renderableContents[i], []rune(" ")...)
	}

	// render selections into the contents. This is dumb, but keeps be moving forward until I
	// want to optimize it.
	for _, selection := range b.buffer.Selections() {
		line := renderableContents[selection.HeadY]
		headRune := []rune(selectionHeadInlineStyle(string(line[selection.HeadX])))
		merged := append(line[:selection.HeadX], append(headRune, line[selection.HeadX+1:]...)...)
		renderableContents[selection.HeadY] = merged
	}

	for _, line := range renderableContents {
		lineWriter := wrap.NewWriter(contentWidth)
		lineWriter.Write([]byte(string(line)))

		contents.WriteString(lineWriter.String() + "\n")
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
