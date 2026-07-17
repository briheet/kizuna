import type { Evidence, SearchResponse } from '../types'

const apiBase = import.meta.env.VITE_API_BASE_URL ?? ''

export async function checkApiHealth(signal?: AbortSignal) {
  const response = await fetch(`${apiBase}/api/v1/health`, { signal })
  if (!response.ok) return false
  const body = (await response.json()) as { status?: string }
  return body.status === 'ok'
}

export async function searchKnowledge(query: string, limit = 6): Promise<Evidence[]> {
  const response = await fetch(`${apiBase}/api/v1/search`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ query, limit }),
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Search failed (${response.status})`)
  }

  const body = (await response.json()) as SearchResponse
  return (body.results ?? []).map((result) => ({
    id: result.chunk_id,
    title: result.title || 'Untitled source',
    content: result.content,
    sourceLink: result.source_link,
    sourceType: inferSourceType(result.source_link, result.title),
    relevance: Math.max(0, Math.min(1, 1 - result.distance)),
  }))
}

export function composeAnswer(evidence: Evidence[]) {
  if (evidence.length === 0) {
    return "I couldn't find a strong match in this knowledge space. Try adding a service, incident number, repository, or narrower time range."
  }

  const primary = evidence[0]
  const supporting = evidence.slice(1, 3).map((item) => item.title)
  const supportLine = supporting.length
    ? `\n\nRelated records include ${supporting.map((title) => `“${title}”`).join(' and ')}.`
    : ''

  return `The closest match is “${primary.title}”. ${primary.content}${supportLine}`
}

function inferSourceType(link: string, title: string) {
  const value = `${link} ${title}`.toLowerCase()
  if (value.includes('github') || value.includes('pull') || value.includes('pr #')) return 'GitHub'
  if (value.includes('slack')) return 'Slack'
  if (value.includes('discord')) return 'Discord'
  if (value.includes('incident')) return 'Incident'
  if (value.includes('adr')) return 'ADR'
  if (value.includes('runbook')) return 'Runbook'
  return 'Documentation'
}
