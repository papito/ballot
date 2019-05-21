package db

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

// FIXME: env  var
const redisUrl = "redis://localhost:6379"


type Store struct {
	redisConn redis.Conn
	redisUrl string
}

func (p *Store) Connect()  {
	var err error
	p.redisConn, err = redis.DialURL(redisUrl)
	if err != nil {
		log.Fatal("Error connecting to Redis ", err)
	}
}

func (p *Store) SetKey(key string, val interface{}) error {
	_, err := p.redisConn.Do("SET", key, val)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (p *Store) GetInt(key string) (int, error) {
	val, err := redis.Int(p.redisConn.Do("GET", key))
	println(val)

	if err != nil {
		log.Println(err)
		return 0, err
	}

	return val, nil
}

func (p *Store) SetHashKey(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	_, err := p.redisConn.Do("HSET", redisArgs[:]...)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (p *Store) AddToSet(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	_, err := p.redisConn.Do("SADD", redisArgs[:]...)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
