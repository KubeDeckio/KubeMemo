package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/KubeDeckio/KubeMemo/internal/model"
	"github.com/charmbracelet/lipgloss"
)

type CardOptions struct {
	Width   int
	NoColor bool
	Header  bool
}

func RenderNotes(notes []model.Note, opts CardOptions) string {
	if opts.Width <= 0 {
		opts.Width = 78
	}
	parts := []string{}
	if opts.Header && len(notes) > 0 {
		title := "KubeMemo  " + TargetLabel(notes[0])
		parts = append(parts, styleFor("header", opts.NoColor).Render(title))
		parts = append(parts, styleFor("rule", opts.NoColor).Render(strings.Repeat("═", maxInt(len(title), 72))))
	}

	groups := map[string][]model.Note{}
	for _, note := range notes {
		groups[note.StoreType] = append(groups[note.StoreType], note)
	}

	for _, key := range []string{"Durable", "Runtime"} {
		group := groups[key]
		if len(group) == 0 {
			continue
		}
		sectionStyle := "durableSection"
		if key == "Runtime" {
			sectionStyle = "runtimeSection"
		}
		parts = append(parts, styleFor(sectionStyle, opts.NoColor).Render(" "+strings.ToUpper(key)+" MEMOS "))
		for _, note := range group {
			parts = append(parts, RenderNoteCard(note, opts.Width, !opts.NoColor))
		}
	}

	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func RenderNoteCard(note model.Note, width int, color bool) string {
	if width < 24 {
		width = 24
	}

	style := noteStyle(note, color)
	bodyWidth := width
	lines := []string{}

	pin := pinIcon(note)
	titleBadge := style.titleBadge.Render(" " + strings.ToUpper(note.NoteType) + " ")
	titleText := fitText(note.Title, maxInt(12, bodyWidth-16))
	lines = append(lines, style.header.Width(bodyWidth).Render(titleBadge+"  "+titleText))

	subtitleParts := []string{TargetLabel(note)}
	if note.StoreType == "Runtime" && note.ExpiresAt != nil {
		subtitleParts = append(subtitleParts, "expires "+relativeTime(*note.ExpiresAt))
	}
	lines = append(lines, mergeStyle(style.meta, style.paper).Width(bodyWidth).Render(strings.Join(subtitleParts, " • ")))

	context := ""
	if note.StoreType == "Runtime" {
		context = "Source: " + firstNonEmpty(note.SourceGenerator, note.CreatedBy, note.SourceType, "runtime memo")
	} else {
		context = "Owner: " + firstNonEmpty(strings.TrimSpace(strings.TrimSpace(note.OwnerTeam+" "+note.OwnerContact)), note.CreatedBy, "curated memo")
	}
	lines = append(lines, mergeStyle(style.subtle, style.paper).Width(bodyWidth).Render(context))
	if authorship := authorshipLabel(note); authorship != "" {
		lines = append(lines, mergeStyle(style.subtle, style.paper).Width(bodyWidth).Render(authorship))
	}
	lines = append(lines, style.paper.Width(bodyWidth).Render(""))

	if note.StoreType == "Runtime" {
		lines = append(lines, mergeStyle(style.runtimeNote, style.paper).Width(bodyWidth).Render("Temporary operational context for the current cluster state."))
		lines = append(lines, style.paper.Width(bodyWidth).Render(""))
	}

	lines = append(lines, wrapSection("Summary", note.Summary, bodyWidth, style)...)

	detailLabel := "Details"
	switch strings.ToLower(note.NoteType) {
	case "warning":
		detailLabel = "Guidance"
	case "incident", "activity":
		detailLabel = "Notes"
	case "runbook":
		detailLabel = "Runbook"
	}
	lines = append(lines, wrapSection(detailLabel, note.Content, bodyWidth, style)...)

	cardLines := []string{renderTopBorder(bodyWidth, pin, style.border)}
	for _, line := range lines {
		cardLines = append(cardLines, wrapCardLine(line, bodyWidth, style.border))
	}
	cardLines = append(cardLines, renderBottomBorder(bodyWidth, style.border))
	return strings.Join(cardLines, "\n")
}

func TargetLabel(note model.Note) string {
	switch note.TargetMode {
	case "namespace":
		return "namespace/" + note.Namespace
	case "app":
		if note.AppInstance != "" {
			return note.AppName + "/" + note.AppInstance
		}
		return note.AppName
	default:
		label := note.Kind
		if note.Namespace != "" {
			label += "/" + note.Namespace
		}
		if note.Name != "" {
			label += "/" + note.Name
		}
		return label
	}
}

type renderStyle struct {
	border      lipgloss.Style
	header      lipgloss.Style
	paper       lipgloss.Style
	meta        lipgloss.Style
	subtle      lipgloss.Style
	label       lipgloss.Style
	body        lipgloss.Style
	titleBadge  lipgloss.Style
	statusBadge lipgloss.Style
	runtimeNote lipgloss.Style
}

func noteStyle(note model.Note, color bool) renderStyle {
	if !color {
		base := lipgloss.NewStyle()
		return renderStyle{
			border:      base,
			header:      base.Bold(true),
			paper:       base,
			meta:        base,
			subtle:      base,
			label:       base,
			body:        base,
			titleBadge:  base.Bold(true),
			statusBadge: base.Bold(true),
			runtimeNote: base,
		}
	}

	style := renderStyle{
		statusBadge: lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("230")).Bold(true),
	}

	if note.StoreType == "Runtime" {
		style.border = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
		style.header = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
		style.paper = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		style.meta = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
		style.subtle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
		style.label = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
		style.body = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
		style.titleBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("214")).Bold(true)
		style.statusBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("229")).Bold(true)
		style.runtimeNote = lipgloss.NewStyle().Foreground(lipgloss.Color("223")).Italic(true)
		return style
	}

	style.border = lipgloss.NewStyle().Foreground(lipgloss.Color("110"))
	style.header = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	style.paper = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	style.meta = lipgloss.NewStyle().Foreground(lipgloss.Color("247"))
	style.subtle = lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	style.label = lipgloss.NewStyle().Foreground(lipgloss.Color("186"))
	style.body = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	style.runtimeNote = style.body
	style.titleBadge = lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(noteTypeColor(note.NoteType)).Bold(true)
	return style
}

