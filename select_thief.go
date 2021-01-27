package got

import (
	"github.com/SlothNinja/sn"
)

func (client *Client) startSelectThief() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	client.Game.Phase = selectThief
}

func (client *Client) selectThief() {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	err := client.validateSelectThief()
	if err != nil {
		client.flashError(err)
		return
	}

	client.Game.SelectedThiefAreaF = client.Game.SelectedArea()
	client.startMoveThief()
}

func (client *Client) validateSelectThief() error {
	client.Log.Debugf(msgEnter)
	defer client.Log.Debugf(msgExit)

	g := client.Game
	area, err := g.SelectedArea(), client.validatePlayerAction()
	switch {
	case err != nil:
		return err
	case area == nil || area.Thief != g.CurrentPlayer().ID():
		return sn.NewVError("You must select one of your thieves.")
	default:
		return nil
	}
}
