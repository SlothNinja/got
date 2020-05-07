package main

import (
	"fmt"
	"net/http"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (client Client) placeThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getHistory(c, 0)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.placeThief(c)
	if err != nil {
		jerr(c, err)
		return
	}

	ks, es := g.cache()
	log.Debugf("ks: %v", ks)
	_, err = client.DS.Put(c, ks, es)
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *History) placeThieves(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = placeThieves
	return nil
}

func (g *History) placeThief(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	a, err := g.validatePlaceThief(c)
	if err != nil {
		return err
	}

	cp := g.CurrentPlayer()
	cp.PerformedAction = true
	cp.Score += a.Card.Value()
	a.Thief = cp.ID

	g.Undo.Update()
	// Log placement
	// e := g.newPlaceThiefEntryFor(cp)
	// restful.AddNoticef(c, string(e.HTML(g)))
	return nil
}

func (g *History) validatePlaceThief(c *gin.Context) (*Area, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	err := g.validatePlayerAction(c)
	if err != nil {
		return nil, err
	}

	a, err := g.getAreaFrom(c)
	if err != nil {
		return nil, err
	}

	switch {
	case a == nil:
		return nil, fmt.Errorf("you must select an area: %w", sn.ErrValidation)
	case a.Card == nil:
		return nil, fmt.Errorf("you must select an area with a card: %w", sn.ErrValidation)
	case a.Thief != noPID:
		return nil, fmt.Errorf("you must select an area without a thief: %w", sn.ErrValidation)
	default:
		return a, nil
	}
}

// type placeThiefEntry struct {
// 	*Entry
// 	Area Area
// }
//
// func (g *History) newPlaceThiefEntryFor(p *Player) (e *placeThiefEntry) {
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
// func (e *placeThiefEntry) HTML(g *History) template.HTML {
// 	return restful.HTML("%s placed thief on %s at %s%s.",
// 		g.NameByPID(e.PlayerID), e.Area.Card.Type, e.Area.RowString(), e.Area.ColString())
// }
