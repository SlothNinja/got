package got

import (
	"encoding/gob"
	"html/template"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
)

func init() {
	gob.Register(new(passEntry))
}

func (client *Client) pass() {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := client.validatePass()
	if err != nil {
		client.flashError(err)
		return
	}

	g := client.Game
	cp := g.CurrentPlayer()
	cp.Passed = true
	cp.PerformedAction = true
	g.Phase = drawCard

	// Log Pass
	e := g.newPassEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(g)))

	client.html("got/pass_update")
}

func (client *Client) validatePass() error {
	return client.validatePlayerAction()
}

type passEntry struct {
	*Entry
}

func (g *Game) newPassEntryFor(p *Player) (e *passEntry) {
	e = &passEntry{
		Entry: g.newEntryFor(p),
	}
	p.Log = append(p.Log, e)
	g.Log = append(g.Log, e)
	return
}

func (e *passEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s passed.", g.NameByPID(e.PlayerID))
}
