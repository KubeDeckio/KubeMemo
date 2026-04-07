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
	allNotes        []model.Note
	opts            Options
	notes           []model.Note
	table           table.Model
	viewport        viewport.Model
	filterInput     textinput.Model
	filterMode      string
	kindFilter      string
	namespaceFilter string
	storeFilter     string
	textFilter      string
	page            int
	expandedDetail  bool
	width           int
	height          int
	showHelp        bool
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
███╗   ███╗███████╗███╗   ███╗ ██████╗
████╗ ████║██╔════╝████╗ ████║██╔═══██╗
██╔████╔██║█████╗  ██╔████╔██║██║   ██║
██║╚██╔╝██║██╔══╝  ██║╚██╔╝██║██║   ██║
██║ ╚═╝ ██║███████╗██║ ╚═╝ ██║╚██████╔╝
╚═╝     ╚═╝╚══════╝╚═╝     ╚═╝ ╚═════╝
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
	rows, notes, err := buildRows(ctx, svc, opts, "", "", "", "")
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
		allNotes:    notes,
		opts:        opts,
		notes:       notes,
		table:       tbl,
		viewport:    vp,
		filterInput: ti,
		storeFilter: "all",
	}
	m.applyPage()
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
		headerHeight := bannerHeight(m.width, m.height) + 4
		footerHeight := 3
		contentHeight := max(8, msg.Height-headerHeight-footerHeight)
		m.updateLayout(contentHeight)
		m.applyPage()
		m.updateViewport()
		return m, nil
	case refreshMsg:
		_, notes, err := buildRows(context.Background(), m.service, m.opts, m.textFilter, m.kindFilter, m.namespaceFilter, m.storeFilter)
		if err == nil {
			m.allNotes = notes
			m.applyPage()
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
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "enter":
			if m.selectedNote() != nil {
				m.expandedDetail = !m.expandedDetail
				m.updateLayout(max(8, m.height-bannerHeight(m.width, m.height)-4-3))
				m.updateViewport()
			}
			return m, nil
		case "j", "down":
			if len(m.notes) > 0 && m.table.Cursor() >= len(m.notes)-1 && m.page < m.pageCount()-1 {
				m.page++
				m.applyPage()
				m.table.SetCursor(0)
				m.updateViewport()
				return m, nil
			}
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			m.updateViewport()
			return m, cmd
		case "k", "up":
			if m.table.Cursor() == 0 && m.page > 0 {
				m.page--
				m.applyPage()
				if len(m.notes) > 0 {
					m.table.SetCursor(len(m.notes) - 1)
				}
				m.updateViewport()
				return m, nil
			}
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
		case "n":
			m.filterMode = "namespace"
			m.filterInput.SetValue(m.namespaceFilter)
			m.filterInput.Focus()
			return m, textinput.Blink
		case "a":
			m.storeFilter = "all"
			return m, func() tea.Msg { return refreshMsg{} }
		case "m":
			m.storeFilter = "durable"
			return m, func() tea.Msg { return refreshMsg{} }
		case "t":
			m.storeFilter = "runtime"
			return m, func() tea.Msg { return refreshMsg{} }
		case "x":
			m.textFilter = ""
			m.kindFilter = ""
			m.namespaceFilter = ""
			m.storeFilter = "all"
			return m, func() tea.Msg { return refreshMsg{} }
		case "g", "home":
			m.page = 0
			m.applyPage()
			m.table.SetCursor(0)
			m.updateViewport()
			return m, nil
		case "G", "end":
			if len(m.allNotes) > 0 {
				m.page = max(0, m.pageCount()-1)
				m.applyPage()
				m.table.SetCursor(max(0, len(m.notes)-1))
				m.updateViewport()
			}
			return m, nil
		case "]":
			if m.page < m.pageCount()-1 {
				m.page++
				m.applyPage()
				m.updateViewport()
			}
			return m, nil
		case "[":
			if m.page > 0 {
				m.page--
				m.applyPage()
				m.updateViewport()
			}
			return m, nil
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
	header := m.renderHeader()
	listTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(" MEMOS ")
	detailTitle := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(" DETAIL ")
	if note := m.selectedNote(); note != nil {
		detailTitle = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(" DETAIL  " + render.TargetLabel(*note) + " ")
	}
	left := listTitle + m.renderListMeta() + "\n" + m.table.View()
	right := detailTitle + "\n" + m.viewport.View() + "\n" + m.renderDetailMeta()
	var content string
	if m.expandedDetail {
		content = right
	} else if m.isStackedLayout() {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Width(max(40, m.width)).Render(left),
			"",
			lipgloss.NewStyle().Width(max(40, m.width)).Render(right),
		)
	} else {
		content = lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(max(48, m.width/2)).Render(left),
			"  ",
			lipgloss.NewStyle().Width(max(32, m.width-max(48, m.width/2)-4)).Render(right),
		)
	}
	status := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("109")).Render(" STATUS " + m.statusText() + " ")
	keys := m.renderFooterKeys()
	filterPrompt := ""
	if m.filterMode != "" {
		filterPrompt = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("Filter "+m.filterMode+": ") + m.filterInput.View()
	}
	errText := ""
	if m.err != nil {
		errText = "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Render(m.err.Error())
	}
	helpBlock := ""
	if m.showHelp {
		helpBlock = "\n\n" + m.renderHelp()
	}
	header = lipgloss.NewStyle().PaddingTop(1).Render(header)
	return strings.Join([]string{header, "", content, "", status, keys + filterPrompt + errText + helpBlock}, "\n")
}

