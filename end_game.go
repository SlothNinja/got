package got

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"html/template"

	"github.com/SlothNinja/contest"
	"github.com/SlothNinja/game"
	"github.com/SlothNinja/log"
	"github.com/SlothNinja/rating"
	"github.com/SlothNinja/restful"
	"github.com/SlothNinja/send"
	"github.com/gin-gonic/gin"
	"google.golang.org/appengine/mail"
)

func init() {
	gob.Register(new(endGameEntry))
	gob.Register(new(announceWinnersEntry))
}

func (g *Game) endGame(c *gin.Context) (ps contest.Places) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = endGame
	ps = g.determinePlaces(c)
	g.setWinners(ps[0])
	g.newEndGameEntry()
	return
}

func toIDS(places []Players) [][]int64 {
	sids := make([][]int64, len(places))
	for i, players := range places {
		for _, p := range players {
			sids[i] = append(sids[i], p.User().ID())
		}
	}
	return sids
}

type endGameEntry struct {
	*Entry
}

func (g *Game) newEndGameEntry() {
	e := &endGameEntry{
		Entry: g.newEntry(),
	}
	g.Log = append(g.Log, e)
}

func (e *endGameEntry) HTML(g *Game) (s template.HTML) {
	rows := restful.HTML("")
	for _, p := range g.Players() {
		rows += restful.HTML("<tr>")
		rows += restful.HTML("<td>%s</td> <td>%d</td> <td>%d</td> <td>%d</td> <td>%d</td>",
			g.NameFor(p), p.Score, lampCount(p.Hand...), camelCount(p.Hand...), len(p.Hand))
		rows += restful.HTML("</tr>")
	}
	s += restful.HTML("<table class='strippedDataTable'><thead><tr><th>Player</th><th>Score</th>")
	s += restful.HTML("<th>Lamps</th><th>Camels</th><th>Cards</th></tr></thead><tbody>")
	s += rows
	s += restful.HTML("</tbody></table>")
	return
}

func (g *Game) setWinners(rmap contest.ResultsMap) {
	g.Phase = announceWinners
	g.Status = game.Completed

	g.setCurrentPlayers()
	g.WinnerIDS = nil
	for key := range rmap {
		p := g.PlayerByUserID(key.ID)
		g.WinnerIDS = append(g.WinnerIDS, p.ID())
	}

	g.newAnnounceWinnersEntry()
}

type result struct {
	Place, GLO, Score int
	Name, Inc         string
}

type results []result

func (g *Game) sendEndGameNotifications(c *gin.Context, ps contest.Places, cs contest.Contests) (err error) {
	log.Debugf("Entering")
	defer log.Debugf("Exiting")

	g.Phase = gameOver
	g.Status = game.Completed
	rs := make(results, g.NumPlayers)

	var (
		i      int
		cr, nr *rating.CurrentRating
	)
	for place, rmap := range ps {
		for k := range rmap {
			p := g.PlayerByUserID(k.ID)
			cr, nr, err = rating.IncreaseFor(c, p.User(), g.Type, cs)
			clo, nlo := cr.Rank().GLO(), nr.Rank().GLO()
			inc := nlo - clo

			rs[i] = result{
				Place: place,
				GLO:   nlo,
				Score: p.Score,
				Name:  g.NameFor(p),
				Inc:   fmt.Sprintf("%+d", inc),
			}
		}
		i++
	}

	var names []string
	for _, p := range g.winners() {
		names = append(names, g.NameFor(p))
	}

	ts := restful.TemplatesFrom(c)
	buf := new(bytes.Buffer)
	tmpl := ts["got/end_game_notification"]
	if err = tmpl.Execute(buf, gin.H{
		"Results": rs,
		"Winners": restful.ToSentence(names),
	}); err != nil {
		return
	}

	ms := make([]*mail.Message, len(g.Players()))
	sender := "webmaster@slothninja.com"
	subject := fmt.Sprintf("SlothNinja Games: Guild of Thieves #%d Has Ended", g.ID)
	body := buf.String()
	for i, p := range g.Players() {
		ms[i] = &mail.Message{
			To:       []string{p.User().Email},
			Sender:   sender,
			Subject:  subject,
			HTMLBody: body,
		}
	}
	err = send.Message(c, ms...)
	return
}

type announceWinnersEntry struct {
	*Entry
}

func (g *Game) newAnnounceWinnersEntry() *announceWinnersEntry {
	e := &announceWinnersEntry{
		Entry: g.newEntry(),
	}
	g.Log = append(g.Log, e)
	return e
}

func (e *announceWinnersEntry) HTML(g *Game) template.HTML {
	names := make([]string, len(g.winners()))
	for i, winner := range g.winners() {
		names[i] = g.NameFor(winner)
	}
	return restful.HTML("Congratulations: %s.", restful.ToSentence(names))
}

func (g *Game) winners() (ps Players) {
	if l := len(g.WinnerIDS); l != 0 {
		ps = make(Players, l)
		for i, pid := range g.WinnerIDS {
			ps[i] = g.PlayerByID(pid)
		}
	}
	return
}
