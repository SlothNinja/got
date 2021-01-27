package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/send"
	"github.com/SlothNinja/sn"
	"github.com/gin-gonic/gin"
	"github.com/mailjet/mailjet-apiv3-go"
)

type crmap map[*datastore.Key]*rating.CurrentRating

func (cl client) endGame(c *gin.Context, g *Game) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.finalClaim(c)
	ps, err := cl.determinePlaces(c, g)
	if err != nil {
		sn.JErr(c, err)
		return
	}
	g.setWinners(ps[0])

	cs := sn.GenContests(c, ps)
	g.Status = sn.Completed

	stats, err := cl.updateUStats(c, g)
	if err != nil {
		sn.JErr(c, err)
		return
	}

	crs := make(crmap, len(g.UserKeys))
	for _, ukey := range g.UserKeys {
		crs[ukey], err = cl.SN.GetProjectedRating(c, ukey, g.Type)
		if err != nil {
			sn.JErr(c, err)
			return
		}
	}

	nrs := make(crmap, len(g.UserKeys))
	for _, ukey := range g.UserKeys {
		nrs[ukey], err = crs[ukey].Projected(cs[ukey])
		if err != nil {
			sn.JErr(c, err)
			return
		}
	}

	_, err = cl.DS.RunInTransaction(c, func(tx *datastore.Transaction) error {
		g.Undo.Commit()
		ks, es := g.save()
		for _, contests := range cs {
			for _, contest := range contests {
				ks = append(ks, contest.Key)
				es = append(es, contest)
			}
		}
		for _, stat := range stats {
			ks = append(ks, stat.Key)
			es = append(es, stat)
		}
		_, err := tx.PutMulti(ks, es)
		return err
	})
	if err != nil {
		sn.JErr(c, err)
		return
	}

	// Need to call SendTurnNotificationsTo before saving the new contests
	// SendEndGameNotifications relies on pulling the old contests from the db.
	// Saving the contests resulting in double counting.
	err = cl.sendEndGameNotifications(c, g, ps, crs, nrs)
	if err != nil {
		// log but otherwise ignore send errors
		log.Warningf(err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"game": g})

}

func (g *Game) setWinners(rmap contest.ResultsMap) {
	g.Status = sn.Completed

	g.setCurrentPlayer(nil)
	g.WinnerKeys = nil
	for k := range rmap {
		p := g.playerByUserKey(k)
		g.WinnerKeys = append(g.WinnerKeys, p.User.Key)
	}
}

type result struct {
	Place, GLO, Score int
	Name, Inc         string
}

type results []result

func (cl client) sendEndGameNotifications(c *gin.Context, g *Game, ps contest.Places, crs, nrs crmap) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Status = sn.Completed
	rs := make(results, g.NumPlayers)

	var i int
	for place, rmap := range ps {
		for k := range rmap {
			p := g.playerByUserID(k.ID)
			cr, nr := crs[k], nrs[k]
			clo, nlo := cr.Rank().GLO(), nr.Rank().GLO()
			inc := nlo - clo

			rs[i] = result{
				Place: place,
				GLO:   nlo,
				Score: p.Score,
				Name:  p.User.Name,
				Inc:   fmt.Sprintf("%+d", inc),
			}
		}
		i++
	}

	var names []string
	for _, p := range g.winners() {
		names = append(names, p.User.Name)
	}

	buf := new(bytes.Buffer)
	tmpl := template.New("end_game_notification")
	tmpl, err := tmpl.Parse(`
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
<html>
        <head>
                <meta http-equiv="content-type" content="text/html; charset=ISO-8859-1">
        </head>
        <body bgcolor="#ffffff" text="#000000">
                {{range $i, $r := $.Results}}
                <div style="height:3em">
                        <div style="height:3em;float:left;padding-right:1em">{{$r.Place}}.</div>
                        <div style="height:1em">{{$r.Name}} scored {{$r.Score}} points.</div>
                        <div style="height:1em">Glicko {{$r.Inc}} (-> {{$r.GLO}})</div>
                </div>
                {{end}}
                <p></p>
                <p>Congratulations: {{$.Winners}}.</p>
        </body>
</html>`)
	if err != nil {
		return err
	}

	err = tmpl.Execute(buf, gin.H{
		"Results": rs,
		"Winners": restful.ToSentence(names),
	})
	if err != nil {
		return err
	}

	ms := make([]mailjet.InfoMessagesV31, len(g.players))
	subject := fmt.Sprintf("SlothNinja Games: Guild of Thieves #%d Has Ended", g.id())
	body := buf.String()
	for i, p := range g.players {
		ms[i] = mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "webmaster@slothninja.com",
				Name:  "Webmaster",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: g.EmailFor(p.User.ID()),
					Name:  p.User.Name,
				},
			},
			Subject:  subject,
			HTMLPart: body,
		}
	}
	_, err = send.Messages(c, ms...)
	return err
}

func (g *Game) winners() Players {
	l := len(g.WinnerKeys)
	if l == 0 {
		return nil

	}
	ps := make(Players, l)
	for i, k := range g.WinnerKeys {
		ps[i] = g.playerByUserKey(k)
	}
	return ps
}
