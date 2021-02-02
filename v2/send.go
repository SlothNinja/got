package main

import (
	"fmt"

	"github.com/SlothNinja/send"
	"github.com/SlothNinja/user"
	"github.com/mailjet/mailjet-apiv3-go"
)

func (cl *client) sendTurnNotificationsTo(ps ...*player) error {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	subject := fmt.Sprintf("It's your turn in %s (%s #%d).", cl.g.Type, cl.g.Title, cl.g.id())
	url := fmt.Sprintf(`<a href="http://www.slothninja.com/%s/game/show/%d">here</a>`, cl.g.Type.Prefix(), cl.g.id())
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
		u, err := cl.userByPlayer(p)
		if err != nil {
			cl.Log.Warningf("unable to find user for player %#v: %w", p, err)
			continue
		}

		if u.EmailNotifications {
			m := msgInfo
			m.To = &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: u.Email,
					Name:  u.Name,
				},
			}
			msgInfos = append(msgInfos, m)
		}
	}
	_, err := send.Messages(cl.ctx, msgInfos...)
	return err
}

func (cl *client) userByPlayer(p *player) (*user.User, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	if p == nil {
		cl.Log.Warningf("player is nil")
		return nil, user.ErrNotFound
	}

	if cl.g == nil {
		cl.Log.Warningf("cl.g is nil")
		return nil, user.ErrNotFound
	}

	index := p.ID - 1
	if index >= 0 && index < len(cl.g.UserIDS) {
		return cl.User.Get(cl.ctx, cl.g.UserIDS[index])
	}
	return nil, user.ErrNotFound
}
