# Traefik GeoIP Metrics Middleware

A treafik middleware exposing prometheus metrics with geoIP data.

## Development

See [traefik example middleware](https://github.com/traefik/plugindemo).

**Running locally**

To run the middleware locally:

```bash
(cd test && docker compose up)
```

Check `whoami.localhost` for the middleware in action and `localhost:8080` for the traefik dashboard.

## Setup

Deploy the [geoip-api](https://github.com/Cubicroots-Playground/geoip-api) somewhere.

The further instructions will assume the API is runnng at `http://geoip-api:8080`.

The geoip-api will expose metrics that can be used to visualize where requests are originating from.

### Docker Compose & Swarm

Add the following command flags and labels to traefik, make sure to set the most recent version:

```yaml
traefik:
  ...
  command:
    - --experimental.plugins.geoip.moduleName=github.com/Cubicroots-Playground/traefik-geoip-metrics-middleware
    - --experimental.plugins.geoip.version=v0.0.2
  deploy:
    labels:
      - traefik.http.middlewares.mw-geoip.plugin.geoip.geoipApi=http://geoip-api:8080
```

Add the `mw-geoip` middleware to all routers that should be intercepted by the geoip middleware.

A working example using local traefik plugins is available in the `test` folder.