package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
)

func (g *Game) placeThieves(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = placeThieves
	return nil
}

func (g *Game) placeThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validatePlaceThief(c)
	if err != nil {
		return err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.Score += g.SelectedArea().Card.Value()
	g.SelectedArea().Thief = cp.ID

	// Log placement
	// e := g.newPlaceThiefEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))
	return nil
}

func (g *Game) validatePlaceThief(c *gin.Context) error {
	err := g.validatePlayerAction(c)
	if err != nil {
		return err
	}

	area := g.SelectedArea()
	switch {
	case area == nil:
		return fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case area.Card == nil:
		return fmt.Errorf("you must select an area with a card: %w", sn.ErrValidation)
	case area.Thief != noPID:
		return fmt.Errorf("you must select an area without a thief: %w", sn.ErrValidation)
	default:
		return nil
	}
}

// type placeThiefEntry struct {
// 	*Entry
// 	Area Area
// }
//
// func (g *Game) newPlaceThiefEntryFor(p *Player) (e *placeThiefEntry) {
// 	area := g.SelectedArea()
// 	e = &placeThiefEntry{
// 		Entry: g.newEntryFor(p),
// 		Area:  *area,
// 	}
// 	p.Log = append(p.Log, e)
// 	g.Log = append(g.Log, e)
// 	return
// }
//
// func (e *placeThiefEntry) HTML(g *Game) template.HTML {
// 	return restful.HTML("%s placed thief on %s at %s%s.",
// 		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
// }