func (m modelState) renderHeader() string {
	banner := renderBanner(m.width)
	scope := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("153")).Render(" SCOPE " + m.scopeText() + " ")
	filters := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("110")).Render(" FILTER " + displayNone(m.textFilter) + " ")
	kind := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("146")).Render(" KIND " + displayNone(m.kindFilter) + " ")
	store := "all"
	if m.storeFilter != "" {
		store = m.storeFilter
	} else if !m.opts.IncludeRuntime {
		store = "durable"
	}
	view := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("214")).Render(" VIEW " + store + " ")
	runtime := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Render(" RUNTIME " + displayNone(m.opts.RuntimeNamespace) + " ")
	counts := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("109")).Render(fmt.Sprintf(" COUNT %d ", len(m.notes)))
	metaBlock := strings.Join([]string{scope, filters, kind, view, runtime, counts}, "\n")

	if m.width >= 118 && bannerVariant(m.width, m.height) != "text" {
		leftWidth := max(50, m.width-42)
		rightWidth := max(34, m.width-leftWidth-2)
		return lipgloss.JoinHorizontal(
			lipgloss.Top,
			lipgloss.NewStyle().Width(leftWidth).Render(banner),
			"  ",
			lipgloss.NewStyle().Width(rightWidth).Align(lipgloss.Right).Render(metaBlock),
		)
	}

	return strings.Join([]string{banner, scope, lipgloss.JoinHorizontal(lipgloss.Top, filters, " ", kind, " ", view), runtime}, "\n")
}

func (m modelState) renderFooterKeys() string {
	key := lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("45")).Bold(true)
	desc := lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	enterLabel := "focus detail"
	if m.expandedDetail {
		enterLabel = "collapse detail"
	}
	items := []string{
		key.Render(" ↑ ↓ / j k ") + desc.Render(" move"),
		key.Render(" g / G ") + desc.Render(" top/end"),
		key.Render(" Enter ") + desc.Render(" " + enterLabel),
		key.Render(" PgUp PgDn / u d ") + desc.Render(" scroll"),
		key.Render(" / ") + desc.Render(" text"),
		key.Render(" n ") + desc.Render(" ns"),
		key.Render(" c ") + desc.Render(" kind"),
		key.Render(" a / m / t ") + desc.Render(" all/durable/runtime"),
		key.Render(" [ ") + desc.Render(" prev page"),
		key.Render(" ] ") + desc.Render(" next page"),
		key.Render(" x ") + desc.Render(" clear"),
		key.Render(" r ") + desc.Render(" refresh"),
		key.Render(" ? ") + desc.Render(" help"),
		key.Render(" q ") + desc.Render(" quit"),
	}
	return lipgloss.NewStyle().
		Width(max(24, m.width)).
		Render(strings.Join(items, "  "))
}

