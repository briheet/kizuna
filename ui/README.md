# Kizuna UI

The chat interface for Kizuna's engineering memory graph.

```bash
npm install
npm run dev
```

The ignored `.env` file points local development at the deployed backend on
`http://167.233.61.32:4000`. Remove `VITE_API_BASE_URL` to use Vite's local
`/api` proxy to `http://localhost:4000` instead.

Chat searches the full engineering knowledge graph and never fabricates fallback
results when the API is unavailable.
