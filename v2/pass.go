package main

import (
	"github.com/SlothNinja/log"
	"github.com/gin-gonic/gin"
)

func (g *History) pass(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validatePass(c)
	if err != nil {
		return err
	}

	cp := g.CurrentPlayer()
	cp.Passed = true
	cp.PerformedAction = true
	g.Phase = drawCard

	// Log Pass
	// e := g.newPassEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))

	return nil
}

func (g *History) validatePass(c *gin.Context) error {
	err := g.validatePlayerAction(c)
	if err != nil {
		return err
	}
	return nil
}

// type passEntry struct {
// 	*Entry
// }
//
// func (g *History) newPassEntryFor(p *Player) (e *passEntry) {
// 	e = &passEntry{
// 		Entry: g.newEntryFor(p),
// 	}
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return
// }
//
// func (e *passEntry) HTML(g *History) template.HTML {
// 	return restful.HTML("%s passed.", g.NameByPID(e.PlayerID))
// }
