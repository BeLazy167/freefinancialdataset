# Documentation site

This directory contains the Mintlify documentation for financialdatasets.rip,
published at [docs.financialdatasets.rip](https://docs.financialdatasets.rip).
The separate `website/` directory contains the landing and comparison pages
served by the Go API server.

## Files

- `docs.json` defines navigation, branding, and API reference configuration.
- `overview/` and `integrations/` explain setup, coverage, and client connections.
- `datasets/` contains pages grouped by dataset and endpoint.
- `guides/` contains task guides and operational documentation.
- `mcp-tools/overview.mdx` documents the tools defined in `../go/mcpserver/tool_schemas.json`.
- `api-reference/openapi.json` is a copy of `../docs/openapi.json`.

## Update API documentation

Edit `docs/openapi.json` at the repository root, then synchronize the copy:

```bash
cp docs/openapi.json docs-site/api-reference/openapi.json
cmp docs/openapi.json docs-site/api-reference/openapi.json
```

Keep dataset page `openapi:` references and `docs.json` navigation consistent
with the specification. When MCP schemas change, update the tool tables in
`mcp-tools/overview.mdx` to match `go/mcpserver/tool_schemas.json`.

## Preview and validate

With the Mintlify CLI installed, run from this directory:

```bash
mint dev
mint broken-links
mint validate
```

## Publish

Connect the repository to Mintlify with `docs-site/` as its documentation root.
Deployment is managed by that integration, separately from the Go server's
Fly.io workflow. Forks need their own Mintlify project and domain configuration.
