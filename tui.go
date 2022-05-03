package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bmedicke/bhdr/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

//go:embed chordmap.json
var chordmapJSON string

// TODO: read times from config.json
const (
	pomodoroDuration time.Duration = time.Minute * 1
	breakDuration    time.Duration = time.Minute * 1
)

func spawnTUI() {
	// TODO: create commandChannel for handlePomodoroState().
	app := tview.NewApplication()
	pom := createPomodoro(pomodoroDuration, breakDuration)
	chord := util.KeyChord{Active: false, Buffer: "", Action: ""}

	chordmap := map[string]interface{}{}
	json.Unmarshal([]byte(chordmapJSON), &chordmap)

	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	frame := tview.NewFrame(layout)
	header := tview.NewFlex()
	headerleft := tview.NewTextView()
	headercenter := tview.NewTextView()
	headerright := tview.NewTextView()
	body := tview.NewTable()
	statusbar := tview.NewTextView()

	frame.AddText(" P 🐕 M ", true, tview.AlignCenter, tcell.ColorLime)
	frame.SetBorders(0, 0, 0, 0, 0, 0)
	frame.SetBackgroundColor(tcell.Color236)

	layout.AddItem(header, 1, 0, false)
	layout.AddItem(body, 0, 1, true)
	layout.AddItem(statusbar, 1, 0, false)

	header.SetBorderPadding(0, 0, 0, 0)
	header.AddItem(headerleft, 0, 2, false)
	header.AddItem(headercenter, 0, 1, false)
	header.AddItem(headerright, 17, 0, false)

	headerleft.SetChangedFunc(func() { app.Draw() })
	headerright.SetChangedFunc(func() { app.Draw() })

	statusbar.SetBackgroundColor(tcell.ColorDarkOliveGreen)
	statusbar.SetBorderPadding(0, 0, 0, 0)
	statusbar.SetChangedFunc(func() { app.Draw() })

	body.SetSelectable(true, true)

	body.SetInputCapture(
		func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				util.ResetChord(&chord)
			}

			if chord.Active {
				util.HandleChords(event.Rune(), &chord, chordmap)
				handleAction(chord.Action, &pom)
			} else {
				switch event.Rune() {
				case 'q':
					app.Stop()
				case 'c', 'd': // start chord:
					util.HandleChords(event.Rune(), &chord, chordmap)
				}
			}
			statusbar.SetText(chord.Buffer)
			return event
		},
	)

	updateBody(body)
	go handlePomodoroState(&pom, statusbar)
	go updateHeader(headerleft, headercenter, headerright, &pom)

	app.SetRoot(frame, true)
	app.SetFocus(body).Run()
}

type pomodoro struct {
	State          string
	StartTime      time.Time
	StopTime       time.Time
	PomDuration    time.Duration
	DurationLeft   time.Duration
	BreakStartTime time.Time
	BreakStopTime  time.Time
	BreakDuration  time.Duration
	Waiting        bool
	CurrentTask    string
}

func createPomodoro(
	duration time.Duration,
	breakDuration time.Duration,
) pomodoro {
	pom := pomodoro{
		State:         "ready",
		PomDuration:   duration,
		DurationLeft:  duration,
		BreakDuration: breakDuration,
		Waiting:       true,
	}
	return pom
}

func updateBody(view *tview.Table) {
	// TODO: read-only from struct, update via commandChannel (task).
	b := []map[string]string{
		{"id": "current task", "value": "research"},
		{"id": "server", "value": "0.0.0.0:8421/api"},
	}
	cols, rows := 3, len(b)

	for row := 0; row < rows; row++ {
		for col := 0; col < cols; col++ {
			var s string
			switch col {
			case 0:
				s = b[row]["id"]
			case 1:
				s = "  ==  "
			case 2:
				s = b[row]["value"]
			}
			cell := tview.NewTableCell(s)
			if col < 2 {
				cell.SetSelectable(false)
			}
			view.SetCell(row, col, cell)
		}
	}
}

// TODO: move this to util.
func attachTicker(timer chan time.Time, interval time.Duration) {
	timer <- time.Now() // send one tick immediately.
	t := time.NewTicker(interval)
	for c := range t.C {
		timer <- c
	}
}

func handleAction(action string, pom *pomodoro) {
	switch action {
	case "continue":
		// TODO send signal instead of mutating state directly! (commandChannel)
		(*pom).Waiting = false // TODO READ ONLY!
	case "create_pomodoro":
	case "create_break":
	case "cancel":
	case "delete_pomodoro":
	case "delete_break":
	}
}

func updateHeader(
	left *tview.TextView,
	center *tview.TextView,
	right *tview.TextView,
	pom *pomodoro,
) {
	tick := make(chan time.Time)
	go attachTicker(tick, time.Millisecond * 200)

	for {
		<-tick
		timeleft := (*pom).DurationLeft.Round(time.Second)
		// TODO left pad text:
		right.SetText(fmt.Sprintf("%v [%v]", (*pom).State, timeleft))

		var color tcell.Color
		switch (*pom).State {
		case "work":
			color = tcell.ColorDarkOliveGreen
		case "break":
			color = tcell.ColorBlue
		case "break_done", "work_done":
			color = tcell.ColorDarkRed
		}

		left.SetBackgroundColor(color)
		center.SetBackgroundColor(color)
		right.SetBackgroundColor(color)
	}
}
