package main

import (
	"time"

	"cloud.google.com/go/datastore"
	"firebase.google.com/go/messaging"
	"github.com/gin-gonic/gin"
)

const subscriptionKind = "Subscription"

type Subscription struct {
	Key       *datastore.Key `datastore:"__key__"`
	Tokens    []string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Subscription) Load(ps []datastore.Property) error {
	var ps2 []datastore.Property
	for _, p := range ps {
		if p.Name != "Key" {
			ps2 = append(ps2, p)
		}
	}
	return datastore.LoadStruct(s, ps2)
}

func (s *Subscription) Save() ([]datastore.Property, error) {
	t := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = t
	}

	s.UpdatedAt = t
	return datastore.SaveStruct(s)
}

func (s *Subscription) LoadKey(k *datastore.Key) error {
	s.Key = k
	return nil
}

func (s *Subscription) Subscribe(token string) bool {
	if token == "" {
		return false
	}

	_, found := s.find(token)
	if found {
		return false
	}

	s.Tokens = append(s.Tokens, token)
	return true
}

func (s *Subscription) Unsubscribe(token string) bool {
	if token == "" {
		return false
	}

	i, found := s.find(token)
	if !found {
		return false
	}

	s.Tokens = append(s.Tokens[:i], s.Tokens[i+1:]...)
	return true
}

func (s *Subscription) find(token string) (int, bool) {
	if token == "" {
		return -1, false
	}
	for i, t := range s.Tokens {
		if t == token {
			return i, true
		}
	}
	return -1, false
}

func (s *Subscription) other(token string) []string {
	if token == "" {
		return s.Tokens
	}
	i, found := s.find(token)
	if !found {
		return s.Tokens
	}

	return append(s.Tokens[:i], s.Tokens[i+1:]...)
}

func (cl *client) getSubcription(c *gin.Context) (*Subscription, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	id, err := getID(c)
	if err != nil {
		return nil, err
	}

	return cl.getCachedSubscription(c, id)
}

func (cl *client) getCachedSubscription(c *gin.Context, id int64) (*Subscription, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	s, err := cl.mcSubscription(id)
	if err == nil {
		return s, nil
	}

	return cl.dsSubscription(c, id)
}

func (cl *client) mcSubscription(id int64) (*Subscription, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	k := newSubscriptionKey(id).Encode()
	item, found := cl.Cache.Get(k)
	if !found {
		return nil, ErrNotFound
	}

	s, ok := item.(*Subscription)
	if !ok {
		cl.Cache.Delete(k)
		return nil, ErrInvalidCache
	}
	return s, nil
}

func (cl *client) dsSubscription(c *gin.Context, id int64) (*Subscription, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	s := newSubscription(id)
	err := cl.DS.Get(c, s.Key, s)
	if err != nil && err != datastore.ErrNoSuchEntity {
		return nil, err
	}
	cl.Cache.SetDefault(s.Key.Encode(), s)
	return s, nil
}

func (cl *client) putSubscription(c *gin.Context, s *Subscription) (*datastore.Key, error) {
	cl.Log.Debugf(msgEnter)
	defer cl.Log.Debugf(msgExit)

	k, err := cl.DS.Put(c, s.Key, s)
	if err != nil {
		return nil, err
	}
	cl.Cache.Delete(k.Encode())
	return k, nil
}

func newSubscription(id int64) *Subscription {
	return &Subscription{Key: newSubscriptionKey(id)}
}

func newSubscriptionKey(id int64) *datastore.Key {
	return datastore.IDKey(subscriptionKind, id, rootKey(id))
}

func (cl *client) getToken(c *gin.Context) (string, error) {

	obj := struct {
		Token string `json:"token"`
	}{}

	err := c.ShouldBind(&obj)
	return obj.Token, err
}

func (cl *client) sendRefreshMessages(c *gin.Context) error {
	cl.Log.Debugf("entering sendRefreshMessages")
	defer cl.Log.Debugf("exiting sendRefreshMessages")

	s, err := cl.getSubcription(c)
	if err != nil {
		return err
	}
	cl.Log.Debugf("subscription: %#v", s)

	token, err := cl.getToken(c)
	if err != nil {
		return err
	}
	cl.Log.Debugf("token: %#v", token)

	tokens := s.other(token)
	cl.Log.Debugf("tokens: %#v", tokens)
	if len(tokens) > 0 {
		resp, err := cl.Messaging.SendMulticast(c, &messaging.MulticastMessage{
			Tokens: tokens,
			Data:   map[string]string{"action": "refresh"},
		})
		if resp != nil {
			cl.Log.Debugf("batch response: %+v", resp)
			for _, r := range resp.Responses {
				cl.Log.Debugf("response: %+v", r)
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}
