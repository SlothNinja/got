package main

import (
	"encoding/json"
	"strings"

	"github.com/SlothNinja/sn/v2"
)

type cKind int

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
	lampCard:       "Lamp",
	camelCard:      "Camel",
	swordCard:      "Sword",
	carpetCard:     "Carpet",
	coinsCard:      "Coins",
	turbanCard: "Turban",
	jewels:     "Jewels",
	guardCard:      "Guard",
	sCamelCard:     "Camel",
	sLampCard:      "Lamp",
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
	lampCard:       1,
	camelCard:      4,
	swordCard:      5,
	carpetCard:     3,
	coinsCard:      3,
	turbanCard: 2,
	jewels:     2,
	guardCard:      -1,
	sCamelCard:     0,
	sLampCard:      0,
}

func (t cKind) String() string {
	return ctypeStrings[t]
}

func (t cKind) LString() string {
	return strings.ToLower(t.String())
}

func (t cKind) IDString() string {
	switch t {
	case sCamelCard:
		return "start-camel"
	case sLampCard:
		return "start-lamp"
	default:
		return t.LString()
	}
}

func (t cKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.IDString())
}

func (t *cKind) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	*t = toCType(s)
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

// IDString outputs a card id.
func (c Card) IDString() string {
	return c.Kind.IDString()
}

func newStartHand() Cards {
	return Cards{newCard(sLampCard, true), newCard(sLampCard, true), newCard(sCamelCard, true)}
}

var toolTipStrings = map[cKind]string{
	noType:     "None",
	lampCard:       "Move in a straight line until coming to the edge of the grid, an empty space, or another Thief.",
	camelCard:      "Move exactly 3 spaces in any direction. The spaces do not have to be in a straight line, but you cannot move over the same space twice.",
	swordCard:      "Move in a straight line until you come to another player's thief. Bump that thief to the next card and place your thief on the vacated card.",
	carpetCard:     "Move in a straight line over at least one empty space.  Stop moving your thief on the first card after the empty space(s).",
	coinsCard:      "Move one space and then draw an additional card during the draw step. Your hand size is permanently increased by 1.",
	turbanCard: "Move two spaces. Claim the first Magic Item you pass over in addition to the card you claim in the Claim Magic Item step.",
	jewels:     "Move as if you played the card that was last played by an opponent.",
	guardCard:      "This card cannot be played and does nothing for you in your hand.",
	sCamelCard:     "Move exactly 3 spaces in any direction. The spaces do not have to be in a straight line, but you cannot move over the same space twice.",
	sLampCard:      "Move in a straight line until coming to the edge of the grid, an empty space, or another Thief.",
}

// ToolTip outputs a description of the cards ability.
func (c Card) ToolTip() string {
	return c.Kind.toolTip()
}

func (t cKind) toolTip() string {
	return toolTipStrings[t]
}

// Value provides the point value of a card.
func (c *Card) Value() int {
	return ctypeValues[c.Kind]
}
