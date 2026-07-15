package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/game"
)

func (m gameModel) View() string {
	switch m.phase {
	case gameWaiting:
		return m.viewWaiting()
	case gamePlacing:
		return m.viewPlacing()
	case gamePlaceWait:
		return m.viewPlaceWait()
	case gameBattle:
		return m.viewBattle()
	case gameOver:
		return m.viewOver()
	}
	return ""
}

// opponentName is who you're up against, for headers.
func (m gameModel) opponentName() string {
	if m.snap.OppName != "" {
		return m.snap.OppName
	}
	return m.t.Opponent
}

func (m gameModel) vsLine() string {
	return m.styles.header(m.t.Tagline) + "  " +
		m.styles.tag.Render("· "+fmt.Sprintf(m.t.VsFmt, m.opponentName()))
}

func (m gameModel) legend() string {
	return m.styles.legend(m.t.LegendShip, m.t.LegendHit, m.t.LegendMiss)
}

func (m gameModel) viewWaiting() string {
	s := m.styles
	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(m.t.Tagline),
		"",
		s.badgeYou.Render(m.t.Room+m.room.Code),
		"",
		s.tag.Render(m.t.WaitingOpp),
		s.dim.Render(m.t.ShareCode),
		"",
		s.help.Render(m.t.BackHelp),
	)
	return screen(body)
}

func (m gameModel) viewPlacing() string {
	s := m.styles
	st := m.fleet[m.placeIndex]
	coords := game.ShipCoords(m.cursor, st.Size, m.orientation)
	valid := m.previewValid(coords)
	preview := make(map[game.Coord]bool, len(coords))
	for _, c := range coords {
		preview[c] = true
	}

	orient := m.t.OrientH
	if m.orientation == game.Vertical {
		orient = m.t.OrientV
	}

	var roster []string
	for i, sh := range m.fleet {
		label := fmt.Sprintf("%s (%d)", sh.Name, sh.Size)
		switch {
		case i < m.placeIndex:
			roster = append(roster, s.rosterDone.Render("✔ "+label))
		case i == m.placeIndex:
			roster = append(roster, s.rosterNow.Render(label))
		default:
			roster = append(roster, s.rosterTodo.Render(label))
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		m.vsLine(),
		"",
		s.dim.Render(m.t.PlaceFleet),
		strings.Join(roster, "   "),
		"",
		s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, preview, valid)),
		"",
		m.legend(),
		"",
		s.help.Render(fmt.Sprintf(m.t.PlaceHelpFmt, orient)),
	)
	return screen(body)
}

func (m gameModel) viewPlaceWait() string {
	s := m.styles
	body := lipgloss.JoinVertical(lipgloss.Left,
		s.header(m.t.Tagline),
		"",
		s.badgeYou.Render(m.t.Ready),
		"",
		s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, nil, false)),
		"",
		s.tag.Render(fmt.Sprintf(m.t.OppPlacingFmt, m.opponentName())),
		"",
		s.help.Render(m.t.BackHelp),
	)
	return screen(body)
}

func (m gameModel) viewBattle() string {
	s := m.styles
	var aim *game.Coord
	if m.snap.YourTurn {
		aim = &m.aim
	}
	own := s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, nil, false))
	enemy := s.boardPanel(m.t.EnemyWaters, s.renderBoard(m.snap.Enemy, aim, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	var turn string
	if m.snap.YourTurn {
		turn = s.badgeYou.Render(m.t.YourTurn)
	} else {
		turn = s.badgeFoe.Render(fmt.Sprintf(m.t.OppAimingFmt, strings.ToUpper(m.opponentName())))
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		m.vsLine(),
		"",
		turn+"   "+m.message,
		"",
		boards,
		"",
		m.legend(),
		"",
		s.help.Render(m.t.BattleHelp),
	)
	return screen(body)
}

func (m gameModel) viewOver() string {
	s := m.styles
	banner := s.win.Render(m.t.Victory)
	msg := fmt.Sprintf(m.t.WinMsgFmt, m.opponentName())
	if !m.snap.YouWon {
		banner = s.lose.Render(m.t.Defeat)
		msg = fmt.Sprintf(m.t.LoseMsgFmt, m.opponentName())
	}
	own := s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, nil, false))
	enemy := s.boardPanel(m.t.EnemyWaters, s.renderBoard(m.snap.EnemyFull, nil, nil, false))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	body := lipgloss.JoinVertical(lipgloss.Left,
		banner+"   "+msg,
		"",
		boards,
		"",
		s.help.Render(m.t.OverHelp),
	)
	return screen(body)
}
