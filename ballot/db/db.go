/*
 * The MIT License
 *
 * Copyright (c) 2020,  Andrei Taranchenko
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package db

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"github.com/joomcode/errorx"
	"github.com/papito/ballot/ballot/config"
	"github.com/papito/ballot/ballot/model"
	"log"
	"sort"
	"strconv"
	"time"
)

type Store struct {
	Pool          *redis.Pool
	SubConn       redis.PubSubConn
	ServiceSubCon redis.PubSubConn
	redisUrl      string
}

var Const = struct {
	SessionState     string
	SessionUsers     string
	SessionObservers string
	User             string
	VoteCount        string
	Tally            string
}{
	"ballot:session:%s:voting",
	"ballot:session:%s:users",
	"ballot:session:%s:observers",
	"ballot:user:%s",
	"ballot:session:%s:vote_count",
	"ballot:session:%s:tally",
}

func newPool(server string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(server, redis.DialUseTLS(true))
			if err != nil {
				return nil, errorx.EnsureStackTrace(err)
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
		fmt.Printf("Error closing connection: %+v", errorx.EnsureStackTrace(err))
	}
}

func (p *Store) Connect(redisUrl string) {
	p.Pool = newPool(redisUrl)
	p.SubConn = redis.PubSubConn{Conn: p.Pool.Get()}
	p.ServiceSubCon = redis.PubSubConn{Conn: p.Pool.Get()}
}

func (p *Store) ExpireKey(key string) error {
	_, err := p.Pool.Get().Do("EXPIRE", key, config.SessionTtl)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}
	return nil
}

func (p *Store) Set(key string, val interface{}) error {
	_, err := p.Pool.Get().Do("SET", key, val)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}

	err = p.ExpireKey(key)
	if err != nil {
		return err
	}

	return nil
}

func (p *Store) GetSetLength(key string) (int, error) {
	size, err := redis.Int(p.Pool.Get().Do("SCARD", key))
	if err != nil {
		return 0, errorx.EnsureStackTrace(err)
	}

	return size, nil
}

func (p *Store) Del(key string) error {
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("DEL", key)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}
	return nil
}

func (p *Store) Incr(key string, num uint8) error {
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("INCRBY", key, num)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}
	return nil
}

func (p *Store) Decr(key string, num uint8) error {
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("DECRBY", key, num)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}
	return nil
}

func (p *Store) GetInt(key string) (int, error) {
	c := p.Pool.Get()
	defer p.Close(c)
	val, err := redis.Int(c.Do("GET", key))
	if err != nil {
		return 0, errorx.EnsureStackTrace(err)
	}
	return val, nil
}

func (p *Store) GetStr(key string) (string, error) {
	c := p.Pool.Get()
	defer p.Close(c)
	val, err := redis.String(c.Do("GET", key))
	if err != nil {
		return "", errorx.EnsureStackTrace(err)
	}
	return val, nil
}

func (p *Store) SetHashKey(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("HSET", redisArgs[:]...)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}

	err = p.ExpireKey(key)
	if err != nil {
		return err
	}

	return nil
}

func (p *Store) DelHashKey(field string) error {
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("HDEL", field)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}
	return nil
}

func (p *Store) GetHashKey(key string, field string) (string, error) {
	c := p.Pool.Get()
	defer p.Close(c)
	val, err := redis.String(c.Do("HGET", key, field))
	if err != nil {
		log.Println(err)
		return "", err
	}
	return val, nil
}

func (p *Store) GetSessionVoterIds(sessionId string) ([]string, error) {
	c := p.Pool.Get()
	defer p.Close(c)
	key := fmt.Sprintf(Const.SessionUsers, sessionId)
	userIds, err := redis.Strings(c.Do("SMEMBERS", key))
	if err != nil {
		return make([]string, 0), errorx.EnsureStackTrace(err)
	}

	return userIds, nil
}

func (p *Store) GetSessionObserverIds(sessionId string) ([]string, error) {
	c := p.Pool.Get()
	defer p.Close(c)
	key := fmt.Sprintf(Const.SessionObservers, sessionId)
	userIds, err := redis.Strings(c.Do("SMEMBERS", key))
	if err != nil {
		return make([]string, 0), errorx.EnsureStackTrace(err)
	}

	return userIds, nil
}

func (p *Store) AddToSet(key string, args ...interface{}) error {
	// combine the key and the args into a list of interfaces
	redisArgs := []interface{}{key}
	redisArgs = append(redisArgs, args...)
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("SADD", redisArgs[:]...)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}

	err = p.ExpireKey(key)
	if err != nil {
		return err
	}

	return nil
}

func (p *Store) RemoveFromSet(key string, val string) error {
	c := p.Pool.Get()
	defer p.Close(c)
	_, err := c.Do("SREM", key, val)
	if err != nil {
		return errorx.EnsureStackTrace(err)
	}
	return nil
}

func (p *Store) GetSessionVoters(sessionId string) ([]model.User, error) {
	return p.GetUsers(sessionId, false)
}

func (p *Store) GetSessionObservers(sessionId string) ([]model.User, error) {
	return p.GetUsers(sessionId, true)
}

func (p *Store) GetUsers(sessionId string, isObserver bool) ([]model.User, error) {
	var userIds = make([]string, 0)
	var err error

	if isObserver {
		userIds, err = p.GetSessionObserverIds(sessionId)
		if err != nil {
			return make([]model.User, 0), errorx.EnsureStackTrace(err)
		}

	} else {
		userIds, err = p.GetSessionVoterIds(sessionId)
		if err != nil {
			return make([]model.User, 0), errorx.EnsureStackTrace(err)
		}

	}

	log.Printf("Session users for [%s]: %s", sessionId, userIds)

	c := p.Pool.Get()
	defer p.Close(c)

	if len(userIds) == 0 {
		return make([]model.User, 0), nil
	}

	for _, userId := range userIds {
		key := fmt.Sprintf(Const.User, userId)
		_ = c.Send("HGETALL", key)
	}

	res, err := redis.Values(c.Do(""))
	if err != nil {
		return make([]model.User, 0), errorx.EnsureStackTrace(err)
	}

	var users = make([]model.User, 0)

	for _, r := range res {
		switch t := r.(type) {
		case redis.Error:
			return make([]model.User, 0), errorx.Decorate(t, "Redis error")
		case []interface{}:
			m, _ := redis.StringMap(r, nil)

			estimate := m["estimate"]
			isObserver, _ := strconv.Atoi(m["is_observer"])

			user := model.User{
				UserId:     m["id"],
				Name:       m["name"],
				Estimate:   estimate,
				Voted:      estimate != model.NoEstimate,
				Joined:     m["joined"],
				IsObserver: isObserver == 1,
			}
			users = append(users, user)
		default:
			return make([]model.User, 0), errorx.EnsureStackTrace(
				fmt.Errorf("unexpected type: %T", t))
		}
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Joined < users[j].Joined
	})

	return users, nil
}

func (p *Store) GetUser(userId string) (model.User, error) {
	key := fmt.Sprintf(Const.User, userId)
	c := p.Pool.Get()
	defer p.Close(c)

	resp, err := c.Do("HGETALL", key)
	if err != nil {
		return model.User{}, errorx.EnsureStackTrace(err)
	}

	m, _ := redis.StringMap(resp, nil)
	estimate := m["estimate"]
	isObserver, _ := strconv.Atoi(m["is_observer"])
	IsAdmin, _ := strconv.Atoi(m["is_admin"])

	user := model.User{
		UserId:     m["id"],
		Name:       m["name"],
		Estimate:   estimate,
		Voted:      estimate != model.NoEstimate,
		Joined:     m["joined"],
		IsObserver: isObserver == 1,
		IsAdmin:    IsAdmin == 1,
	}

	return user, nil
}
