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

package response

import "github.com/papito/ballot/ballot/model"

type HealthResponse struct {
	Status string `json:"status"`
}

type WsVoteStarted struct {
	Event string `json:"event"`
}

type WsVoteFinished struct {
	Users []model.User `json:"users"`
	Tally string `json:"tally"`
	Event string `json:"event"`
}

type WsNewUser struct {
	model.User
	Event  string `json:"event"`
}

type WsUserVote struct {
	Event    string `json:"event"`
	UserId   string `json:"user_id"`
}

type WsSession struct {
	Event        string `json:"event"`
	SessionState int    `json:"session_state"`
	Users        []model.User  `json:"users"`
	Observers    []model.User  `json:"observers"`
	Tally        string        `json:"tally"`
}

type WsUserLeftEvent struct {
	Event    string `json:"event"`
	SessionId   string `json:"session_id"`
	UserId   string `json:"user_id"`
}

type WsObserverLeftEvent struct {
	Event    string `json:"event"`
	SessionId   string `json:"session_id"`
	UserId   string `json:"user_id"`
}

const (
	UserAddedEvent = "USER_ADDED"
	ObserverAddedEvent = "OBSERVER_ADDED"
	UserVotedEVent = "USER_VOTED"
	VoteStartedEVent = "VOTING"
	VoteFinishedEvent = "VOTE_FINISHED"
)

