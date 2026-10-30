# Torznab

TorrServer can talk to **Torznab** indexers so you can search for torrents from tools like **Jackett** and **Prowlarr**, including searching several configured indexers at once.

Configure it in the web UI: **Settings → Search**.

## Indexer parameters

Each Torznab indexer needs:

- **Host URL**: Torznab base URL. TorrServer always calls the `/api` endpoint and accepts Host with or without a trailing `/api` or `/api/` (it will not be doubled).
  - Jackett example (either form works):

  ```shell
  http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/
  http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/api
  ```

  TorrServer normalizes this to `…/torznab/api` when the path does not already end with `/api`.

  - Prowlarr example:

  ```shell
  http://localhost:9696/1
  ```

  That becomes `http://localhost:9696/1/api`.

  - Other indexers often document the full endpoint with `/api` (e.g. `http://127.0.0.1:9117/torznab/api`). You can paste that as-is.
  - A missing `http://` / `https://` scheme is added automatically.

- **API Key**: the key from your Torznab indexer manager.
- **Name** (optional): label shown in the UI instead of the host URL.
- **Categories** / **CatType** (optional, Pro mode in the UI):
  - `default` — Movies and TV (`5000,2000`)
  - `manual` — comma-separated Torznab category IDs
  - `all` — no category filter

Search requests are not limited to Movies/TV categories; results come from all categories the indexer returns.

In the web **Search** dialog, **All Trackers** queries every configured Torznab indexer and, when RuTor search is also enabled, merges RuTor results into the same list.

## Enabling Torznab search

1. Open **Settings**.
2. Open the **Search** tab.
3. Turn on **Enable Torznab Search**.
4. Enter **Host URL** and **API Key**, then **Add Server** for each indexer.
5. **Save** settings.
