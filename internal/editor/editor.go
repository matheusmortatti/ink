package editor

import (
	tea "charm.land/bubbletea/v2"
	"github.com/matheusmortatti/ink/internal/block"
	"github.com/matheusmortatti/ink/internal/render"
	"github.com/matheusmortatti/ink/internal/ui"
)

// EditorModel is the single Bubbletea model. Components are fields.
type EditorModel struct {
	blocks   []block.Block
	viewport *ui.Viewport
	renderer *render.Renderer
	cache    *render.RenderCache
	filePath string
	ready    bool
	width    int
	height   int
}

// NewEditor creates an EditorModel with the given file path and parsed blocks.
func NewEditor(filePath string, blocks []block.Block) *EditorModel {
	return &EditorModel{
		filePath: filePath,
		blocks:   blocks,
	}
}

func (e *EditorModel) Init() tea.Cmd {
	return nil
}

func (e *EditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
		if !e.ready {
			e.initViewport()
		} else {
			// Resize may fail if re-rendering fails; viewport retains previous layout
			_ = e.viewport.Resize(e.width, e.height)
		}
		return e, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return e, tea.Quit
		case "j":
			if e.ready && e.viewport != nil {
				e.viewport.ScrollDown(1)
			}
		case "k":
			if e.ready && e.viewport != nil {
				e.viewport.ScrollUp(1)
			}
		}
	}

	return e, nil
}

func (e *EditorModel) View() tea.View {
	if !e.ready {
		return tea.NewView("loading...")
	}

	v := tea.NewView(e.viewport.View())
	v.AltScreen = true
	return v
}

// initViewport creates renderer, cache, viewport on first WindowSizeMsg.
func (e *EditorModel) initViewport() {
	colWidth := ui.CalculateColumnWidth(e.width)
	if colWidth <= 0 {
		colWidth = 1
	}

	r, err := render.NewRenderer(colWidth)
	if err != nil {
		e.ready = true
		e.viewport = ui.NewViewport(e.width, e.height)
		return
	}
	e.renderer = r
	e.cache = render.NewRenderCache()

	// PreRenderAll continues on partial failure; non-fatal
	_ = r.PreRenderAll(e.blocks, e.cache)

	e.viewport = ui.NewViewport(e.width, e.height)
	// SetContent may fail if rendering fails; viewport shows empty content
	_ = e.viewport.SetContent(e.blocks, e.renderer, e.cache)

	e.ready = true
}
