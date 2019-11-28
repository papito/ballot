# Ballot

## Development setup

### Build
```bash
make build
```

### Redis Schema

#### user:{user_id} -> Hash 

User state for a session is stored here, and yes, this assumes that a user can only vote in one session.
`estimate` is an empty string by default. 

| Field    | Type/Example          |
|----------|-----------------------|
| id       | UUID                  |
| name     | String                |
| estimate | String                |
| joined   | String (datetime)     |


`joined` is used to initially sort users in a session by the order in which they had joined.

#### session:{session_id}:users -> Set

A set of users in this current session.

#### session:{session_id}:vote_count -> Integer

Number of users in a session who cast a vote.

#### session:{session_id}:voting -> Flag

  * 0 - Not voting (idle before start, or vote finished) 
  * 1 - Voting
