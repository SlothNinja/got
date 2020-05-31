package main

import (
	"bytes"
	"fmt"
	"html/template"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/send"
	"github.com/SlothNinja/sn/v2"
	"github.com/gin-gonic/gin"
	"github.com/mailjet/mailjet-apiv3-go"
)

func (cl client) endGame(c *gin.Context, g *game) (sn.Places, error) {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	ps, err := cl.determinePlaces(c, g)
	if err != nil {
		return nil, err
	}
	g.setWinners(ps[0])
	return ps, nil
}

func (g *game) setWinners(rmap sn.ResultsMap) {
	g.Status = sn.Completed

	g.setCurrentPlayer(nil)
	g.WinnerIDS = nil
	for key := range rmap {
		p := g.playerByUserID(key.ID)
		g.WinnerIDS = append(g.WinnerIDS, p.ID)
	}
}

type result struct {
	Place, GLO, Score int
	Name, Inc         string
}

type results []result

func (cl client) sendEndGameNotifications(c *gin.Context, g *game, ps sn.Places, cs []*sn.Contest) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	g.Status = sn.Completed
	rs := make(results, g.NumPlayers)

	var i int
	for place, rmap := range ps {
		for k := range rmap {
			p := g.playerByUserID(k.ID)
			cr, nr, err := cl.Game.IncreaseFor(c, p.User.Key, g.Type, cs)
			if err != nil {
				log.Warningf(err.Error())
			}
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

func (g *game) winners() Players {
	l := len(g.WinnerIDS)
	if l == 0 {
		return nil

	}
	ps := make(Players, l)
	for i, pid := range g.WinnerIDS {
		ps[i] = g.playerByID(pid)
	}
	return ps
}