func (m modelState) renderDetailMeta() string {
	total := max(1, m.viewport.TotalLineCount())
	visible := m.viewport.VisibleLineCount()
	if visible <= 0 || total <= visible {
		return lipgloss.NewStyle().
			Width(max(20, m.viewport.Width)).
			Render("")
	}

	position := m.viewport.YOffset + 1
	percent := int(m.viewport.ScrollPercent() * 100)

	parts := []string{
		lipgloss.NewStyle().Foreground(lipgloss.Color("109")).Render(fmt.Sprintf("scroll %d/%d", position, total)),
	}
	if !m.viewport.AtTop() {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("↑ more"))
	}
	if !m.viewport.AtBottom() {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render("↓ more"))
	}
	parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render(fmt.Sprintf("%d%%", percent)))

	return lipgloss.NewStyle().
		Width(max(20, m.viewport.Width)).
		Align(lipgloss.Right).
		Render(strings.Join(parts, "  "))
}

func (m modelState) renderHelp() string {
	enterLine := "[Enter] toggle focused detail view"
	if m.expandedDetail {
		enterLine = "[Enter] collapse focused detail view"
	}
	lines := []string{
		"[/] text filter  [n] namespace filter  [c] kind filter",
		"[a] all memos  [m] durable only  [t] runtime only  [x] clear filters",
		enterLine,
		"[j][k] or arrows move  [g]/[G] jump top/end  [u][d] or [PgUp]/[PgDn] scroll detail",
		"[r] refresh  [q] quit",
	}
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("45")).
		Padding(0, 1).
		Foreground(lipgloss.Color("252")).
		Render(strings.Join(lines, "\n"))
}

func (m modelState) statusText() string {
	if m.err != nil {
		return "Error"
	}
	view := "all"
	if m.storeFilter != "" {
		view = m.storeFilter
	} else if !m.opts.IncludeRuntime {
		view = "durable"
	}
	if len(m.notes) == 0 {
		return fmt.Sprintf("0/%d | page %d/%d | view %s", len(m.allNotes), m.page+1, max(1, m.pageCount()), view)
	}
	return fmt.Sprintf("%d/%d | page %d/%d | view %s", (m.page*m.pageSize())+m.table.Cursor()+1, len(m.allNotes), m.page+1, max(1, m.pageCount()), view)
}

func (m modelState) renderListMeta() string {
	pageCount := m.pageCount()
	filterBits := []string{}
	if strings.TrimSpace(m.textFilter) != "" {
		filterBits = append(filterBits, "text:"+m.textFilter)
	}
	if strings.TrimSpace(m.namespaceFilter) != "" {
		filterBits = append(filterBits, "ns:"+m.namespaceFilter)
	}
	if strings.TrimSpace(m.kindFilter) != "" {
		filterBits = append(filterBits, "kind:"+m.kindFilter)
	}
	view := "all"
	if m.storeFilter != "" {
		view = m.storeFilter
	} else if !m.opts.IncludeRuntime {
		view = "durable"
	}
	if view != "all" {
		filterBits = append(filterBits, "view:"+view)
	}

	parts := []string{}
	if pageCount > 1 {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Render(fmt.Sprintf("  page %d/%d", m.page+1, pageCount)))
	}
	if len(filterBits) > 0 {
		parts = append(parts, lipgloss.NewStyle().Foreground(lipgloss.Color("110")).Render("  "+strings.Join(filterBits, "  ")))
	}
	return strings.Join(parts, "")
}

