package main

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/gin-gonic/gin"
)

type ustats struct {
	Key       *datastore.Key `json:"key" datastore:"__key__"`
	Encoded   string         `json:"-" datastore:",noindex"`
	stats2P   ustatsPNum     `json:"stats2P"`
	stats3P   ustatsPNum     `json:"stats3P"`
	stats4P   ustatsPNum     `json:"stats4P"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type jstats struct {
	Stats2P ustatsPNum `json:"stats2P"`
	Stats3P ustatsPNum `json:"stats3P"`
	Stats4P ustatsPNum `json:"stats4P"`
}

func (ustat *ustats) Load(ps []datastore.Property) error {
	err := datastore.LoadStruct(ustat, ps)
	if err != nil {
		return err
	}

	var stats jstats
	err = json.Unmarshal([]byte(ustat.Encoded), &stats)
	if err != nil {
		return err
	}

	ustat.stats2P = stats.Stats2P
	ustat.stats3P = stats.Stats3P
	ustat.stats4P = stats.Stats4P
	return nil
}

func (ustat *ustats) Save() ([]datastore.Property, error) {
	stats := jstats{
		Stats2P: ustat.stats2P,
		Stats3P: ustat.stats3P,
		Stats4P: ustat.stats4P,
	}

	encoded, err := json.Marshal(stats)
	if err != nil {
		return nil, err
	}
	ustat.Encoded = string(encoded)

	t := time.Now()
	if ustat.CreatedAt.IsZero() {
		ustat.CreatedAt = t
	}

	ustat.UpdatedAt = t
	return datastore.SaveStruct(ustat)
}

func (ustat *ustats) LoadKey(k *datastore.Key) error {
	ustat.Key = k
	return nil
}

type ustatsPNum struct {
	Played        int64   `json:"played"`
	Won           int64   `json:"won"`
	WinPercentage float32 `json:"winPercentage"`
	stats
	advStats
}

type stats struct {
	Moves    uint64        `json:"moves"`
	Played   cardCount     `json:"played"`
	JewelsAs cardCount     `json:"jewelsAs"`
	Claimed  cardCount     `json:"claimed"`
	Placed   [3]cardCount  `json:"placed"`
	Think    time.Duration `json:"think"`
	Finish   uint64        `json:"finish"`
}

type advStats struct {
	PlayedAvg  cardCountAverages    `json:"playedAvg"`
	JewelAsAvg cardCountAverages    `json:"jewelAsAvg"`
	ClaimedAvg cardCountAverages    `json:"claimedAvg"`
	PlacedAvg  [3]cardCountAverages `json:"placedAvg"`
	ThinkAvg   float32              `json:"thinkAvg"`
	Finish1    uint64               `json:"finish1"`
	Finish2    uint64               `json:"finish2"`
	Finish3    uint64               `json:"finish3"`
	Finish4    uint64               `json:"finish4"`
	FinishAvg  float32              `json:"finishAvg"`
}

func (stat ustatsPNum) update(g *game, ukey *datastore.Key) ustatsPNum {
	stat.Played++
	for _, key := range g.WinnerKeys {
		if key.Equal(ukey) {
			stat.Won++
			break
		}
	}

	if stat.Played > 0 {
		stat.WinPercentage = float32(stat.Won) / float32(stat.Played)
	}

	p := g.playerByUserKey(ukey)
	if p == nil {
		return stat
	}

	stat.stats.Moves += p.Stats.Moves

	stat.stats.Played.add(p.Stats.Played)
	stat.advStats.PlayedAvg = stat.stats.Played.avg(stat.Played)

	stat.stats.JewelsAs.add(p.Stats.JewelsAs)
	stat.advStats.JewelAsAvg = stat.stats.JewelsAs.avg(stat.Played)

	stat.stats.Claimed.add(p.Stats.Claimed)
	stat.advStats.ClaimedAvg = stat.stats.Claimed.avg(stat.Played)

	for i := range stat.stats.Placed {
		stat.stats.Placed[i].add(p.Stats.Placed[i])
		stat.advStats.PlacedAvg[i] = stat.stats.Placed[i].avg(stat.Played)
	}

	stat.stats.Think += p.Stats.Think

	stat.stats.Finish += p.Stats.Finish
	if stat.Played > 0 {
		stat.advStats.ThinkAvg = float32(stat.stats.Think) / float32(stat.Played)
		stat.advStats.FinishAvg = float32(stat.stats.Finish) / float32(stat.Played)
	}

	switch p.Stats.Finish {
	case 1:
		stat.advStats.Finish1++
	case 2:
		stat.advStats.Finish2++
	case 3:
		stat.advStats.Finish3++
	case 4:
		stat.advStats.Finish4++
	}
	return stat
}

func newUStats(ukey *datastore.Key) *ustats {
	return &ustats{Key: newUStatsKey(ukey)}
}

func newUStatsKey(ukey *datastore.Key) *datastore.Key {
	return datastore.NameKey(ustatsKind, "singleton", ukey)
}

func (cl client) getUStats(c *gin.Context, ukeys ...*datastore.Key) ([]*ustats, error) {
	l := len(ukeys)
	stats := make([]*ustats, l)
	ks := make([]*datastore.Key, l)
	for i, ukey := range ukeys {
		stats[i] = newUStats(ukey)
		ks[i] = stats[i].Key
	}

	err := cl.DS.GetMulti(c, ks, stats)
	if err == nil {
		return stats, nil
	}

	if merr, ok := err.(datastore.MultiError); ok {
		for i, e := range merr {
			if e == datastore.ErrNoSuchEntity {
				stats[i] = newUStats(ukeys[i])
			} else if err != nil {
				return nil, err
			}
		}
	}
	return stats, nil
}

func (cl client) updateUStats(c *gin.Context, g *game) ([]*ustats, error) {
	stats, err := cl.getUStats(c, g.UserKeys...)
	if err != nil {
		return nil, err
	}

	for i := range stats {
		switch g.NumPlayers {
		case 2:
			stats[i].stats2P.update(g, g.UserKeys[i])
		case 3:
			stats[i].stats3P.update(g, g.UserKeys[i])
		case 4:
			stats[i].stats4P.update(g, g.UserKeys[i])
		default:
			return nil, fmt.Errorf("invalid number of players")
		}
	}
	return stats, nil
}
