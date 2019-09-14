package db

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

type Store struct {
	RedisConn redis.Conn
	SubConn   redis.PubSubConn
	redisUrl string
}

var Const = struct {
	SessionVoting string
	SessionUsers string
	User string
}{
	"session:%s:voting",
	"session:%s:users",
	"user:%s"}

func (p *Store) Connect(redisUrl string)  {
	var err error
	p.RedisConn, err = redis.DialURL(redisUrl)
	if err != nil {log.Fatal("Error connecting to Redis ", err)}

	c, err := redis.DialURL(redisUrl)
	if err != nil {log.Fatal("Error connecting to Redis ", err)}
	p.SubConn = redis.PubSubConn{Conn: c}
}

func (p *Store) SetKey(key string, val interface{}) error {
	_, err := p.RedisConn.Do("SET", key, val)

	if err != nil {
		log.Println(err)
		return err
	}

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

func (p *Store) AddToSet(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	_, err := p.RedisConn.Do("SADD", redisArgs[:]...)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
