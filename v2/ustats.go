package main

import (
	"encoding/json"
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

func (ustat *ustats) gamesPlayed() int64 {
	return ustat.stats2P.GamesPlayed + ustat.stats3P.GamesPlayed + ustat.stats4P.GamesPlayed
}

func (ustat *ustats) gamesWon() int64 {
	return ustat.stats2P.Won + ustat.stats3P.Won + ustat.stats4P.Won
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

// User stats at a specific player count (e.g., 2P, 3P, 4P games)
type ustatsPNum struct {
	// Stats at player count
	stats
	// Advanced stats at player count
	advStats
}

// Player stats for a single game
type stats struct {
	// Number of games played at player count
	GamesPlayed int64 `json:"gamesPlayed"`
	// Number of games won at player count
	Won int64 `json:"won"`
	// Number of points scored at player count
	Scored int64 `json:"scored"`
	// Number of moves made by player
	Moves uint64 `json:"moves"`
	// Number of each type of card played by player
	CardsPlayed cardCount `json:"cardsPlayed"`
	// Number of each type of card jewel played as by player
	JewelsAs cardCount `json:"jewelsAs"`
	// Number of each type of card claimed by player
	Claimed cardCount `json:"claimed"`
	// Card type that thief number is place on by player
	Placed [3]cardCount `json:"placed"`
	// Amount of time passed between player moves by player
	Think time.Duration `json:"think"`
	// Position player finished (e.g., 1st, 2nd, etc.)
	Finish uint64 `json:"finish"`
}

type advStats struct {
	// Average number of each type of card played
	PlayedAvg cardCountAverages `json:"playedAvg"`
	// Average number of each type of card jewel played as
	JewelAsAvg cardCountAverages `json:"jewelAsAvg"`
	// Average number of each type of card claimed
	ClaimedAvg cardCountAverages `json:"claimedAvg"`
	// Average number of each type of card theif number placed on
	PlacedAvg [3]cardCountAverages `json:"placedAvg"`
	// Average time for player to make a move
	ThinkAvg float32 `json:"thinkAvg"`
	// Number of first place finishes
	Finish1 uint64 `json:"finish1"`
	// Number of second place finishes
	Finish2 uint64 `json:"finish2"`
	// Number of third place finishes
	Finish3 uint64 `json:"finish3"`
	// Number of fourth place finishes
	Finish4 uint64 `json:"finish4"`
	// Average finishing position
	FinishAvg float32 `json:"finishAvg"`
	// Average Score
	ScoreAvg float32 `json:"scoreAvg"`
	// Win percentage at player count
	WinPercentage float32 `json:"winPercentage"`
}

func (g *Game) updatePNum(stat *ustatsPNum, ukey *datastore.Key) {
	stat.GamesPlayed++
	for _, key := range g.WinnerKeys {
		if key.Equal(ukey) {
			stat.Won++
			break
		}
	}

	if stat.GamesPlayed > 0 {
		stat.WinPercentage = float32(stat.Won) / float32(stat.GamesPlayed)
	}

	p := g.playerByUserKey(ukey)
	if p == nil {
		return
	}

	stat.stats.Moves += p.Stats.Moves

	stat.stats.CardsPlayed.add(p.Stats.CardsPlayed)
	stat.advStats.PlayedAvg = stat.stats.CardsPlayed.avg(stat.GamesPlayed)

	stat.stats.JewelsAs.add(p.Stats.JewelsAs)
	stat.advStats.JewelAsAvg = stat.stats.JewelsAs.avg(stat.GamesPlayed)

	stat.stats.Claimed.add(p.Stats.Claimed)
	stat.advStats.ClaimedAvg = stat.stats.Claimed.avg(stat.GamesPlayed)

	for i := range stat.stats.Placed {
		stat.stats.Placed[i].add(p.Stats.Placed[i])
		stat.advStats.PlacedAvg[i] = stat.stats.Placed[i].avg(stat.GamesPlayed)
	}

	stat.stats.Think += p.Stats.Think

	stat.stats.Scored += p.Stats.Scored

	stat.stats.Finish += p.Stats.Finish
	if stat.GamesPlayed > 0 {
		stat.advStats.ThinkAvg = float32(stat.stats.Think) / float32(stat.GamesPlayed)
		stat.advStats.FinishAvg = float32(stat.stats.Finish) / float32(stat.GamesPlayed)
		stat.advStats.ScoreAvg = float32(stat.stats.Scored) / float32(stat.GamesPlayed)
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
}

func newUStats(ukey *datastore.Key) *ustats {
	return &ustats{Key: newUStatsKey(ukey)}
}

func newUStatsKey(ukey *datastore.Key) *datastore.Key {
	return datastore.NameKey(ustatsKind, "singleton", ukey)
}

func (cl *client) getUStats(c *gin.Context, ukeys ...*datastore.Key) ([]*ustats, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

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
			} else if e != nil {
				return nil, err
			}
		}
	}
	return stats, nil
}

func (g *Game) updateUStats(stats []*ustats) {
	for i := range stats {
		switch g.NumPlayers {
		case 2:
			g.updatePNum(&(stats[i].stats2P), g.UserKeys[i])
		case 3:
			g.updatePNum(&(stats[i].stats3P), g.UserKeys[i])
		case 4:
			g.updatePNum(&(stats[i].stats4P), g.UserKeys[i])
		}
	}
}
