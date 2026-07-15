package ui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/i18n"
)

// welcomeAnimTick is how fast the torpedo animation advances.
const welcomeAnimTick = 110 * time.Millisecond

// blockLetters is a small 5-row block font for the TORPIDO wordmark.
var blockLetters = map[rune][5]string{
	'T': {"█████", "  █  ", "  █  ", "  █  ", "  █  "},
	'O': {" ███ ", "█   █", "█   █", "█   █", " ███ "},
	'R': {"████ ", "█   █", "████ ", "█  █ ", "█   █"},
	'P': {"████ ", "█   █", "████ ", "█    ", "█    "},
	'I': {"█", "█", "█", "█", "█"},
	'D': {"████ ", "█   █", "█   █", "█   █", "████ "},
}

// banner builds the 5 rows of a word from the block font.
func banner(word string) []string {
	rows := make([]string, 5)
	for i := 0; i < 5; i++ {
		var parts []string
		for _, ch := range word {
			parts = append(parts, blockLetters[ch][i])
		}
		rows[i] = strings.Join(parts, " ")
	}
	return rows
}

// bannerShades colors the wordmark rows top-to-bottom for a bit of depth.
var bannerShades = []string{"51", "45", "39", "33", "27"}

type welcomePage int

const (
	pageMenu welcomePage = iota
	pageHowTo
	pageWhatSSH
	pageAbout
)

// welcome menu item indices.
const (
	miPlay = iota
	miHowTo
	miWhatSSH
	miAbout
	miLanguage
	miQuit
	miCount
)

type welcomeTickMsg struct{}

// startLobbyMsg tells the root to leave the welcome screen for the lobby, in the
// language the player picked.
type startLobbyMsg struct{ lang i18n.Lang }

type welcomeModel struct {
	lang     i18n.Lang
	t        i18n.Strings
	renderer *lipgloss.Renderer
	styles   styles

	page   welcomePage
	cursor int
	frame  int
	width  int
}

func newWelcomeModel(lang i18n.Lang, r *lipgloss.Renderer) welcomeModel {
	return welcomeModel{
		lang:     lang,
		t:        i18n.For(lang),
		renderer: r,
		styles:   newStyles(r),
		width:    80,
	}
}

func welcomeTick() tea.Cmd {
	return tea.Tick(welcomeAnimTick, func(time.Time) tea.Msg { return welcomeTickMsg{} })
}

func (m welcomeModel) Init() tea.Cmd { return welcomeTick() }

func (m welcomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		return m, nil
	case welcomeTickMsg:
		m.frame++
		return m, welcomeTick()
	case tea.KeyMsg:
		if m.page != pageMenu {
			m.page = pageMenu // any key returns from an info page
			return m, nil
		}
		return m.updateMenu(msg)
	}
	return m, nil
}

func (m welcomeModel) updateMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < miCount-1 {
			m.cursor++
		}
	case "left", "right", "h", "l":
		if m.cursor == miLanguage {
			m.setLang(m.lang.Next())
		}
	case "enter", " ":
		return m.selectItem()
	}
	return m, nil
}

func (m *welcomeModel) setLang(l i18n.Lang) {
	m.lang = l
	m.t = i18n.For(l)
}

func (m welcomeModel) selectItem() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case miPlay:
		return m, func() tea.Msg { return startLobbyMsg{lang: m.lang} }
	case miHowTo:
		m.page = pageHowTo
	case miWhatSSH:
		m.page = pageWhatSSH
	case miAbout:
		m.page = pageAbout
	case miLanguage:
		m.setLang(m.lang.Next())
	case miQuit:
		return m, tea.Quit
	}
	return m, nil
}
