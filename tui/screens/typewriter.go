package screens

import "time"

type Typewriter struct {
	fullText string
	charRate time.Duration
	start    time.Time
	revealed int
	started  bool
	skipped  bool
}

func NewTypewriter(text string, charRate time.Duration) *Typewriter {
	tw := &Typewriter{
		fullText: text,
		charRate: charRate,
	}
	if charRate <= 0 {
		tw.Skip()
	}
	return tw
}

func (tw *Typewriter) Start(now time.Time) {
	if !tw.started {
		tw.start = now
		tw.started = true
	}
}

func (tw *Typewriter) Update(now time.Time) {
	if tw.skipped || !tw.started {
		return
	}
	elapsed := now.Sub(tw.start)
	tw.revealed = int(elapsed / tw.charRate)
	if tw.revealed > len(tw.fullText) {
		tw.revealed = len(tw.fullText)
	}
}

func (tw *Typewriter) View() string {
	if tw.skipped || !tw.started {
		if tw.skipped {
			return tw.fullText
		}
		return ""
	}
	return tw.fullText[:tw.revealed]
}

func (tw *Typewriter) Skip() {
	tw.skipped = true
	tw.revealed = len(tw.fullText)
}

func (tw *Typewriter) Done() bool {
	return tw.skipped || tw.revealed >= len(tw.fullText)
}
