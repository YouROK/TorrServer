# Torznab Support for TorrServer

This package implements Torznab support for TorrServer, allowing integration with indexer managers like Jackett and Prowlarr.

## Features

- **Search Integration**: Search for torrents directly from supported Torznab indexers.
- **Aggregated Search**: Search across multiple configured indexers simultaneously.

## Configuration

Torznab settings can be configured via the Web UI under `Settings > Torznab`.

### Parameters

Each Torznab indexer requires the following:

- **Host URL**: The full URL to the Torznab API endpoint.
  - Example: `http://192.168.1.10:9117/api/v2.0/indexers/all/results/torznab/` (Jackett)
  - Example: `http://localhost:9696/1/api` (Prowlarr)
  - *Note*: Ensure the URL ends with `/` or `/api` as appropriate for your indexer manager, though the server attempts to handle pathing intelligently.
- **API Key**: The API key provided by your Torznab indexer manager.

### enabling support

To enable Torznab search:
1. Go to **Settings**.
2. Navigate to the **Torznab** tab.
3. Toggle **Enable Torznab Search**.
4. Add your indexers using the "Host URL" and "API Key" fields.
5. Click **Add Server**.
6. **Save** your settings.