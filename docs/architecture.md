# Architecture

The server is one Go binary. REST routes and MCP tools share service handlers,
which request provider data through Monid and convert it to Financial Datasets
response shapes. Compatibility covers the interface; coverage and data quality
depend on the provider. See [compatibility notes](compatibility.md).

## Request flow

1. The HTTP layer authenticates the caller, applies rate limits, and checks the response cache.
2. REST routes or MCP tools dispatch to the shared service handlers.
3. Handlers validate inputs and select provider endpoints from the discovery allowlist.
4. The Monid client runs requests with the caller's key and polls asynchronous runs.
5. Provider adapters validate and normalize the returned data.
6. The transport formats the response and pagination links.

Provider errors propagate rather than becoming invented data. An empty result
is distinct from malformed data. Some handlers compose multiple sources or use
a fallback for a documented coverage gap.

## Package boundaries

| Package | Responsibility |
| --- | --- |
| `go/cmd/server` | Configuration, startup, and health checks |
| `go/httpapi` | REST routing, authentication, rate limits, CORS, response caching, and static files |
| `go/mcpserver` | MCP protocol and tool schemas |
| `go/service` | Input validation, provider selection, orchestration, and upstream caching |
| `go/providers` | Provider response parsing and normalization |
| `go/monid` | Monid HTTP requests, polling, run errors, and artifacts |
| `go/fd` | Response types, pagination, and optional receipts |

The server uses Go's standard library. Python scripts in `tools/` are optional
maintenance utilities, not runtime dependencies.

## Caches and receipts

The service caches upstream runs by provider, endpoint, and input. The REST
layer also caches responses. Responses are shared across callers by default;
`CACHE_PER_CALLER=1` separates REST cache entries. Optional Redis or Upstash
storage shares cache entries across machines and deployments.

Set `RECEIPTS_PATH` to record upstream calls, run IDs, and measured costs.
Receipt writing is optional and best-effort. Keep generated ledgers out of Git.
See [configuration](../.env.example) for the available settings.

## Website and documentation

`website/` is served by the Go binary. `docs-site/` is the separate Mintlify
project. The canonical OpenAPI file is `docs/openapi.json`; its copy at
`docs-site/api-reference/openapi.json` must remain identical.
