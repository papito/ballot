# Ballot ![Publish Docker](https://github.com/papito/ballot/workflows/Publish%20Docker/badge.svg?branch=master)

A web-based replacement for physical Scrum estimation cards, most useful for distributed teams. 

![Ballot](img/snapshot.png)

It's live - [try it here](https://ballot.renegadeotter.com/#/)

## Features

- A vote will end automatically when all votes are in
- A vote can be finalized even if someone doesn't vote, because they are raiding the company fridge
- Observers can join a vote

## Installing and running

To just get it working out of the box:

    docker pull papito/ballot:latest
    docker run -td -p8080:8080 --name ballot papito/ballot:latest

With more options:

    docker pull papito/ballot:latest
    docker run -td -p8080:8080 --name ballot -e"REDIS_URL=[redis host]" -e"HTTP_HOST=http://your optional ballot host"  papito/ballot:latest


## Integrations

Ballot is not meant to be integrated with systems like Slack or JIRA. It is meant to be a frictionless voting tool,
and any integration with 3rd-party systems is:

  * More unnecessary work
  * Extra clicks for a task that should be very simple

Users in a voting session already *know* what story they are discussing. The voting cards don't need to know what it is.
The group uses Ballot to take a vote, enter the result into whatever system they leverage, and that's it.

## Development setup

### Prerequisites
  * Computer technology
  * Docker Compose
  * Node
  * Go 1.13

### Build for development

To invoke NPM and transpile Typescript:
```bash
make build
```

To run the app on port `8080`:

```bash
make run
```

Dev and test Redis containers will be brought up when running `make run` or `make up`. Run `make down` to stop the containers. You do NOT need to have local Redis running - the `run` command will bring up a Redis container.

### IntelliJ IDEA

Checked in at the top level is `watchers.xml`. The config can be imported into your IDEA file watcher settings to detect Typescript file changes and automatically transpile to Javascript.


### Build with Docker

`docker build .`

Note that this will install local Redis in the container, but that instance can be ignored if you configure Redis with environment variables (see below).

### Environment variables

  * HTTP_PORT - dictates which port the application will run on.
  * HTTP_HOST - used to correctly display the session URL (does not affect the behavior).
  * REDIS_URL - Redis URL. Otherwise will connect to Docker Redis on the default port.
  * ENV - context environment. `test`, `development`, or `production`. You can ignore this.


### Connecting to Redis on Docker host

By default, the Docker container will have its own Redis instance, but you can have a persistent Redis running on Docker
host, by using the `--network="host"` flag of Docker `run` command.


### Running tests

    make test

## Redis schema

#### user:{user_id} -> Hash

User state for a session is stored here, and yes, this assumes that a user can only vote in one session.

| Field       | Type                  |
|-------------|-----------------------|
| id          | UUID                  |
| name        | String                |
| estimate    | String                |
| joined      | String (datetime)     |
| is_observer | Flag                  |
| is_admin    | Flag                  |

`estimate` is an empty string by default.

`joined` is used to sort users in a session by the order in which they had joined,
to make the order deterministic.

#### ballot:session:{session_id}:users -> Set[String]

A set of users in this current session.

#### ballot:session:{session_id}:observers -> Set[String]

A set of observers in this current session.

#### ballot:session:{session_id}:vote_count -> Int

Number of users in a session who cast a vote.

#### ballot:session:{session_id}:tally -> String

Final vote tally.

#### ballot:session:{session_id}:voting -> Int

  * 0 - Not voting (idle before start, or vote finished)
  * 1 - Voting
