# Kizuna UI

The chat interface for Kizuna's engineering memory graph.

```bash
npm install
npm run dev
```

Vite proxies `/api` to `http://localhost:4000`. Set `VITE_API_BASE_URL` to use another backend origin. Chat searches the full engineering knowledge graph and never fabricates fallback results when the API is unavailable.
