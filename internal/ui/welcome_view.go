package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m welcomeModel) View() string {
	switch m.page {
	case pageHowTo:
		return m.viewInfo(m.t.WHowToTitle, m.t.WHowToBody)
	case pageWhatSSH:
		return m.viewInfo(m.t.WWhatSSHTitle, m.t.WWhatSSHBody)
	case pageAbout:
		return m.viewInfo(m.t.WAboutTitle, m.t.WAboutBody)
	default:
		return m.viewMenu()
	}
}

// renderBanner colors the wordmark rows with a top-to-bottom gradient.
func (m welcomeModel) renderBanner() string {
	rows := banner("TORPIDO")
	out := make([]string, len(rows))
	for i, row := range rows {
		out[i] = m.renderer.NewStyle().Bold(true).Foreground(lipgloss.Color(bannerShades[i])).Render(row)
	}
	return strings.Join(out, "\n")
}

// renderTorpedo draws a torpedo sailing across a waterline, animated by frame.
func (m welcomeModel) renderTorpedo(width int) string {
	if width < 8 {
		width = 8
	}
	const sprite = "══►"
	pos := m.frame % (width + len(sprite))
	line := []rune(strings.Repeat("·", width))
	for i, ch := range sprite {
		p := pos - len(sprite) + 1 + i
		if p >= 0 && p < width {
			line[p] = ch
		}
	}
	s := string(line)
	// Color the water dim and the torpedo bright.
	torp := m.styles.tierAdmiral.Render(sprite)
	if idx := strings.Index(s, sprite); idx >= 0 {
		return m.styles.water.Render(s[:idx]) + torp + m.styles.water.Render(s[idx+len(sprite):])
	}
	return m.styles.water.Render(s)
}

func (m welcomeModel) viewMenu() string {
	s := m.styles
	logo := m.renderBanner()
	bannerW := lipgloss.Width(logo)

	tagline := s.tag.Render(m.t.Tagline)
	torpedo := m.renderTorpedo(bannerW)

	items := []string{m.t.WPlay, m.t.WHowTo, m.t.WWhatSSH, m.t.WAbout,
		fmt.Sprintf("%s: %s", m.t.WLanguage, m.lang.Label()), m.t.WQuit}
	var menu []string
	for i, it := range items {
		if i == m.cursor {
			menu = append(menu, s.logo.Render("▸ ")+s.rosterNow.Render(it))
		} else {
			menu = append(menu, "  "+s.dim.Render(it))
		}
	}

	block := lipgloss.JoinVertical(lipgloss.Center,
		logo,
		tagline,
		"",
		torpedo,
		"",
		"",
		lipgloss.JoinVertical(lipgloss.Left, menu...),
		"",
		s.help.Render(m.t.WNav),
		s.dim.Render("ssh torpido.dev"),
	)
	return lipgloss.NewStyle().Width(m.width).Align(lipgloss.Center).Padding(1, 0).Render(block)
}

func (m welcomeModel) viewInfo(title, body string) string {
	s := m.styles
	head := s.logo.Render(title)
	content := lipgloss.JoinVertical(lipgloss.Left,
		head,
		"",
		body,
		"",
		s.help.Render(m.t.WInfoBack),
	)
	return screen(content)
}
