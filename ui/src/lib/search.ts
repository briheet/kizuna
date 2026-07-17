import type { Evidence, SearchResponse } from '../types'
import { apiUrl } from './api'

export async function checkApiHealth(signal?: AbortSignal) {
  const response = await fetch(apiUrl('/api/v1/health'), { signal })
  if (!response.ok) return false
  const body = (await response.json()) as { status?: string }
  return body.status === 'ok'
}

export async function searchKnowledge(query: string, limit = 6): Promise<{ summary: string; evidence: Evidence[] }> {
  const response = await fetch(apiUrl('/api/v1/search'), {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ query, limit }),
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Search failed (${response.status})`)
  }

  const body = (await response.json()) as SearchResponse
  return {
    summary: body.summary,
    evidence: (body.results ?? []).map((result) => ({
      id: result.chunk_id,
      title: result.title || 'Untitled source',
      content: result.content,
      sourceLink: result.source_link,
      sourceType: formatSourceType(result.node_type, result.source_type),
      relevance: Math.max(0, Math.min(1, 1 - result.distance)),
    })),
  }
}

function formatSourceType(nodeType: string, sourceType: string) {
  const value = nodeType || sourceType || 'documentation'
  return value
    .split('_')
    .filter(Boolean)
    .map((part) => part === 'github' ? 'GitHub' : part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}
