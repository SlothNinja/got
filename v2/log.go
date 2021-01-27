package main

import (
	"encoding/json"
	"time"
)

type glog []entry

func (g *Game) newEntry(m ...message) {
	g.glog = append(g.glog, entry{
		messages:  append([]message(nil), m...),
		turn:      g.Turn,
		rev:       g.rev(),
		updatedAt: time.Now(),
	})
}

func (g *Game) newEntryFor(pid int, m ...message) {
	g.glog = append(g.glog, entry{
		messages:  append([]message(nil), m...),
		pid:       pid,
		turn:      g.Turn,
		rev:       g.rev(),
		updatedAt: time.Now(),
	})
}

type entry struct {
	messages  []message
	pid       int
	turn      int
	rev       int64
	updatedAt time.Time
}

type jEntry struct {
	Messages  []message `json:"messages"`
	PID       int       `json:"pid,omitempty"`
	Turn      int       `json:"turn"`
	Rev       int64     `json:"rev"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (e entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(jEntry{
		Messages:  e.messages,
		PID:       e.pid,
		Turn:      e.turn,
		Rev:       e.rev,
		UpdatedAt: e.updatedAt,
	})
}

func (e *entry) UnmarshalJSON(bs []byte) error {
	var obj jEntry
	err := json.Unmarshal(bs, &obj)
	if err != nil {
		return err
	}
	e.messages = obj.Messages
	e.pid = obj.PID
	e.turn = obj.Turn
	e.rev = obj.Rev
	e.updatedAt = obj.UpdatedAt
	return nil
}

func (g *Game) appendEntry(m ...message) {
	l := len(g.glog)
	if l == 0 {
		return
	}

	last := l - 1
	g.glog[last].messages = append(g.glog[last].messages, m...)
	g.glog[last].updatedAt = time.Now()
	g.glog[last].rev = g.rev()
}

type message map[string]interface{}
