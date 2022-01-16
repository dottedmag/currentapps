package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"time"

	"github.com/ridge/must"
)

type lineType string

const (
	lineStarted      lineType = "started"
	lineApp          lineType = "app"
	lineIdle         lineType = "idle"
	lineNotIdle      lineType = "notIdle"
	lineScreenSleep  lineType = "screenSleep"
	lineScreenWakeUp lineType = "screenWakeUp"
	lineSleep        lineType = "sleep"
	lineWakeUp       lineType = "wakeUp"
)

type logLine struct {
	t   time.Time
	typ lineType
	app string
}

var rx = regexp.MustCompile(`^(\d+-\d+-\d+ \d+:\d+:\d+\.\d+) ` +
	`(Started|Screen sleep|Idle timer|Idle|Sleep|Screen wake up|Wake up|(Not idle) @ (\d+-\d+-\d+ \d+:\d+:\d+\.\d+)|(Application activated) ([a-zA-Z0-9.-]+))$`)

const timeFmt = "2006-01-02 15:04:05.9999"

func parseLogLine(s string) logLine {
	m := rx.FindStringSubmatch(s)
	if m == nil {
		panic(s)
	}
	t := must.Time(time.Parse(timeFmt, m[1]))
	switch {
	case m[2] == "Started":
		return logLine{t: t, typ: lineStarted}
	case m[2] == "Idle timer":
		return logLine{}
	case m[2] == "Idle":
		return logLine{t: t, typ: lineIdle}
	case m[2] == "Sleep":
		return logLine{t: t, typ: lineSleep}
	case m[2] == "Wake up":
		return logLine{t: t, typ: lineWakeUp}
	case m[2] == "Screen sleep":
		return logLine{t: t, typ: lineScreenSleep}
	case m[2] == "Screen wake up":
		return logLine{t: t, typ: lineScreenWakeUp}
	case m[3] == "Not idle":
		return logLine{t: must.Time(time.Parse(timeFmt, m[4])), typ: lineNotIdle}
	case m[5] == "Application activated":
		return logLine{t: t, typ: lineApp, app: m[6]}
	}
	panic(s)
}

type state string

const (
	unknown  state = "unknown"
	active   state = "active"
	idle     state = "idle"
	sleeping state = "sleeping"
)

func update(state state, app string, ll logLine) (state, string) {

	switch state {
	case unknown:
		switch ll.typ {
		case lineApp:
			return active, ll.app
		default:
			// cannot set any other state, because app is not known
			return state, app
		}
	case active:
		switch ll.typ {
		case lineApp:
			return active, ll.app
		case lineIdle, lineScreenWakeUp, lineWakeUp:
			return idle, app
		case lineNotIdle:
			return active, app
		case lineScreenSleep, lineSleep:
			return sleeping, app
		}
	case idle:
		switch ll.typ {
		case lineApp:
			return active, ll.app
		case lineIdle:
			return idle, app
		case lineNotIdle:
			return active, app
		case lineScreenSleep:
			return sleeping, app
		case lineScreenWakeUp:
			return idle, app
		case lineSleep:
			return sleeping, app
		case lineWakeUp:
			return idle, app
		}
	case sleeping:
		switch ll.typ {
		case lineApp:
			// switch app, but keep sleeping
			return sleeping, ll.app
		case lineIdle:
			return sleeping, app
		case lineNotIdle:
			return sleeping, app
		case lineScreenSleep:
			return sleeping, app
		case lineSleep:
			return sleeping, app
		case lineScreenWakeUp:
			return idle, app
		case lineWakeUp:
			return idle, app
		}
	}
	panic(fmt.Errorf("%s %s %v\n", state, app, ll))
}

var appsPerDay = map[time.Time]map[string]time.Duration{}

var timePerDay = map[time.Time]time.Duration{}

var prevto time.Time

func countActive(from, to time.Time, app string) {
	dayFrom := from.Truncate(24 * time.Hour)
	if appsPerDay[dayFrom] == nil {
		appsPerDay[dayFrom] = map[string]time.Duration{}
	}
	dayTo := to.Truncate(24 * time.Hour)
	if appsPerDay[dayTo] == nil {
		appsPerDay[dayTo] = map[string]time.Duration{}
	}

	if dayFrom.Equal(dayTo) {
		// simple case
		appsPerDay[dayFrom][app] += to.Sub(from)
		timePerDay[dayFrom] += to.Sub(from)
	} else {
		appsPerDay[dayFrom][app] += dayTo.Sub(from)
		appsPerDay[dayTo][app] += to.Sub(dayTo)

		timePerDay[dayFrom] += dayTo.Sub(from)
		timePerDay[dayTo] += to.Sub(dayTo)
	}
}

func main() {
	fh := must.OSFile(os.Open(os.Args[1]))
	defer fh.Close()

	var t time.Time
	state := unknown
	var app string

	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		ll := parseLogLine(scanner.Text())
		if ll.typ == "" || ll.typ == lineStarted {
			// ignore, nothing to do
			continue
		}

		newstate, newapp := update(state, app, ll)
		if state != newstate || app != newapp {
			if state == active && app != "com.apple.loginwindow" {
				countActive(t, ll.t, app)
			}
			if state == active && app == "com.apple.loginwindow" {
				countActive(t, ll.t, "**LOCKED**")
			}
			if state == idle {
				countActive(t, ll.t, "**IDLE**")
			}
			if state == sleeping {
				countActive(t, ll.t, "**SLEEPING**")
			}
			if state == unknown && !t.IsZero() {
				countActive(t, ll.t, "**UNKNOWN**")
			}

			state, app = newstate, newapp
			t = ll.t
		}
	}

	for d, apps := range appsPerDay {
		fmt.Printf("%s:\n", d.Format("2006-01-02"))
		type appdur struct {
			app string
			dur time.Duration
		}
		var appdurs []appdur
		for app, dur := range apps {
			appdurs = append(appdurs, appdur{app: app, dur: dur})
		}
		sort.Slice(appdurs, func(i, j int) bool {
			return appdurs[i].dur > appdurs[j].dur
		})
		var sumappdur time.Duration
		for _, appdur := range appdurs {
			sumappdur += appdur.dur
			if appdur.dur > 5*time.Minute {
				fmt.Printf("%40s %s\n", appdur.app, appdur.dur)
			}
		}
	}
}
