# ballot

## Development setup

### Install dependencies
```bash
npm install --save-dev webpack webpack-cli typescript ts-loader css-loader \
    vue-loader vue-router vue-class-component vue-property-decorator vue-mixin-decorator \
    acorn tslint node-fetch

npm install vue vue-template-compiler
```

### Build
```bash
npm run build
```

Build for debugging:
```bash
./node_modules/.bin/webpack --mode=development
```

Build for production:
```bash
./node_modules/.bin/webpack --mode=production
```

### Redis records

#### user:{user_id} -> Hash 

| Field    | Type/Example          |
|----------|-----------------------|
| id       | UUID                  |
| name     | String                |
| estimate | Integer, -1 to 100    |

Estimate `-1` is "idle" state. User has not voted yet. 
This is also the value set when a voting session restarts.

#### session:{session_id}:users -> Set

A set of users in this current session

#### session:{session_id}:voting -> Flag

  * 0 - Voting
  * 2 - Not voting (idle before start, or vote finished) 