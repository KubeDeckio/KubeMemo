package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/model"
	"github.com/KubeDeckio/KubeMemo/internal/render"
	"github.com/KubeDeckio/KubeMemo/internal/service"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Options struct {
	IncludeRuntime     bool
	RuntimeNamespace   string
	Namespaces         []string
	AutoRefreshSeconds int
}

type modelState struct {
	service         *service.Service
	opts            Options
	notes           []model.Note
	table           table.Model
	viewport        viewport.Model
	filterInput     textinput.Model
	filterMode      string
	kindFilter      string
	namespaceFilter string
	textFilter      string
	width           int
	height          int
	err             error
}

type refreshMsg struct{}

const wideBanner = `
██╗  ██╗██╗   ██╗██████╗ ███████╗███╗   ███╗███████╗███╗   ███╗ ██████╗
██║ ██╔╝██║   ██║██╔══██╗██╔════╝████╗ ████║██╔════╝████╗ ████║██╔═══██╗
█████╔╝ ██║   ██║██████╔╝█████╗  ██╔████╔██║█████╗  ██╔████╔██║██║   ██║
██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║██║   ██║
██║  ██╗╚██████╔╝██████╔╝███████╗██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║╚██████╔╝
╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝
`

const compactBanner = `
██╗  ██╗██╗   ██╗██████╗ ███████╗███╗   ███╗███████╗███╗   ███╗
██║ ██╔╝██║   ██║██╔══██╗██╔════╝████╗ ████║██╔════╝████╗ ████║
█████╔╝ ██║   ██║██████╔╝█████╗  ██╔████╔██║█████╗  ██╔████╔██║
██╔═██╗ ██║   ██║██╔══██╗██╔══╝  ██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║
██║  ██╗╚██████╔╝██████╔╝███████╗██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║
╚═╝  ╚═╝ ╚═════╝ ╚═════╝ ╚══════╝╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝
`

func Run(ctx context.Context, svc *service.Service, opts Options) error {
	m, err := newModel(ctx, svc, opts)
	if err != nil {
		return err
	}
	_, err = tea.NewProgram(m, tea.WithAltScreen()).Run()
	return err
}

func newModel(ctx context.Context, svc *service.Service, opts Options) (modelState, error) {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Placeholder = "filter"
	ti.Blur()
	vp := viewport.New(40, 20)
	rows, notes, err := buildRows(ctx, svc, opts, "", "", "")
	if err != nil {
		return modelState{}, err
	}
	tbl := table.New(
		table.WithColumns([]table.Column{
			{Title: "SRC", Width: 5},
			{Title: "TYPE", Width: 12},
			{Title: "NS", Width: 12},
			{Title: "TITLE", Width: 40},
		}),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	tbl.SetHeight(12)
	tbl.SetStyles(defaultTableStyles())
	m := modelState{
		service:     svc,
		opts:        opts,
		notes:       notes,
		table:       tbl,
		viewport:    vp,
		filterInput: ti,
	}
	m.updateViewport()
	return m, nil
}

func (m modelState) Init() tea.Cmd {
	return m.tickCmd()
}

func (m modelState) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := bannerHeight(m.width, m.height) + 3
		footerHeight := 3
		contentHeight := max(8, msg.Height-headerHeight-footerHeight)
		listWidth := max(48, msg.Width/2)
		m.table.SetWidth(listWidth)
		m.table.SetHeight(contentHeight)
		m.viewport.Width = max(32, msg.Width-listWidth-4)
		m.viewport.Height = contentHeight
		m.updateViewport()
		return m, nil
	case refreshMsg:
		rows, notes, err := buildRows(context.Background(), m.service, m.opts, m.textFilter, m.kindFilter, m.namespaceFilter)
		if err == nil {
			m.notes = notes
			m.table.SetRows(rows)
			if len(rows) == 0 {
				m.table.SetCursor(0)
			} else if m.table.Cursor() >= len(rows) {
				m.table.SetCursor(len(rows) - 1)
			}
			m.updateViewport()
		} else {
			m.err = err
		}
		return m, m.tickCmd()
	case tea.KeyMsg:
		if m.filterMode != "" {
			switch msg.String() {
			case "enter":
				value := strings.TrimSpace(m.filterInput.Value())
				switch m.filterMode {
				case "text":
					m.textFilter = value
				case "kind":
					m.kindFilter = value
				case "namespace":
					m.namespaceFilter = value
				}
				m.filterMode = ""
				m.filterInput.Blur()
				return m, func() tea.Msg { return refreshMsg{} }
			case "esc":
				m.filterMode = ""
				m.filterInput.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			m.updateViewport()
			return m, cmd
		case "k", "up":
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			m.updateViewport()
			return m, cmd
		case "pgdown", "d":
			m.viewport.LineDown(6)
			return m, nil
		case "pgup", "u":
			m.viewport.LineUp(6)
			return m, nil
		case "/":
			m.filterMode = "text"
			m.filterInput.SetValue(m.textFilter)
			m.filterInput.Focus()
			return m, textinput.Blink
		case "c":
			m.filterMode = "kind"
			m.filterInput.SetValue(m.kindFilter)
			m.filterInput.Focus()
			return m, textinput.Blink
		case "f":
			m.filterMode = "namespace"
			m.filterInput.SetValue(m.namespaceFilter)
			m.filterInput.Focus()
			return m, textinput.Blink
		case "r":
			return m, func() tea.Msg { return refreshMsg{} }
		}
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		m.updateViewport()
		return m, cmd
	}
	return m, nil
}

