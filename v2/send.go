package main

import (
	"context"
	"fmt"

	"github.com/SlothNinja/send"
	"github.com/mailjet/mailjet-apiv3-go"
)

func (cl *client) sendTurnNotificationsTo(g *Game, ps ...*player) error {
	subject := fmt.Sprintf("It's your turn in %s (%s #%d).", g.Type, g.Title, g.id())
	url := fmt.Sprintf(`<a href="https://got.slothninja.com/#/game/%d">here</a>`, g.id())
	body := fmt.Sprintf(`<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN" "http://www.w3.org/TR/html4/loose.dtd">
		<html>
			<head>
				<meta http-equiv="content-type" content="text/html; charset=ISO-8859-1">
			</head>
			<body bgcolor="#ffffff" text="#000000">
				<p>%s</p>
				<p>You can take your turn %s.</p>
			</body>
		</html>`, subject, url)

	msgInfo := mailjet.InfoMessagesV31{
		From: &mailjet.RecipientV31{
			Email: "webmaster@slothninja.com",
			Name:  "Webmaster",
		},
		Subject:  subject,
		HTMLPart: body,
	}

	msgInfos := []mailjet.InfoMessagesV31{}

	for _, p := range ps {
		if g.emailNotificationsFor(p) {
			m := msgInfo
			m.To = &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: g.emailFor(p),
					Name:  g.nameFor(p),
				},
			}
			msgInfos = append(msgInfos, m)
		}
	}
	_, err := send.Messages(context.Background(), msgInfos...)
	if err != nil {
		cl.Log.Warningf(err.Error())
	}
	return err
}

// func (cl *client) userByPlayer(g *Game, p *player) (*user.User, error) {
// 	cl.Log.Debugf(msgEnter)
// 	defer cl.Log.Debugf(msgExit)
//
// 	if p == nil {
// 		return nil, user.ErrNotFound
// 	}
//
// 	if g == nil {
// 		return nil, user.ErrNotFound
// 	}
//
// 	index := p.ID - 1
// 	if index >= 0 && index < len(cl.g.UserIDS) {
// 		return cl.User.Get(cl.ctx, cl.g.UserIDS[index])
// 	}
// 	return nil, user.ErrNotFound
// }