func styleFor(name string, noColor bool) lipgloss.Style {
	if noColor {
		return lipgloss.NewStyle()
	}
	switch name {
	case "header":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("45")).Bold(true)
	case "rule":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("60"))
	case "durableSection":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("110")).Bold(true)
	case "runtimeSection":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("16")).Background(lipgloss.Color("214")).Bold(true)
	default:
		return lipgloss.NewStyle()
	}
}

func wrapSection(label, text string, width int, style renderStyle) []string {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	lines := []string{mergeStyle(style.label, style.paper).Width(width).Render(strings.ToLower(label))}
	for _, line := range wrapText(text, maxInt(10, width-2)) {
		lines = append(lines, mergeStyle(style.body, style.paper).Width(width).Render("  "+line))
	}
	lines = append(lines, style.paper.Width(width).Render(""))
	return lines
}

func wrapText(text string, width int) []string {
	rawLines := strings.Split(strings.TrimSpace(text), "\n")
	out := []string{}
	for _, raw := range rawLines {
		line := strings.TrimRight(raw, " ")
		if line == "" {
			out = append(out, "")
			continue
		}
		for len(line) > width {
			cut := width
			if idx := strings.LastIndex(line[:width], " "); idx >= width/3 {
				cut = idx
			}
			out = append(out, strings.TrimSpace(line[:cut]))
			line = strings.TrimSpace(line[cut:])
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func fitText(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	}
	if width <= 1 {
		return "…"
	}
	runes := []rune(text)
	if len(runes) >= width {
		return string(runes[:width-1]) + "…"
	}
	return text
}

func padVisible(text string, width int) string {
	padding := width - lipgloss.Width(text)
	if padding <= 0 {
		return text
	}
	return text + strings.Repeat(" ", padding)
}

func mergeStyle(primary, fallback lipgloss.Style) lipgloss.Style {
	s := fallback
	if primary.GetForeground() != nil {
		s = s.Foreground(primary.GetForeground())
	}
	if primary.GetBackground() != nil {
		s = s.Background(primary.GetBackground())
	}
	if primary.GetBold() {
		s = s.Bold(true)
	}
	if primary.GetItalic() {
		s = s.Italic(true)
	}
	return s
}

func pinIcon(note model.Note) string {
	if note.StoreType == "Runtime" {
		return "📍"
	}
	return "📌"
}

func renderTopBorder(width int, pin string, borderStyle lipgloss.Style) string {
	inner := width + 2
	pinWidth := lipgloss.Width(pin)
	leftPad := 1
	rightPad := 1
	dashes := inner - leftPad - rightPad - (pinWidth * 2)
	if dashes < 2 {
		dashes = 2
	}
	leftDashes := dashes / 2
	rightDashes := dashes - leftDashes
	return borderStyle.Render("╭" + strings.Repeat("─", leftPad) + pin + strings.Repeat("─", leftDashes) + strings.Repeat("─", rightDashes) + pin + strings.Repeat("─", rightPad) + "╮")
}

func renderBottomBorder(width int, borderStyle lipgloss.Style) string {
	return borderStyle.Render("╰" + strings.Repeat("─", width+2) + "╯")
}

func wrapCardLine(text string, width int, borderStyle lipgloss.Style) string {
	return borderStyle.Render("│ ") + padVisible(text, width) + borderStyle.Render(" │")
}

func relativeTime(ts time.Time) string {
	now := time.Now().UTC()
	diff := ts.Sub(now)
	prefix := "in "
	if diff < 0 {
		diff = -diff
		prefix = ""
	}
	switch {
	case diff < time.Minute:
		return prefix + fmt.Sprintf("%ds", int(diff.Seconds()))
	case diff < time.Hour:
		return prefix + fmt.Sprintf("%dm", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return prefix + fmt.Sprintf("%dh", int(diff.Hours()))
	default:
		return prefix + fmt.Sprintf("%dd", int(diff.Hours()/24))
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func authorshipLabel(note model.Note) string {
	updatedBy := strings.TrimSpace(note.UpdatedBy)
	createdBy := strings.TrimSpace(note.CreatedBy)
	switch {
	case updatedBy != "" && updatedBy != createdBy:
		return "Updated by: " + updatedBy
	case createdBy != "":
		return "Created by: " + createdBy
	default:
		return ""
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func noteTypeColor(noteType string) lipgloss.Color {
	switch strings.ToLower(noteType) {
	case "warning", "temporary-warning":
		return lipgloss.Color("214")
	case "ownership":
		return lipgloss.Color("120")
	case "runbook":
		return lipgloss.Color("81")
	case "maintenance":
		return lipgloss.Color("179")
	case "handover":
		return lipgloss.Color("176")
	case "suppression", "temporary-suppression":
		return lipgloss.Color("245")
	case "incident":
		return lipgloss.Color("203")
	case "activity":
		return lipgloss.Color("81")
	default:
		return lipgloss.Color("153")
	}
}
