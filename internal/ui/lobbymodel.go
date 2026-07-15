package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/torpido/internal/game"
	"github.com/ensardev/torpido/internal/lobby"
)

// lobbyRefresh is how often the room list is refreshed so new rooms appear.
const lobbyRefresh = 2 * time.Second

type lobbyTickMsg time.Time

// enterBotGameMsg tells the root model to start a game in a bot room the player
// just joined.
type enterBotGameMsg struct {
	difficulty game.Difficulty
	room       *lobby.Room
	seat       *lobby.Seat
}

// lobbyNoticeMsg carries a transient message to show in the lobby.
type lobbyNoticeMsg string

// lobbyModel is the screen a player sees after connecting: the list of joinable
// rooms and how to enter one.
type lobbyModel struct {
	lobby    *lobby.Lobby
	name     string
	renderer *lipgloss.Renderer
	styles   styles

	rooms  []lobby.RoomInfo
	cursor int
	notice string
}

func newLobbyModel(l *lobby.Lobby, name string, r *lipgloss.Renderer) lobbyModel {
	m := lobbyModel{
		lobby:    l,
		name:     name,
		renderer: r,
		styles:   newStyles(r),
	}
	m.refresh()
	return m
}

func (m *lobbyModel) refresh() {
	m.rooms = m.lobby.PublicRooms()
	// Stable order: bot rooms by tier first, then human rooms by code.
	sort.SliceStable(m.rooms, func(i, j int) bool {
		a, b := m.rooms[i], m.rooms[j]
		if a.Kind != b.Kind {
			return a.Kind == lobby.BotRoom
		}
		if a.Kind == lobby.BotRoom {
			return a.Tier < b.Tier
		}
		return a.Code < b.Code
	})
	if m.cursor >= len(m.rooms) {
		m.cursor = len(m.rooms) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func lobbyTick() tea.Cmd {
	return tea.Tick(lobbyRefresh, func(t time.Time) tea.Msg { return lobbyTickMsg(t) })
}

func (m lobbyModel) Init() tea.Cmd { return lobbyTick() }

func (m lobbyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case lobbyTickMsg:
		m.refresh()
		return m, lobbyTick()
	case lobbyNoticeMsg:
		m.notice = string(msg)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			m.notice = ""
		case "down", "j":
			if m.cursor < len(m.rooms)-1 {
				m.cursor++
			}
			m.notice = ""
		case "enter", " ":
			return m.selectRoom()
		case "c", "h", "g":
			// Create / quick-match / join-by-code arrive in the next step.
			m.notice = "İnsan-vs-insan sıradaki adımda geliyor 🔜"
		}
	}
	return m, nil
}

func (m lobbyModel) selectRoom() (tea.Model, tea.Cmd) {
	if len(m.rooms) == 0 {
		return m, nil
	}
	info := m.rooms[m.cursor]
	if info.Kind != lobby.BotRoom {
		m.notice = "İnsan odaları sıradaki adımda oynanabilir olacak 🔜"
		return m, nil
	}

	l, name, tier, code := m.lobby, m.name, info.Tier, info.Code
	return m, func() tea.Msg {
		seat := lobby.NewHumanSeat(name)
		room, err := l.JoinByCode(code, seat, "")
		if err != nil {
			return lobbyNoticeMsg(err.Error())
		}
		return enterBotGameMsg{difficulty: tier, room: room, seat: seat}
	}
}

// tierStyle picks the accent color for a bot difficulty.
func (s styles) tierStyle(d game.Difficulty) lipgloss.Style {
	switch d {
	case game.Rookie:
		return s.tierRookie
	case game.Admiral:
		return s.tierAdmiral
	case game.SeaWolf:
		return s.tierWolf
	default:
		return s.dim
	}
}

func (m lobbyModel) View() string {
	s := m.styles
	var rows []string
	for i, info := range m.rooms {
		var line string
		if info.Kind == lobby.BotRoom {
			line = fmt.Sprintf("%s %-12s %s",
				s.tierStyle(info.Tier).Render("●"),
				info.HostName,
				s.dim.Render("bot · 1/2 bekliyor"))
		} else {
			lock := ""
			if info.HasPassword {
				lock = "🔒 "
			}
			host := info.HostName
			if host == "" {
				host = "oyuncu"
			}
			line = fmt.Sprintf("%s⚔ %s %s",
				lock, s.logo.Render(info.Code),
				s.dim.Render(fmt.Sprintf("%s · %d/2 bekliyor", host, info.Players)))
		}
		if i == m.cursor {
			line = s.rosterNow.Render("▸ " + stripLeadingSpace(line))
		} else {
			line = "  " + line
		}
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = append(rows, s.dim.Render("  (oda yok)"))
	}

	list := s.box.Render(strings.Join(rows, "\n"))

	notice := ""
	if m.notice != "" {
		notice = s.tag.Render(m.notice) + "\n"
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(),
		"",
		s.dim.Render("AÇIK ODALAR — bir bot seç ve oyna:"),
		list,
		"",
		notice+s.help.Render("↑↓ seç · enter gir · c oda kur · h hızlı eşleş · q çık"),
	)
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}

// stripLeadingSpace trims one leading space so the cursor arrow lines up.
func stripLeadingSpace(sline string) string {
	return strings.TrimPrefix(sline, " ")
}