func (m modelState) View() string {
	banner := renderBanner(m.width)
	scope := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("153")).Render(" SCOPE " + m.scopeText() + " ")
	filters := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("110")).Render(" FILTERS text=" + displayNone(m.textFilter) + " kind=" + displayNone(m.kindFilter) + " ns=" + displayNone(m.namespaceFilter) + " ")
	header := strings.Join([]string{banner, scope, filters}, "\n")
	listTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(fmt.Sprintf(" LIST  %d memo(s) ", len(m.notes)))
	detailTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(" DETAIL ")
	if note := m.selectedNote(); note != nil {
		detailTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(" DETAIL  " + render.TargetLabel(*note) + " ")
	}
	left := listTitle + "\n" + m.table.View()
	right := detailTitle + "\n" + m.viewport.View()
	content := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(max(48, m.width/2)).Render(left), "  ", lipgloss.NewStyle().Width(max(32, m.width-max(48, m.width/2)-4)).Render(right))
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("109")).Render(" STATUS Ready ")
	keys := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Render(" [Arrows]/[j][k] move  [PgUp]/[PgDn] or [u][d] scroll  [/] text  [f] ns  [c] kind  [r] refresh  [q] quit ")
	filterPrompt := ""
	if m.filterMode != "" {
		filterPrompt = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Filter "+m.filterMode+": ") + m.filterInput.View()
	}
	errText := ""
	if m.err != nil {
		errText = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err.Error())
	}
	return strings.Join([]string{header, "", content, "", status, keys + filterPrompt + errText}, "\n")
}

func (m *modelState) updateViewport() {
	if m.viewport.Width == 0 {
		m.viewport.Width = 40
	}
	if m.viewport.Height == 0 {
		m.viewport.Height = 16
	}
	if note := m.selectedNote(); note != nil {
		m.viewport.SetContent(render.RenderNotes([]model.Note{*note}, render.CardOptions{Header: false, Width: max(28, m.viewport.Width-4)}))
	} else {
		m.viewport.SetContent("No memo selected.")
	}
}

func (m modelState) selectedNote() *model.Note {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.notes) {
		return nil
	}
	return &m.notes[cursor]
}

func (m modelState) tickCmd() tea.Cmd {
	if m.opts.AutoRefreshSeconds <= 0 {
		return nil
	}
	return tea.Tick(time.Duration(m.opts.AutoRefreshSeconds)*time.Second, func(time.Time) tea.Msg {
		return refreshMsg{}
	})
}

func (m modelState) scopeText() string {
	if m.namespaceFilter != "" {
		return m.namespaceFilter
	}
	if len(m.opts.Namespaces) > 0 {
		return strings.Join(m.opts.Namespaces, ",")
	}
	return "all accessible namespaces"
}

func buildRows(ctx context.Context, svc *service.Service, opts Options, textFilter, kindFilter, namespaceFilter string) ([]table.Row, []model.Note, error) {
	namespace := strings.TrimSpace(namespaceFilter)
	notes, err := svc.FindNotes(ctx, textFilter, "", kindFilter, namespace, "", opts.IncludeRuntime, true, opts.RuntimeNamespace)
	if err != nil {
		return nil, nil, err
	}
	rows := make([]table.Row, 0, len(notes))
	for _, note := range notes {
		src := "MEM"
		if note.StoreType == "Runtime" {
			src = "RUN"
		}
		rows = append(rows, table.Row{src, strings.ToUpper(note.NoteType), note.Namespace, note.Title})
	}
	return rows, notes, nil
}

func defaultTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.Foreground(lipgloss.Color("244")).BorderStyle(lipgloss.NormalBorder()).BorderBottom(true)
	s.Selected = s.Selected.Foreground(lipgloss.Color("16")).Background(lipgloss.Color("153")).Bold(true)
	return s
}

func displayNone(text string) string {
	if strings.TrimSpace(text) == "" {
		return "none"
	}
	return text
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderBanner(width int) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	switch bannerVariant(width, 0) {
	case "text":
		return style.Render("KubeMemo")
	case "compact":
		return style.Render(strings.TrimSpace(compactBanner))
	default:
		return style.Render(strings.TrimSpace(wideBanner))
	}
}

func bannerHeight(width, height int) int {
	switch bannerVariant(width, height) {
	case "text":
		return 1
	case "compact":
		return 6
	default:
		return 6
	}
}

func bannerVariant(width, height int) string {
	if width > 0 && width < 74 {
		return "text"
	}
	if height > 0 && height < 24 {
		return "text"
	}
	if width > 0 && width < 96 {
		return "compact"
	}
	if height > 0 && height < 30 {
		return "compact"
	}
	return "wide"
}
