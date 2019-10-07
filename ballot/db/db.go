package db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/papito/ballot/ballot/model"
	"log"
	"time"
)

type Store struct {
	Pool *redis.Pool
	SubConn   redis.PubSubConn
	ServiceSubCon redis.PubSubConn
	redisUrl  string
}

var Const = struct {
	SessionState string
	SessionUsers string
	User         string
	UserCount    string
	VoteCount    string
}{
	"session:%s:voting",
	"session:%s:users",
	"user:%s",
	"session:%s:user_count",
	"session:%s:vote_count",
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle: 3,
		IdleTimeout: 240 * time.Second,
		Dial: func () (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (p *Store) Close(c redis.Conn) {
	err := c.Close()
	if err != nil {
		fmt.Printf("Error closing connection: %v", err)
	}
}

func (p *Store) Connect(redisUrl string)  {
	p.Pool = newPool(redisUrl)
	p.SubConn = redis.PubSubConn{Conn: p.Pool.Get()}
	p.ServiceSubCon = redis.PubSubConn{Conn: p.Pool.Get()}
}

func (p *Store) Set(key string, val interface{}) error {
	_, err := p.Pool.Get().Do("SET", key, val)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) Del(key string) error {
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("DEL", key)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) Incr(key string, num uint8) error {
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("INCRBY", key, num)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) Decr(key string, num uint8) error {
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("DECRBY", key, num)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) GetInt(key string) (int, error) {
	c  := p.Pool.Get()
	defer p.Close(c)
	val, err := redis.Int(c.Do("GET", key))
	if err != nil {log.Println(err); return 0, err}
	return val, nil
}

func (p *Store) SetHashKey(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("HSET", redisArgs[:]...)

	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) DelHashKey(key string, field string) error {
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("HDEL", field)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) GetHashKey(key string, field string) (string, error) {
	c  := p.Pool.Get()
	defer p.Close(c)
	val, err := redis.String(c.Do("HGET", key, field))
	if err != nil {log.Println(err); return "", err}
	return val, nil
}

func (p* Store) GetSessionUserIds(sessionId string) ([]string, error) {
	c  := p.Pool.Get()
	defer p.Close(c)
	key := fmt.Sprintf(Const.SessionUsers, sessionId)
	userIds, err := redis.Strings(c.Do("SMEMBERS", key))
	if err != nil {
		return make([]string, 0), fmt.Errorf("ERROR %v", err)
	}

	return userIds, nil
}

func (p *Store) AddToSet(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err :=c.Do("SADD", redisArgs[:]...)
	if err != nil {log.Println(err); return err}
	return nil
}

func (p *Store) RemoveFromSet(key string, val string) error {
	c  := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("SREM", key, val)
	if err != nil {log.Println(err); return err}
	return nil
}

func (p *Store) GetSessionUsers(sessionId string) ([]model.User, error) {
	userIds, err := p.GetSessionUserIds(sessionId)
	if err != nil {
		return make([]model.User, 0), fmt.Errorf("ERROR %v", err)
	}
	log.Printf("Session voters for [%s]: %s", sessionId, userIds)

	c  := p.Pool.Get()
	defer p.Close(c)

	for _, userId := range userIds {
		key := fmt.Sprintf("user:%s", userId)
		_ = c.Send("HGETALL", key)
	}

	res, err := redis.Values(c.Do(""))
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	var users []model.User

	for i, r := range res {
		switch t := r.(type) {
		case redis.Error:
			return make([]model.User, 0), fmt.Errorf("res[%d] is redis.Error %v\n", i, r)
		case []interface{}:
			m, _ := redis.StringMap(r, nil)

			estimate := m["estimate"]

			user := model.User{
				UserId:   m["id"],
				Name:     m["name"],
				Estimate: estimate,
				Voted:    estimate != model.NoEstimate,
			}
			users = append(users, user)
		default:
			return make([]model.User, 0), fmt.Errorf("UNEXPECTED TYPE: %T", t)
		}
	}

	return users, nil
}

func (p *Store) GetUser(userId string) (model.User, error) {
	key := fmt.Sprintf("user:%s", userId)
	c  := p.Pool.Get()
	defer p.Close(c)

	resp, err := c.Do("HGETALL", key)
	if err != nil {log.Println(err); return model.User{}, err}

	m, _ := redis.StringMap(resp, nil)
	estimate := m["estimate"]
	user := model.User{
		UserId:   m["id"],
		Name:     m["name"],
		Estimate: estimate,
		Voted:    estimate != model.NoEstimate,
	}

	return user, nil
}