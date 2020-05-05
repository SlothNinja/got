package main

import (
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
)

func (client Client) placeThief(c *gin.Context) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g, err := client.getGame(c)
	if err != nil {
		jerr(c, err)
		return
	}

	err = g.placeThief(c)
	if err != nil {
		jerr(c, err)
		return
	}

	_, err = client.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		_, err := tx.PutMulti(g.cache())
		return err
	})
	if err != nil {
		jerr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"game": g})
}

func (g *Game) placeThieves(c *gin.Context) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Phase = placeThieves
	return nil
}

func (g *Game) placeThief(c *gin.Context) error {
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

func (g *Game) validatePlaceThief(c *gin.Context) (*Area, error) {
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
