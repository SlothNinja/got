package main

import "time"

type Log []Entry

func (g *Game) newEntry(m ...Message) {
	g.Log = append(g.Log, Entry{
		Messages:  append([]Message(nil), m...),
		Turn:      g.Turn,
		Rev:       g.Rev(),
		UpdatedAt: time.Now(),
	})
}

func (g *Game) newEntryFor(pid int, m ...Message) {
	g.Log = append(g.Log, Entry{
		Messages:  append([]Message(nil), m...),
		PID:       pid,
		Turn:      g.Turn,
		Rev:       g.Rev(),
		UpdatedAt: time.Now(),
	})
}

type Entry struct {
	Messages  []Message `json:"messages"`
	PID       int       `json:"pid,omitempty"`
	Turn      int       `json:"turn"`
	Rev       int64     `json:"rev"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (g *Game) appendEntry(m ...Message) {
	l := len(g.Log)
	if l == 0 {
		return
	}

	last := l - 1
	g.Log[last].Messages = append(g.Log[last].Messages, m...)
	g.Log[last].UpdatedAt = time.Now()
	g.Log[last].Rev = g.Rev()
}

type Message map[string]interface{}
