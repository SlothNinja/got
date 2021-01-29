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

	cp := client.Game.CurrentPlayer()
	cp.Passed = true
	cp.PerformedAction = true
	client.Game.Phase = drawCard

	// Log Pass
	e := client.newPassEntryFor(cp)
	restful.AddNoticef(client.Context, string(e.HTML(client.Game)))

	client.html("got/pass_update")
}

func (client *Client) validatePass() error {
	return client.validatePlayerAction()
}

type passEntry struct {
	*Entry
}

func (client *Client) newPassEntryFor(p *Player) *passEntry {
	e := &passEntry{
		Entry: client.newEntryFor(p),
	}
	p.Log = append(p.Log, e)
	client.Game.Log = append(client.Game.Log, e)
	return e
}

func (e *passEntry) HTML(g *Game) template.HTML {
	return restful.HTML("%s passed.", g.NameByPID(e.PlayerID))
}