func (m *modelState) updateLayout(contentHeight int) {
	if m.expandedDetail {
		m.viewport.Width = max(40, m.width-2)
		m.viewport.Height = max(8, contentHeight)
		return
	}
	listWidth := max(48, m.width/2)
	viewportWidth := max(32, m.width-listWidth-4)
	if m.isStackedLayout() {
		listWidth = max(44, m.width)
		viewportWidth = max(36, m.width)
	}
	tableWidth := max(40, listWidth-2)
	m.table.SetWidth(tableWidth)
	tableHeight := contentHeight
	viewportHeight := max(6, contentHeight-1)
	if m.isStackedLayout() {
		tableHeight = max(8, contentHeight/2)
		viewportHeight = max(8, contentHeight-tableHeight-2)
	}
	m.table.SetHeight(tableHeight)
	titleWidth := max(16, tableWidth-36)
	m.table.SetColumns([]table.Column{
		{Title: "SRC", Width: 5},
		{Title: "TYPE", Width: 12},
		{Title: "NS", Width: 12},
		{Title: "TITLE", Width: titleWidth},
	})
	m.viewport.Width = viewportWidth
	m.viewport.Height = viewportHeight
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

func (m modelState) isStackedLayout() bool {
	return m.width > 0 && m.width < 120
}

func (m modelState) pageSize() int {
	if m.table.Height() > 0 {
		return m.table.Height()
	}
	return 12
}

func (m modelState) pageCount() int {
	size := m.pageSize()
	if size <= 0 || len(m.allNotes) == 0 {
		return 1
	}
	count := len(m.allNotes) / size
	if len(m.allNotes)%size != 0 {
		count++
	}
	if count < 1 {
		return 1
	}
	return count
}

func (m *modelState) applyPage() {
	pageCount := m.pageCount()
	if m.page >= pageCount {
		m.page = pageCount - 1
	}
	if m.page < 0 {
		m.page = 0
	}
	size := m.pageSize()
	start := m.page * size
	if start > len(m.allNotes) {
		start = len(m.allNotes)
	}
	end := start + size
	if end > len(m.allNotes) {
		end = len(m.allNotes)
	}
	m.notes = append([]model.Note(nil), m.allNotes[start:end]...)
	rows := make([]table.Row, 0, len(m.notes))
	for _, note := range m.notes {
		src := "MEM"
		if note.StoreType == "Runtime" {
			src = "RUN"
		}
		rows = append(rows, table.Row{src, strings.ToUpper(note.NoteType), note.Namespace, note.Title})
	}
	m.table.SetRows(rows)
	if len(rows) == 0 {
		m.table.SetCursor(0)
	} else if m.table.Cursor() >= len(rows) {
		m.table.SetCursor(len(rows) - 1)
	}
}

func buildRows(ctx context.Context, svc *service.Service, opts Options, textFilter, kindFilter, namespaceFilter, storeFilter string) ([]table.Row, []model.Note, error) {
	namespace := strings.TrimSpace(namespaceFilter)
	notes, err := svc.FindNotes(ctx, textFilter, "", kindFilter, namespace, "", opts.IncludeRuntime, false, opts.RuntimeNamespace)
	if err != nil {
		return nil, nil, err
	}
	rows := make([]table.Row, 0, len(notes))
	filtered := make([]model.Note, 0, len(notes))
	for _, note := range notes {
		if storeFilter == "durable" && note.StoreType != "Durable" {
			continue
		}
		if storeFilter == "runtime" && note.StoreType != "Runtime" {
			continue
		}
		src := "MEM"
		if note.StoreType == "Runtime" {
			src = "RUN"
		}
		rows = append(rows, table.Row{src, strings.ToUpper(note.NoteType), note.Namespace, note.Title})
		filtered = append(filtered, note)
	}
	return rows, filtered, nil
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
	if width > 0 && width < 120 {
		return style.Render("KubeMemo")
	}
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
