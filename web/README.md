# TorrServer web client

### How to start project

0. ignore first two steps if the server is on `localhost`
1. duplicate `.env_example` and rename it to `.env`
2. in `.env` file add server address to `REACT_APP_SERVER_HOST` (without last "/")
> `http://192.168.78.4:8090` - correct
>
> `http://192.168.78.4:8090/` - wrong
3. in `.env` file add TMDB api key
4. `NODE_OPTIONS=--openssl-legacy-provider yarn start`

### Eslint
> Prettier will fix the code every time the code is saved

- `yarn lint` - to find all linting problems
- `yarn fix` - to fix code

### How images were generated
`npx pwa-asset-generator public/logo.png public -m public/site.webmanifest -p "calc(50vh - 25%) calc(50vw - 25%)" -b "linear-gradient(135deg, rgb(50,54,55), rgb(84,90,94))" -q 100 -i public/index.html -f`