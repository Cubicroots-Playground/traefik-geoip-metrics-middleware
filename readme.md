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
