# Ballot

## Development setup

### Build
```bash
make build
```

### Redis Schema

#### user:{user_id} -> Hash 

| Field    | Type/Example          |
|----------|-----------------------|
| id       | UUID                  |
| name     | String                |
| estimate | String                |

Estimate is an empty string by default. 

#### session:{session_id}:users -> Set

A set of users in this current session.

#### session:{session_id}:user_count -> Integer

Number of users in a session.

#### session:{session_id}:vote_count -> Integer

Number of users in a session who cast a vote.

#### session:{session_id}:voting -> Flag

  * 0 - Not voting (idle before start, or vote finished) 
  * 1 - Voting
