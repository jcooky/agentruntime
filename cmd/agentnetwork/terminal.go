package main

import (
	"context"
	"github.com/gdamore/tcell/v2"
)

// printText prints a string starting at the specified (x, y) coordinates on the screen.
func printText(s tcell.Screen, x, y int, str string) {
	for i, r := range str {
		s.SetContent(x+i, y, r, nil, tcell.StyleDefault)
	}
	s.Show() // Update the screen to reflect the changes
}

type ListScreenRequest struct {
	startMessage  string
	noMoreMessage string
}

func listScreen(ctx context.Context, screen tcell.Screen, req ListScreenRequest, fetch func() ([]string, error)) error {
	if req.startMessage == "" {
		req.startMessage = "Press Enter to load more data. Press ESC to exit."
	}
	if req.noMoreMessage == "" {
		req.noMoreMessage = "No more data to load."
	}

	defStyle := tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset)
	screen.SetStyle(defStyle)
	screen.Clear()

	printText(screen, 0, 0, req.startMessage)

	messages, err := fetch()
	if err != nil {
		return err
	}
	var lastY = 3
	for idx, msg := range messages {
		printText(screen, 0, lastY+idx, msg)
	}
	lastY += len(messages)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	eventCh := make(chan tcell.Event)
	go screen.ChannelEvents(eventCh, ctx.Done())

	for ev := range eventCh {
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyCtrlD, tcell.KeyCtrlC, tcell.KeyRune:
				if ev.Key() == tcell.KeyRune && ev.Rune() != 'Q' && ev.Rune() != 'q' {
					break
				}
				return nil
			case tcell.KeyEnter:
				messages, err = fetch()
				if err != nil {
					return err
				}
				if len(messages) == 0 {
					printText(screen, 0, lastY, req.noMoreMessage)
					break
				}

				for idx, msg := range messages {
					printText(screen, 0, lastY+idx, msg)
				}
				lastY += len(messages)
			default:
				// Other key events can be handled here if needed
			}
		case *tcell.EventResize:
			// Synchronize the screen when a resize event occurs
			screen.Sync()
		}
	}

	return nil
}
