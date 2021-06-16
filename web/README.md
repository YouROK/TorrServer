# TorrServer web client

### How to start project

0. ignore first two steps if the server is on `localhost`
1. duplicate `.env_example` and rename it to `.env`
2. in `.env` file add server address to `REACT_APP_SERVER_HOST` (without last "/")
> `http://192.168.78.4:8090` - correct
>
> `http://192.168.78.4:8090/` - wrong
3. in `.env` file add TMDB api key
4. `yarn start`

### Eslint
> Prettier will fix the code every time the code is saved

- `yarn lint` - to find all linting problems
- `yarn fix` - to fix code