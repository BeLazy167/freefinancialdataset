# Monid US Finance MCP

A Go server for US financial data, available through REST and the Model Context Protocol. MCP lets an AI client request data through named tools such as `get_income_statement` and `get_stock_prices`.

The server implements the 27 Financial Datasets MCP tool names and exposes compatible REST routes. It sources data independently through Monid and converts provider responses into Financial Datasets response shapes. It is not affiliated with Financial Datasets. Matching the interface does not mean matching its data coverage, history, or licensing.

- [API](https://financialdatasets.rip)
- [Documentation](https://docs.financialdatasets.rip)
- MCP endpoint: `https://financialdatasets.rip/mcp`, also available at `/api`

## Make a request

Send your Monid API key in the `X-API-KEY` header. When a request needs an upstream provider call, Monid charges your wallet.

```bash
export MONID_API_KEY='your-monid-api-key'

curl -H "X-API-KEY: $MONID_API_KEY" \
  'https://financialdatasets.rip/financials/income-statements?ticker=AAPL&period=annual&limit=1'
```

This asks for Apple's latest annual income statement. Change `ticker` to select a company, `period` to select the reporting period, and `limit` to control the number of records.

For an MCP client, use the endpoint above and configure the same header. To print connection examples for the hosted server:

```bash
make connect URL=https://financialdatasets.rip
```

## Run locally

Install Go 1.22 or newer and Make, then run these commands:

```bash
git clone https://github.com/BeLazy167/freefinancialdataset.git
cd freefinancialdataset
make run
```

The server listens on port 8080 and serves REST, MCP, and the static website. `make run` sets the paths to the provider allowlist and website for you. The server can start without an API key. Send your key with each data request.

In another terminal:

```bash
curl http://localhost:8080/healthz

export MONID_API_KEY='your-monid-api-key'
curl -H "X-API-KEY: $MONID_API_KEY" \
  'http://localhost:8080/financials/income-statements?ticker=AAPL&period=annual&limit=1'
```

For local MCP access, use `http://localhost:8080/mcp` with the same header.

## How requests work

REST routes and MCP tools call the same Go service handlers. The handlers validate arguments, select an allowed provider endpoint, and call Monid with the caller's key. If a provider runs asynchronously, the server polls for its result before returning the data.

The service converts each result into the expected response shape. It omits optional fields it cannot source and returns an error when it cannot fulfill a request. For filing sections, it parses SEC item headings locally to extract the requested text.

A request can need several upstream calls. Cash flow statements, for example, use MarketBeat after checks against SEC filings found incorrect subtotals in the normalized feed. This adds a provider call. Income statements and balance sheets use the normalized feed.

### Caching and cost

There is no fixed cost per API request. Cost depends on the providers called and whether the data is already cached.

REST responses are cached in memory and shared across callers by default. A cache hit avoids another provider call. Set `CACHE_URL` to use Redis or Upstash across machines and deployments. Set `CACHE_PER_CALLER=1` to keep separate REST cache entries for each caller.

To record upstream run IDs, status, and measured costs locally:

```bash
RECEIPTS_PATH=../receipts/ledger.jsonl make run
```

The path is relative to `go/`, where `make run` starts the server. Receipt recording is optional. An unwritable ledger does not block a response.

## Coverage and limits

The server provides financial statements, stock prices, filings and filing sections, ownership data, news, screening, and other financial data. The [MCP tool definitions](go/mcpserver/tool_schemas.json) and [OpenAPI specification](docs/openapi.json) describe the available operations and inputs.

Data availability and freshness depend on the source. Some tools support fewer filters or return fewer fields than Financial Datasets. For example, stock screening supports `exchange` and `market_cap` equality filters, and cash flow statements omit `share_based_compensation` and `ending_cash_balance` when the source does not provide them.

This is not a real-time trading feed. It does not provide tick WebSockets or guarantee the same history, redistribution rights, or uptime as Financial Datasets. Check the [route notes](docs/openapi-notes.md) before switching an existing client.

## Configure and deploy

[.env.example](.env.example) lists the settings for authentication, rate limits, browser origins, caching, and receipts. Export the variables you need before starting the server. It does not load `.env` files automatically.

To deploy on Fly.io, install `flyctl`, sign in, and choose an app name:

```bash
make deploy APP=your-app-name
make connect APP=your-app-name
```

`make deploy` builds remotely, deploys the server, and checks that `/healthz` reports the current commit. You can repeat that check with `make verify APP=your-app-name`.

For another container host, `make docker` builds the image defined by [Dockerfile](Dockerfile). The container runs one Go binary and serves the API and website together.

## Work on the server

```bash
make build   # compile to bin/server
make test    # run tests with the race detector
make vet     # check formatting and run go vet
make help    # list available targets
```

| Directory | What it owns |
| --- | --- |
| `go/cmd/server` | Startup and configuration |
| `go/httpapi` | REST routes, authentication, caching, and rate limits |
| `go/mcpserver` | MCP transport and tool definitions |
| `go/service` | Shared tool handlers and provider data conversion |
| `go/monid` | Monid requests and asynchronous run handling |
| `go/fd` | Response shapes, pagination, and receipts |
| `website` | Static website served by the binary |
| `docs-site` | Documentation site |

See [CONTRIBUTING.md](CONTRIBUTING.md) for route registration and test conventions.

## License

[MIT](LICENSE). See [third-party notices](THIRD_PARTY_NOTICES.md) for source and attribution details.
