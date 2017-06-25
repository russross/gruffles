package main

import "bytes"

type pair struct {
	x, y int
}

func GetMap(state *State, current *Room, visited []bool) string {
	depth := 3
	var buf bytes.Buffer
	text := trace(state, current, visited, depth)
	for y := depth*4 + 2; y >= -depth*4-2; y-- {
		for x := -depth*4 - 2; x <= depth*4+2; x++ {
			if ch, present := text[pair{x, y}]; present {
				buf.WriteRune(ch)
			} else {
				buf.WriteRune(' ')
			}
		}
		buf.WriteString("\n")
	}
	return buf.String()
}

func trace(state *State, start *Room, visited []bool, depth int) map[pair]rune {
	q := []pair{pair{0, 0}}
	grid := map[pair]*Room{pair{0, 0}: start}
	text := make(map[pair]rune)

	for len(q) > 0 {
		here := q[0]
		room := grid[here]
		q = q[1:]
		if here.x < -depth || here.x > depth || here.y < -depth || here.y > depth {
			continue
		}

		// draw the room
		x, y := here.x*4, here.y*4
		if here.x == 0 && here.y == 0 {
			text[pair{x, y}] = '╳'
		} else {
			text[pair{x, y}] = ' '
		}
		text[pair{x - 1, y}] = '│'
		text[pair{x + 1, y}] = '│'
		text[pair{x, y + 1}] = '─'
		text[pair{x, y - 1}] = '─'
		text[pair{x - 1, y + 1}] = '╭'
		text[pair{x + 1, y + 1}] = '╮'
		text[pair{x - 1, y - 1}] = '╰'
		text[pair{x + 1, y - 1}] = '╯'

		// draw the exits and follow them
		handleDir := func(forward, reverse rune, bi, uni, out, new rune, dx, dy int) {
			target := room.Exit(state, forward)
			if target == nil {
				return
			}
			seen := visited[target.ID]
			existing := grid[pair{here.x + dx, here.y + dy}]
			back := target.Exit(state, reverse)

			switch {
			case room.Zone() != target.Zone():
				text[pair{x + dx, y + dy}] = out
			case !seen:
				text[pair{x + dx, y + dy}] = new
			case existing == nil:
				grid[pair{here.x + dx, here.y + dy}] = target
				q = append(q, pair{here.x + dx, here.y + dy})

				fallthrough
			case existing == target:
				if room == back {
					text[pair{x + dx + dx, y + dy + dy}] = bi
				} else {
					text[pair{x + dx + dx, y + dy + dy}] = uni
				}
			default:
				text[pair{x + dx, y + dy}] = uni
			}
		}

		handleDir('n', 's', '↕', '↑', '⇑', '⇡', 0, 1)
		handleDir('s', 'n', '↕', '↓', '⇓', '⇣', 0, -1)
		handleDir('e', 'w', '↔', '→', '⇒', '⇢', 1, 0)
		handleDir('w', 'e', '↔', '←', '⇐', '⇠', -1, 0)

		if target := room.Exit(state, 'u'); target != nil {
			if room.Zone() != target.Zone() {
				text[pair{x + 1, y + 1}] = '⇗'
			} else if !visited[target.ID] {
				text[pair{x + 1, y + 1}] = '⤴'
			} else {
				text[pair{x + 1, y + 1}] = '↗'
			}
		}
		if target := room.Exit(state, 'd'); target != nil {
			if room.Zone() != target.Zone() {
				text[pair{x - 1, y - 1}] = '⇙'
			} else if !visited[target.ID] {
				text[pair{x - 1, y - 1}] = '⤶'
			} else {
				text[pair{x - 1, y - 1}] = '↙'
			}
		}
	}

	return text
}
