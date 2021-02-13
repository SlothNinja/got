package main

import "encoding/json"

type color int

func defaultColors() []color {
	return []color{yellow, purple, green, black}
}

const (
	none color = iota
	yellow
	purple
	green
	black
)

var colorString = [5]string{"none", "yellow", "purple", "green", "black"}

func toColor(s string) color {
	toColor := map[string]color{"none": none, "yellow": yellow, "purple": purple, "green": green, "black": black}
	c, ok := toColor[s]
	if !ok {
		return none
	}
	return c
}

func (c color) String() string {
	return colorString[c]
}

func (c color) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c *color) UnmarshalJSON(bs []byte) error {
	var s string
	err := json.Unmarshal(bs, &s)
	if err != nil {
		return err
	}
	*c = toColor(s)
	return nil
}
