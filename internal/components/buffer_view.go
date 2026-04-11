package components

import (
	"strconv"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/robmerrell/eldritch/internal/buffer"
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
	_, cmd := b.modeline.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width = msg.Width
		b.height = msg.Height

		// case state.MsgModeKeyPress:
		// 	switch msg.Mode {
		// 	case state.InputModeNormal:
		// 	}
	}

	return b, cmd
}

func (b *BufferView) View() tea.View {
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

	startLine := 0
	lineNum := startLine

	var contents strings.Builder
	var lineNums strings.Builder
	for renderableLine := range b.buffer.ContentsForRendering(startLine, contentHeight, contentWidth) {
		contents.WriteString(renderableLine.LineContents + "\n")

		lineNums.WriteString(strconv.Itoa(lineNum + 1))

		if renderableLine.RenderedRows == 1 {
			lineNums.WriteString("\n")
		} else {
			lineNums.WriteString(strings.Repeat("\n", renderableLine.RenderedRows-1))
		}

		lineNum += 1
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
