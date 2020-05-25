package main

import (
	"fmt"

	"github.com/SlothNinja/log"
	"github.com/SlothNinja/send"
	"github.com/gin-gonic/gin"
	"github.com/mailjet/mailjet-apiv3-go"
)

func (g *Game) SendTurnNotificationsTo(c *gin.Context, ps ...*Player) error {
	log.Debugf(msgEnter)
	defer log.Debugf(msgExit)

	subject := fmt.Sprintf("It's your turn in %s (%s #%d).", g.Type, g.Title, g.ID())
	url := fmt.Sprintf(`<a href="http://www.slothninja.com/%s/game/show/%d">here</a>`, g.Type.Prefix(), g.ID())
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
		u := p.User
		if u.EmailNotifications {
			m := msgInfo
			m.To = &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: g.EmailFor(u.ID()),
					Name:  u.Name,
				},
			}
			msgInfos = append(msgInfos, m)
		}
	}
	_, err := send.Messages(c, msgInfos...)
	return err
}
