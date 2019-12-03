# Ballot

A web-based replacement for physical Scrum estimation cards, most useful for distributed teams. 

## Integrations

Ballot is not meant to be integrated with systems like Slack or JIRA. It is meant to be a frictionless voting tool, 
and any integration with 3rd party systems is:

  * More unnecessary work
  * Extra clicks for a task that should be very simple

Users in a voting session already *know* what story they are discussing. The voting cards don't need to know what it is. 
The group uses Ballot to take a vote, enter the result into whatever management system they leverage, that's it.

## Development setup

### Prerequisites 
  * Computer technology
  * Docker Compose
  * Node
  * Go 1.13

### Build for Development

To invoke NPM:
```bash
make build
```

To compile the Go code and run the app on port `8080`:

```bash
make run
```

Dev and test Redis containers will be brought up when running `make run` or `make up`. Run `make down` to stop the containers. You do NOT need to have local Redis running - the `run` command will bring up a Redis container.


### Build with Docker

`docker build .`

Note that this will install local Redis in the container, but that instance can be ignored if you configure Redis with environment variables (see below).

### Environment Variables

  * HTTP_PORT - dictates which port the application will run on
  * HTTP_HOST - used to correctly display the session URL (does not affect the behavior)
  * REDIS_URL - full Redis URL (with `redis://`). Otherwise will connect to Docker Redis on the default port
  * ENV - context environment. `test`, `development`, or `production`. You can ignore this.

## Running Tests

    make test

### Redis Schema

#### user:{user_id} -> Hash 

User state for a session is stored here, and yes, this assumes that a user can only vote in one session.

| Field    | Type/Example          |
|----------|-----------------------|
| id       | UUID                  |
| name     | String                |
| estimate | String                |
| joined   | String (datetime)     |

`estimate` is an empty string by default. 
`joined` is used to initially sort users in a session by the order in which they had joined.

#### session:{session_id}:users -> Set

A set of users in this current session.

#### session:{session_id}:vote_count -> Int

Number of users in a session who cast a vote.

#### session:{session_id}:tally -> String

Final vote tally.

#### session:{session_id}:voting -> Int

  * 0 - Not voting (idle before start, or vote finished) 
  * 1 - Voting
