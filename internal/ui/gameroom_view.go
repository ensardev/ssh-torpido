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

// logLine formats one battle-log event from this player's point of view.
func (m gameModel) logLine(e game.Event) string {
	s := m.styles
	coord := coordName(e.Coord)
	opp := m.opponentName()
	byMe := e.Side == m.side
	switch {
	case byMe && e.Result == game.FireMiss:
		return s.dim.Render("• " + fmt.Sprintf(m.t.LogYouMissFmt, coord))
	case byMe && e.Result == game.FireHit:
		return s.logGood.Render("• " + fmt.Sprintf(m.t.LogYouHitFmt, coord))
	case byMe && e.Result == game.FireSunk:
		return s.logGood.Render("• " + fmt.Sprintf(m.t.LogYouSunkFmt, e.Ship))
	case !byMe && e.Result == game.FireMiss:
		return s.dim.Render("• " + fmt.Sprintf(m.t.LogOppMissFmt, opp, coord))
	case !byMe && e.Result == game.FireHit:
		return s.logHit.Render("• " + fmt.Sprintf(m.t.LogOppHitFmt, opp, coord))
	case !byMe && e.Result == game.FireSunk:
		return s.logHit.Render("• " + fmt.Sprintf(m.t.LogOppSunkFmt, opp, e.Ship))
	}
	return ""
}

// logPanel renders a fixed-height battle log so the layout never jumps.
func (m gameModel) logPanel(width int) string {
	s := m.styles
	const show = 6
	log := m.snap.Log
	if len(log) > show {
		log = log[len(log)-show:]
	}
	lines := make([]string, 0, show)
	for _, e := range log {
		lines = append(lines, m.logLine(e))
	}
	for len(lines) < show {
		lines = append([]string{" "}, lines...) // pad at the top, newest at bottom
	}
	inner := s.dim.Render(m.t.LogTitle) + "\n" + strings.Join(lines, "\n")
	box := s.logBox
	if width > 2 {
		box = box.Width(width - 2)
	}
	return box.Render(inner)
}

// headerRow lays the "vs" line on the left and a badge on the right.
func (m gameModel) headerRow(width int, badge string) string {
	left := m.vsLine()
	gap := width - lipgloss.Width(left) - lipgloss.Width(badge)
	if gap < 2 {
		gap = 2
	}
	return left + strings.Repeat(" ", gap) + badge
}

func (m gameModel) viewWaiting() string {
	s := m.styles
	content := lipgloss.JoinVertical(lipgloss.Center,
		s.header(m.t.Tagline),
		"",
		s.badgeYou.Render(m.t.Room+m.room.Code),
		"",
		s.tag.Render(m.t.WaitingOpp),
		s.dim.Render(m.t.ShareCode),
		"",
		s.help.Render(m.t.BackHelp),
	)
	return s.framed(m.width, content)
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

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.vsLine(),
		"",
		s.dim.Render(m.t.PlaceFleet),
		strings.Join(roster, "   "),
		"",
		s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, preview, valid, nil)),
		"",
		m.legend(),
		"",
		s.help.Render(fmt.Sprintf(m.t.PlaceHelpFmt, orient)),
	)
	return s.framed(m.width, content)
}

func (m gameModel) viewPlaceWait() string {
	s := m.styles
	content := lipgloss.JoinVertical(lipgloss.Center,
		s.header(m.t.Tagline),
		"",
		s.badgeYou.Render(m.t.Ready),
		"",
		s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, nil, false, nil)),
		"",
		s.tag.Render(fmt.Sprintf(m.t.OppPlacingFmt, m.opponentName())),
		"",
		s.help.Render(m.t.BackHelp),
	)
	return s.framed(m.width, content)
}

func (m gameModel) viewBattle() string {
	s := m.styles

	var aim *game.Coord
	if m.snap.YourTurn {
		aim = &m.aim
	}
	var ownBoom, enemyBoom *boomOverlay
	if m.boom != nil {
		if m.boom.onEnemy {
			enemyBoom = m.boom
		} else {
			ownBoom = m.boom
		}
	}

	own := s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, nil, false, ownBoom))
	enemy := s.boardPanel(m.t.EnemyWaters, s.renderBoard(m.snap.Enemy, aim, nil, false, enemyBoom))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)
	bw := lipgloss.Width(boards)

	var badge string
	if m.snap.YourTurn {
		badge = s.badgeYou.Render(m.t.YourTurn)
	} else {
		badge = s.badgeFoe.Render(fmt.Sprintf(m.t.OppAimingFmt, strings.ToUpper(m.opponentName())))
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		m.headerRow(bw, badge),
		"",
		boards,
		s.dim.Render(m.message),
		"",
		m.logPanel(bw),
		"",
		s.help.Render(m.t.BattleHelp),
	)
	return s.framed(m.width, content)
}

func (m gameModel) viewOver() string {
	s := m.styles
	banner := s.win.Render(m.t.Victory)
	msg := fmt.Sprintf(m.t.WinMsgFmt, m.opponentName())
	if !m.snap.YouWon {
		banner = s.lose.Render(m.t.Defeat)
		msg = fmt.Sprintf(m.t.LoseMsgFmt, m.opponentName())
	}
	own := s.boardPanel(m.t.YourWaters, s.renderBoard(m.snap.You, nil, nil, false, nil))
	enemy := s.boardPanel(m.t.EnemyWaters, s.renderBoard(m.snap.EnemyFull, nil, nil, false, nil))
	boards := lipgloss.JoinHorizontal(lipgloss.Top, own, "    ", enemy)

	content := lipgloss.JoinVertical(lipgloss.Center,
		banner,
		s.dim.Render(msg),
		"",
		boards,
		"",
		s.help.Render(m.t.OverHelp),
	)
	return s.framed(m.width, content)
}
