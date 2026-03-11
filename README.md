# GTFS Viewer

A Go/Gin web application for visualising [GTFS](https://gtfs.org/) (General Transit Feed Specification) data on an interactive map.


# Version 1.0.0

First release

## Features

- Upload and manage multiple GTFS zip files
- Visualise stops and trip shapes on a Mapbox map
- Browse any GTFS file (stops, routes, trips, calendar, etc.) in a paginated table
- Click a stop to see all routes and trips serving it (including transfers)
- View trip details with stop times and frequencies
- Highlight stops and trip lines on hover
- Load a GTFS file directly at startup via CLI flag

## Requirements

- Go 1.25+
- Node.js 24+ (for building the frontend)
- A [Mapbox](https://mapbox.com/) access token

## Mapbox API Key

The Mapbox access token can be provided in two ways, in order of precedence:

1. **Build-time env var** — set `VITE_MAPBOX_KEY` when building the frontend:
   ```sh
   VITE_MAPBOX_KEY=pk.your_token npm run build
   ```

2. **URL query parameter** — append `?api_key=YOUR_MAPBOX_TOKEN` when opening the app:
   ```
   http://localhost:8080/?api_key=YOUR_MAPBOX_TOKEN
   ```
   Useful when sharing a deployment without rebuilding.

## Building

### Quick build (all platforms)

```sh
bash build.sh
```

This builds the frontend and cross-compiles Go binaries for Linux, macOS, and Windows (amd64 + arm64) into `./dist/`.

### Manual build

#### 1. Build the frontend

```sh
cd viewer
npm install
npm run build
cd ..
```

This produces `viewer/dist/`, which is embedded into the Go binary at compile time.

#### 2. Build the Go binary

```sh
go build -o gtfs-viewer .
```

## Running

### Basic

```sh
./gtfs-viewer
```

Starts the server on `0.0.0.0:8080`. Open `http://localhost:8080` in your browser, then upload a GTFS zip file via the UI.

### With a preloaded GTFS file

```sh
./gtfs-viewer -f path/to/feed.zip
```

The feed is loaded at startup. The upload UI is disabled and the feed is selected automatically.

### Environment variables

| Variable          | Default                                              | Description                                      |
|-------------------|------------------------------------------------------|--------------------------------------------------|
| `HOST`            | *(all interfaces)*                                   | Host/IP to bind to                               |
| `PORT`            | `8080`                                               | Port to listen on                                |
| `TRUSTED_PROXIES` | `127.0.0.1,::1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16` | Comma-separated CIDRs/IPs of trusted reverse proxies |

### Example

```sh
PORT=8888 HOST=127.0.0.1 ./gtfs-viewer -f feed.zip
```

## Reverse Proxy

The server respects `X-Forwarded-For` and `X-Real-IP` headers from trusted proxies. Set `TRUSTED_PROXIES` to the CIDR(s) of your proxy.

Example nginx config:

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header Host $host;
}
```

## API

| Method | Endpoint                        | Description                                         |
|--------|---------------------------------|-----------------------------------------------------|
| GET    | `/gtfs/`                        | List uploaded GTFS feeds                            |
| GET    | `/gtfs/config`                  | Server config (`withFile`, `fileName`)              |
| POST   | `/gtfs/upload`                  | Upload a GTFS zip (`multipart/form-data`, field `file`). Add `?validate=true` to run full validation |
| GET    | `/gtfs/files/stops`             | All stops as GeoJSON `FeatureCollection<Point>`. Add `?geojson=true` for full collection, `?trip=<id>` to filter by trip |
| GET    | `/gtfs/files/:fileName`         | Paginated file contents (`?page=1&gtfs=<name>`)     |
| GET    | `/gtfs/files/:fileName/:id`     | Single record by ID                                 |
| GET    | `/gtfs/stop/:stopId`            | Routes serving a stop (including transfers)         |
| GET    | `/gtfs/route/:routeId`          | Trips on a route                                    |
| GET    | `/gtfs/trip/:tripId`            | Trip detail: stop times and frequencies. Add `?geojson=true` for `Feature<LineString>` |
| GET    | `/gtfs/trip`                    | All trips as GeoJSON `FeatureCollection<LineString>` coloured by route |

## Development

Run the Go server and the Vite dev server separately:

```sh
# Terminal 1 — Go backend
go run .

# Terminal 2 — Vite frontend (proxies /gtfs to localhost:8080)
cd viewer
npm run dev
```

The Vite dev server proxies `/gtfs` requests to `http://localhost:8080`, so the frontend and backend work together without rebuilding.

## Tech Stack

- **Backend**: Go, [Gin](https://github.com/gin-gonic/gin), [gtfs-parser](https://github.com/sutanmufti/gtfs-parser)
- **Frontend**: Svelte 5, Vite, Tailwind CSS, Mapbox GL JS, Lucide icons
