package db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/papito/ballot/ballot/model"
	"log"
)

type Store struct {
	RedisConn redis.Conn
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

func (p *Store) Connect(redisUrl string)  {
	var err error
	p.RedisConn, err = redis.DialURL(redisUrl)
	if err != nil {log.Fatalf("Error connecting to Redis: %s", err)}

	c, err := redis.DialURL(redisUrl)
	if err != nil {log.Fatalf("Error connecting to Redis: %s ", err)}
	p.SubConn = redis.PubSubConn{Conn: c}

	c2, err := redis.DialURL(redisUrl)
	if err != nil {log.Fatalf("Error connecting to Redis: %s ", err)}
	p.ServiceSubCon = redis.PubSubConn{Conn: c2}
}

func (p *Store) Set(key string, val interface{}) error {
	_, err := p.RedisConn.Do("SET", key, val)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) Del(key string) error {
	_, err := p.RedisConn.Do("DEL", key)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) Incr(key string, num uint8) error {
	_, err := p.RedisConn.Do("INCRBY", key, num)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) Decr(key string, num uint8) error {
	_, err := p.RedisConn.Do("DECRBY", key, num)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) GetInt(key string) (int, error) {
	val, err := redis.Int(p.RedisConn.Do("GET", key))
	if err != nil {log.Println(err); return 0, err}
	return val, nil
}

func (p *Store) SetHashKey(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	_, err := p.RedisConn.Do("HSET", redisArgs[:]...)

	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) DelHashKey(key string, field string) error {
	_, err := p.RedisConn.Do("HDEL", field)
	if err != nil {log.Println(err); return err}

	return nil
}

func (p *Store) GetHashKey(key string, field string) (string, error) {
	val, err := redis.String(p.RedisConn.Do("HGET", key, field))
	if err != nil {log.Println(err); return "", err}
	return val, nil
}

func (p* Store) GetSessionUserIds(sessionId string) ([]string, error) {
	key := fmt.Sprintf(Const.SessionUsers, sessionId)
	userIds, err := redis.Strings(p.RedisConn.Do("SMEMBERS", key))
	if err != nil {
		return make([]string, 0), fmt.Errorf("ERROR %v", err)
	}

	return userIds, nil
}

func (p *Store) AddToSet(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	_, err := p.RedisConn.Do("SADD", redisArgs[:]...)
	if err != nil {log.Println(err); return err}
	return nil
}

func (p *Store) RemoveFromSet(key string, val string) error {
	_, err := p.RedisConn.Do("SREM", key, val)
	if err != nil {log.Println(err); return err}
	return nil
}

func (p *Store) GetSessionUsers(sessionId string) ([]model.User, error) {
	userIds, err := p.GetSessionUserIds(sessionId)
	if err != nil {
		return make([]model.User, 0), fmt.Errorf("ERROR %v", err)
	}
	log.Printf("Session voters for [%s]: %s", sessionId, userIds)

	for _, userId := range userIds {
		key := fmt.Sprintf("user:%s", userId)
		_ = p.RedisConn.Send("HGETALL", key)
	}

	res, err := redis.Values(p.RedisConn.Do(""))
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
	resp, err := p.RedisConn.Do("HGETALL", key)
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