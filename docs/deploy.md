# Deploy

The Go binary serves the REST API, MCP endpoints at `/mcp` and `/api`, health
checks, and `website/`. Mintlify hosts `docs-site/` separately. There is no
separate static host required for the API's landing page.

## Run locally

From the repository root:

```bash
make run
```

This supplies the website and provider allowlist paths and listens on port 8080.
The server does not load `.env` files automatically. Export settings from
[.env.example](../.env.example) before starting it.

## Deploy to Fly.io

Install `flyctl`, sign in, and choose an app name:

```bash
make deploy APP=your-app-name
make verify APP=your-app-name
make connect APP=your-app-name
```

The Makefile creates the app if needed, builds remotely, stamps the current
commit into the binary, and checks `/healthz` after deployment. The Dockerfile
builds a static Go binary and packages it with the website and provider allowlist.

Requests use the caller's Monid key in `X-API-KEY`. Optional server settings
include `API_KEYS` to restrict accepted keys, `DEMO_MONID_API_KEY` for the
operator-funded demo, `CORS_ALLOWED_ORIGINS`, and shared cache configuration.
See [.env.example](../.env.example) for details. Configure secrets on the host;
do not commit their values.

## Automatic deployment

[GitHub Actions](../.github/workflows/ci.yml) checks formatting, builds, runs
`go vet`, and tests with the race detector. A push to `main` deploys to Fly.io
only after those checks pass and when the repository has a `FLY_API_TOKEN`
Actions secret. Pull requests run verification without deploying.

A fork must configure its own Fly app in `fly.toml` and its own deployment token.
Without the token, the workflow skips deployment.

## Verify the hosted API

```bash
curl https://financialdatasets.rip/healthz

export MONID_API_KEY='your-monid-api-key'
curl -H "X-API-KEY: $MONID_API_KEY" \
  'https://financialdatasets.rip/prices?ticker=AAPL&start_date=2026-09-01&end_date=2026-09-02&interval=day'
```

Use your own deployment URL when testing a fork. The health response includes
the build version; `make verify` compares it with the current commit.

## Other hosts and documentation

`make docker` builds the container for other hosts. Expose port 8080 and supply
configuration through environment variables.

The documentation site has its own publishing setup. See
[docs-site/README.md](../docs-site/README.md).
