package components

import (
	"log"
	"slices"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/davecgh/go-spew/spew"
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
			case "h":
				b.buffer.ShiftSelections(buffer.SelectionDirectionLeft, 1)
			case "j":
				b.buffer.ShiftSelections(buffer.SelectionDirectionDown, 1)
			case "k":
				b.buffer.ShiftSelections(buffer.SelectionDirectionUp, 1)
			case "l":
				b.buffer.ShiftSelections(buffer.SelectionDirectionRight, 1)
			}
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
	// lineNum := startLine

	var contents strings.Builder
	var lineNums strings.Builder
	renderableContents := b.buffer.ContentsForRendering(startLine, startLine+contentHeight)

	for i := range renderableContents {
		if len(renderableContents[i]) == 0 {
			renderableContents[i] = " "
		}
	}

	// render selections into the contents. Replace this.
	for _, selection := range b.buffer.Selections() {
		log.Println("=========")
		v := renderableContents[selection.HeadY]
		log.Println(spew.Sdump(v))
		log.Printf("x:%d, y:%d, len %d", selection.HeadX, selection.HeadY, len(v))
		log.Println("=========")
		if int(selection.HeadX) < len(v) {
			line := []rune(renderableContents[selection.HeadY])
			value := []rune(selectionHeadInlineStyle(string(line[selection.HeadX])))
			merged := slices.Replace(line, int(selection.HeadX), 1, value...)
			renderableContents[selection.HeadY] = string(merged)
		}
	}

	for _, line := range renderableContents {
		lineWriter := wrap.NewWriter(contentWidth)
		lineWriter.Write([]byte(line))

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
