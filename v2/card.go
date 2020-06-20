package main

import (
	"encoding/json"
	"strings"

	"github.com/SlothNinja/sn/v2"
)

type cKind uint8

const (
	noType cKind = iota
	lampCard
	camelCard
	swordCard
	carpetCard
	coinsCard
	turbanCard
	jewels
	guardCard
	sCamelCard
	sLampCard
)

type cardCount map[string]uint32
type cardCountAverages map[string]float32

func (cc *cardCount) inc(k cKind) {
	m := *cc
	if m == nil {
		m = make(cardCount)
	}
	m[k.String()]++
	*cc = m
}

func (cc *cardCount) add(cc2 cardCount) {
	m := *cc
	if m == nil {
		m = make(cardCount)
	}
	for k, v := range cc2 {
		m[k] += v
	}
	*cc = m
}

func (cc *cardCount) avg(played int64) cardCountAverages {
	if cc == nil || played == 0 {
		return make(cardCountAverages)
	}

	cca := make(cardCountAverages)
	for k, v := range *cc {
		cca[k] = float32(v) / float32(played)
	}
	return cca
}

func cardTypes() []cKind {
	return []cKind{
		lampCard,
		camelCard,
		swordCard,
		carpetCard,
		coinsCard,
		turbanCard,
		jewels,
		guardCard,
		sCamelCard,
		sLampCard,
	}
}

var ctypeStrings = map[cKind]string{
	noType:     "None",
	lampCard:   "Lamp",
	camelCard:  "Camel",
	swordCard:  "Sword",
	carpetCard: "Carpet",
	coinsCard:  "Coins",
	turbanCard: "Turban",
	jewels:     "Jewels",
	guardCard:  "Guard",
	sCamelCard: "Camel",
	sLampCard:  "Lamp",
}

var stringsCType = map[string]cKind{
	"none":        noType,
	"lamp":        lampCard,
	"camel":       camelCard,
	"sword":       swordCard,
	"carpet":      carpetCard,
	"coins":       coinsCard,
	"turban":      turbanCard,
	"jewels":      jewels,
	"guard":       guardCard,
	"start-camel": sCamelCard,
	"start-lamp":  sLampCard,
}

func toCType(s string) (t cKind) {
	s = strings.ToLower(s)

	var ok bool
	if t, ok = stringsCType[s]; !ok {
		t = noType
	}
	return
}

var ctypeValues = map[cKind]int{
	noType:     0,
	lampCard:   1,
	camelCard:  4,
	swordCard:  5,
	carpetCard: 3,
	coinsCard:  3,
	turbanCard: 2,
	jewels:     2,
	guardCard:  -1,
	sCamelCard: 0,
	sLampCard:  0,
}

func (ck cKind) String() string {
	return ctypeStrings[ck]
}

func (ck cKind) lString() string {
	return strings.ToLower(ck.String())
}

func (ck cKind) idString() string {
	switch ck {
	case sCamelCard:
		return "start-camel"
	case sLampCard:
		return "start-lamp"
	default:
		return ck.lString()
	}
}

func (ck cKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(ck.idString())
}

func (ck *cKind) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	*ck = toCType(s)
	return nil
}

// Card is a playing card used to form grid, player's hand, and player's deck.
type Card struct {
	Kind   cKind `json:"kind"`
	FaceUp bool  `json:"faceUp"`
}

func newCard(t cKind, f bool) *Card {
	return &Card{
		Kind:   t,
		FaceUp: f,
	}
}

// value provides the point value of a card.
func (c *Card) value() int {
	return ctypeValues[c.Kind]
}

// Cards is a slice of cards used to form player's hand or deck.
type Cards []*Card

func newDeck() Cards {
	deck := make(Cards, 64)
	for j := 0; j < 8; j++ {
		for i, typ := range []cKind{lampCard, camelCard, swordCard, carpetCard, coinsCard, turbanCard, jewels, guardCard} {
			deck[i+j*8] = &Card{Kind: typ}
		}
	}
	return deck
}

func (cs Cards) removeAt(i int) Cards {
	return append(cs[:i], cs[i+1:]...)
}

func (cs *Cards) playCardAt(i int) *Card {
	card := (*cs)[i]
	*cs = cs.removeAt(i)
	return card
}

func (cs *Cards) indexFor(c *Card) (int, bool) {
	if cs == nil {
		return -1, false
	}

	for i := range *cs {
		if (*cs)[i].Kind == c.Kind {
			return i, true
		}
	}
	return -1, false
}

func (cs *Cards) play(c *Card) {
	i, found := cs.indexFor(c)
	if found {
		*cs = cs.removeAt(i)
	}
}

func (cs *Cards) draw() *Card {
	var card *Card
	*cs, card = cs.drawS()
	return card
}

func (cs Cards) drawS() (Cards, *Card) {
	i := sn.MyRand.Intn(len(cs))
	card := cs[i]
	cards := cs.removeAt(i)
	return cards, card
}

func (cs *Cards) append(cards ...*Card) {
	*cs = cs.appendS(cards...)
}

func (cs Cards) appendS(cards ...*Card) Cards {
	if len(cards) == 0 {
		return cs
	}
	return append(cs, cards...)
}

func newStartHand() Cards {
	return Cards{newCard(sLampCard, true), newCard(sLampCard, true), newCard(sCamelCard, true)}
}
