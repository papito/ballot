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

package model

type Session struct {
	SessionId string `json:"id"`
}

type User struct {
	UserId     string `json:"id"`
	Name       string `json:"name"`
	Estimate   string `json:"estimate"`
	Voted      bool   `json:"voted"`
	Joined     string `json:"joined"`
	IsObserver bool   `json:"is_observer"`
	IsAdmin    bool   `json:"is_admin"`
}

type PendingVote struct {
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
	// Pending vote has no estimate since we are hiding it while the vote is going on
}

const (
	NotVoting = iota
	Voting
)

const NoEstimate = ""
