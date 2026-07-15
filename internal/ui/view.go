package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ensardev/ssh-torpido/internal/game"
	"github.com/ensardev/ssh-torpido/internal/i18n"
)

// This file holds the shared drawing helpers used by every screen. They take
// value-copy grids (never a live board), so rendering is always race-free.

type grid = [game.BoardSize][game.BoardSize]game.Cell
type hullGrid = [game.BoardSize][game.BoardSize]game.Hull

// botName is the difficulty's display name in the player's language.
func botName(t i18n.Strings, d game.Difficulty) string {
	switch d {
	case game.Rookie:
		return t.BotRookie
	case game.Admiral:
		return t.BotAdmiral
	case game.SeaWolf:
		return t.BotWolf
	default:
		return "Bot"
	}
}

// waterCell draws a sea square, with a moving wave crest rolling through.
func (s styles) waterCell(row, col, frame int) string {
	if (row+col+frame)%7 == 0 {
		return s.waterWave.Render("~ ")
	}
	return s.water.Render("· ")
}

// hullBlock draws a ship square as part of a pointed vessel on the sea.
func (s styles) hullBlock(h game.Hull) string {
	switch h {
	case game.HullBowH:
		return s.shipHull.Render("◀█")
	case game.HullSternH:
		return s.shipHull.Render("█▶")
	case game.HullBowV:
		return s.shipHull.Render("▲▲")
	case game.HullSternV:
		return s.shipHull.Render("▼▼")
	case game.HullSingle:
		return s.shipHull.Render("◈ ")
	default: // HullMidH / HullMidV
		return s.shipHull.Render("██")
	}
}

// coordName turns a coord into its player-facing name, e.g. {0,0} -> "A1".
func coordName(c game.Coord) string {
	return fmt.Sprintf("%c%d", 'A'+c.Col, c.Row+1)
}

// record renders a win/loss tally to sit right after a name, wins green and
// losses red, e.g. "(1/2)".
func (s styles) record(wins, losses int) string {
	return s.dim.Render("(") + s.logGood.Render(fmt.Sprintf("%d", wins)) +
		s.dim.Render("/") + s.logHit.Render(fmt.Sprintf("%d", losses)) + s.dim.Render(")")
}

// wl renders a colored "W-L" for tables (wins green, losses red).
func (s styles) wl(wins, losses int) string {
	return s.logGood.Render(fmt.Sprintf("%d", wins)) + s.dim.Render("-") + s.logHit.Render(fmt.Sprintf("%d", losses))
}

// cellBlock renders one square as a 2-column colored block.
func (s styles) cellBlock(c game.Cell) string {
	switch c {
	case game.CellShip:
		return s.ship.Render("  ")
	case game.CellHit:
		return s.hit.Render("✖ ")
	case game.CellMiss:
		return s.miss.Render("○ ")
	case game.CellSunk:
		return s.sunk.Render("✖ ")
	default:
		return s.water.Render("· ")
	}
}

// renderBoard draws a 10x10 grid with A-J column and 1-10 row labels. Squares
// are drawn edge-to-edge (2 columns each) so ships and the sea look solid.
//
//   - aim, if set, highlights the targeting reticle (enemy board only).
//   - preview, if set, highlights where a ship is about to be placed;
//     previewValid tints it green (fits) or red (blocked).
func (s styles) renderBoard(g grid, aim *game.Coord, preview map[game.Coord]bool, previewValid bool, boom *boomOverlay, frame int, hull *hullGrid) string {
	var sb strings.Builder

	sb.WriteString("   ")
	for c := 0; c < game.BoardSize; c++ {
		sb.WriteString(s.dim.Render(fmt.Sprintf("%-2s", string(rune('A'+c)))))
	}
	sb.WriteString("\n")

	for r := 0; r < game.BoardSize; r++ {
		sb.WriteString(s.dim.Render(fmt.Sprintf("%2d ", r+1)))
		for c := 0; c < game.BoardSize; c++ {
			coord := game.Coord{Row: r, Col: c}
			cell := g[r][c]
			switch {
			case boom != nil && boom.Coord == coord && boom.Frame < len(explosionGlyphs):
				sb.WriteString(s.boom[boom.Frame].Render(explosionGlyphs[boom.Frame] + " "))
			case preview != nil && preview[coord]:
				if previewValid {
					sb.WriteString(s.previewOK.Render("  "))
				} else {
					sb.WriteString(s.previewBad.Render("  "))
				}
			case aim != nil && *aim == coord:
				sb.WriteString(s.aim.Render("◎ "))
			case cell == game.CellShip && hull != nil:
				sb.WriteString(s.hullBlock(hull[r][c]))
			case cell == game.CellEmpty:
				sb.WriteString(s.waterCell(r, c, frame))
			default:
				sb.WriteString(s.cellBlock(cell))
			}
		}
		sb.WriteString("\n")
	}

	return strings.TrimRight(sb.String(), "\n")
}

// framed wraps content in the outer bordered container, centered in width.
func (s styles) framed(width int, content string) string {
	boxed := s.frame.Render(content)
	if width <= 0 {
		return boxed
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, boxed)
}

// boardPanel stacks a caption above a bordered board.
func (s styles) boardPanel(caption string, board string) string {
	return lipgloss.JoinVertical(lipgloss.Center, s.dim.Render(caption), s.box.Render(board))
}

// header is the logo line shown on every screen.
func (s styles) header(tagline string) string {
	return s.logo.Render("🚢 TORPIDO") + "  " + s.tag.Render(tagline)
}

// legend explains the glyphs, using the real colored blocks as a key.
func (s styles) legend(ship, hit, miss string) string {
	return s.ship.Render("  ") + s.dim.Render(" "+ship+"   ") +
		s.hit.Render("✖ ") + s.dim.Render(" "+hit+"   ") +
		s.miss.Render("○ ") + s.dim.Render(" "+miss)
}

// screen wraps a screen body with the standard outer padding.
func screen(body string) string {
	return lipgloss.NewStyle().Padding(1, 2).Render(body)
}
