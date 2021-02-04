package main

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"cloud.google.com/go/datastore"
	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/send"
	"github.com/gin-gonic/gin"
	"github.com/mailjet/mailjet-apiv3-go"
)

type crmap map[*datastore.Key]*rating.CurrentRating

func (cl *client) endGame() {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	cl.finalClaim()
	ps, err := cl.determinePlaces()
	if err != nil {
		cl.jerr(err)
		return
	}
	cl.setWinners(ps[0])

	cs := cl.Rating.Contest.GenContests(ps)
	cl.g.Status = game.Completed

	stats, err := cl.updateUStats()
	if err != nil {
		cl.jerr(err)
		return
	}

	crs := make(crmap, len(cl.g.UserKeys))
	for _, ukey := range cl.g.UserKeys {
		crs[ukey], err = cl.Rating.GetProjected(cl.ctx, ukey, cl.g.Type)
		if err != nil {
			cl.jerr(err)
			return
		}
	}

	nrs := make(crmap, len(cl.g.UserKeys))
	for _, ukey := range cl.g.UserKeys {
		nrs[ukey], err = crs[ukey].Projected(cs[ukey])
		if err != nil {
			cl.jerr(err)
			return
		}
	}

	_, err = cl.DS.RunInTransaction(cl.ctx, func(tx *datastore.Transaction) error {
		cl.g.Undo.Commit()
		ks, es := cl.g.save()
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
		cl.jerr(err)
		return
	}

	// Need to call SendTurnNotificationsTo before saving the new contests
	// SendEndGameNotifications relies on pulling the old contests from the db.
	// Saving the contests resulting in double counting.
	err = cl.sendEndGameNotifications(ps, crs, nrs)
	if err != nil {
		// log but otherwise ignore send errors
		cl.Log.Warningf(err.Error())
	}
	cl.ctx.JSON(http.StatusOK, gin.H{"game": cl.g})

}

func (cl *client) setWinners(rmap contest.ResultsMap) {
	cl.g.Status = game.Completed

	cl.setCurrentPlayer(nil)
	cl.g.WinnerKeys = nil
	for k := range rmap {
		p := cl.playerByUserKey(k)
		cl.g.WinnerKeys = append(cl.g.WinnerKeys, p.User.Key)
	}
}

type result struct {
	Place, GLO, Score int
	Name, Inc         string
}

type results []result

func (cl *client) sendEndGameNotifications(ps []contest.ResultsMap, crs, nrs crmap) error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	if cl.g == nil {
		return errors.New("cl.g was nil")
	}

	cl.g.Status = game.Completed
	rs := make(results, cl.g.NumPlayers)

	var i int
	for place, rmap := range ps {
		for k := range rmap {
			p := cl.playerByUserID(k.ID)
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
	for _, p := range cl.winners() {
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

	ms := make([]mailjet.InfoMessagesV31, len(cl.g.players))
	subject := fmt.Sprintf("SlothNinja Games: Guild of Thieves #%d Has Ended", cl.g.id())
	body := buf.String()
	for i, p := range cl.g.players {
		ms[i] = mailjet.InfoMessagesV31{
			From: &mailjet.RecipientV31{
				Email: "webmaster@slothninja.com",
				Name:  "Webmaster",
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: cl.emailFor(p),
					Name:  cl.nameFor(p),
				},
			},
			Subject:  subject,
			HTMLPart: body,
		}
	}
	_, err = send.Messages(cl.ctx, ms...)
	return err
}

func (cl *client) winners() []*player {
	if cl.g == nil {
		cl.Log.Warningf("cl.g was nil")
		return nil
	}

	l := len(cl.g.WinnerKeys)
	if l == 0 {
		return nil

	}
	ps := make([]*player, l)
	for i, k := range cl.g.WinnerKeys {
		ps[i] = cl.playerByUserKey(k)
	}
	return ps
}
