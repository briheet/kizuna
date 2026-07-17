import type { JobsStatusResponse } from '../types'
import { apiUrl } from './api'

export type GithubIngestionInput = {
  owner: string
  repo: string
  scopes: string[]
  since?: string
  limit: number
}

export type CreateIngestionResponse = {
  topic_id: string
  source_type: 'github'
  jobs_created: number
  state: 'available'
}

export async function createGithubIngestion(input: GithubIngestionInput): Promise<CreateIngestionResponse> {
  const sourceLink = `https://github.com/${input.owner}/${input.repo}`
  const response = await fetch(apiUrl('/api/v1/createJobs'), {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({
      source_type: 'github',
      name: `${input.owner}/${input.repo}`,
      source_link: sourceLink,
      scope: input.scopes,
      config: {
        owner: input.owner,
        repo: input.repo,
        since: input.since ? new Date(`${input.since}T00:00:00Z`).toISOString() : '',
        limit: input.limit,
        page_size: Math.min(100, input.limit),
        page: 1,
      },
    }),
  })

  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Could not start sync (${response.status})`)
  }

  return response.json() as Promise<CreateIngestionResponse>
}

export async function getIngestionStatus(topicId: string, signal?: AbortSignal): Promise<JobsStatusResponse> {
  const query = new URLSearchParams({ topic_id: topicId, source_type: 'github', limit: '10' })
  const response = await fetch(apiUrl(`/api/v1/jobsStatus?${query}`), { signal })
  if (!response.ok) {
    const message = await response.text()
    throw new Error(message || `Could not load sync status (${response.status})`)
  }
  return response.json() as Promise<JobsStatusResponse>
}

export function parseGithubRepository(value: string) {
  const normalized = value.trim().replace(/\.git$/, '').replace(/^git@github\.com:/, 'https://github.com/')
  const withProtocol = normalized.includes('://') ? normalized : `https://github.com/${normalized}`

  try {
    const url = new URL(withProtocol)
    const [owner, repo, ...rest] = url.pathname.split('/').filter(Boolean)
    if (url.hostname !== 'github.com' || !owner || !repo || rest.length > 0) return null
    return { owner, repo }
  } catch {
    return null
  }
}
