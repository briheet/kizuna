import {
  AlertCircle,
  ArrowLeft,
  BookOpenText,
  Check,
  CircleDashed,
  ExternalLink,
  GitFork,
  Hash,
  LoaderCircle,
  MessageSquareText,
  PanelsTopLeft,
  Plus,
  RefreshCw,
  X,
} from 'lucide-react'
import { useCallback, useEffect, useRef, useState } from 'react'
import type { FormEvent } from 'react'
import type { ConnectionStatus, IngestionState, JobsStatusResponse, KnowledgeSource } from '../types'
import { createGithubIngestion, getIngestionStatus, parseGithubRepository } from '../lib/ingestion'
import { loadSources, saveSources } from '../lib/sources'

type SourcesPanelProps = {
  connectionStatus: ConnectionStatus
  open: boolean
  onClose: () => void
}

const githubScopes = [
  { id: 'repository', label: 'Repository', detail: 'Metadata and ownership' },
  { id: 'pull_requests', label: 'Pull requests', detail: 'Changes and reviews' },
  { id: 'issues', label: 'Issues', detail: 'Work and discussions' },
  { id: 'commits', label: 'Commits', detail: 'Code history' },
  { id: 'releases', label: 'Releases', detail: 'Published versions' },
]

const connectors = [
  { id: 'github', label: 'GitHub', detail: 'Repositories, pull requests, issues, commits, and releases', icon: GitFork, enabled: true },
  { id: 'slack', label: 'Slack', detail: 'Channels, threads, files, and reactions', icon: Hash, enabled: false },
  { id: 'discord', label: 'Discord', detail: 'Guilds, channels, messages, and threads', icon: MessageSquareText, enabled: false },
  { id: 'jira', label: 'Jira', detail: 'Projects, issues, comments, and attachments', icon: PanelsTopLeft, enabled: false },
  { id: 'confluence', label: 'Confluence', detail: 'Spaces, pages, comments, and attachments', icon: BookOpenText, enabled: false },
]

export function SourcesPanel({ connectionStatus, open, onClose }: SourcesPanelProps) {
  const [view, setView] = useState<'overview' | 'github'>('overview')
  const [sources, setSources] = useState<KnowledgeSource[]>(loadSources)
  const [refreshing, setRefreshing] = useState(false)
  const [statusError, setStatusError] = useState('')
  const sourcesRef = useRef(sources)

  useEffect(() => {
    sourcesRef.current = sources
    saveSources(sources)
  }, [sources])

  const refreshStatuses = useCallback(async () => {
    const current = sourcesRef.current
    if (current.length === 0) return

    setRefreshing(true)
    try {
      const updates = await Promise.all(current.map(async (source) => {
        const response = await getIngestionStatus(source.topicId)
        return updateSourceStatus(source, response)
      }))
      setSources(updates)
      setStatusError('')
    } catch (error) {
      setStatusError(error instanceof Error ? error.message : 'Could not refresh source activity.')
    } finally {
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    if (!open) return
    void refreshStatuses()
    const interval = window.setInterval(() => {
      const hasActiveSync = sourcesRef.current.some((source) => source.state === 'queued' || source.state === 'syncing')
      if (hasActiveSync) void refreshStatuses()
    }, 3000)
    return () => window.clearInterval(interval)
  }, [open, refreshStatuses])

  useEffect(() => {
    if (!open) setView('overview')
  }, [open])

  function sourceCreated(source: KnowledgeSource) {
    setSources((current) => [source, ...current])
    setView('overview')
    window.setTimeout(() => void refreshStatuses(), 700)
  }

  return (
    <>
      <div className={`sources-backdrop ${open ? 'is-open' : ''}`} onClick={onClose} />
      <aside aria-label="Knowledge sources" className={`sources-panel ${open ? 'is-open' : ''}`}>
        <header className="sources-header">
          <div>
            <span className="eyebrow">Knowledge graph</span>
            <h2>{view === 'github' ? 'Connect GitHub' : 'Sources'}</h2>
          </div>
          <button className="icon-button" onClick={onClose} title="Close sources" type="button"><X /></button>
        </header>

        {view === 'github' ? (
          <GithubSetup
            connectionStatus={connectionStatus}
            onBack={() => setView('overview')}
            onCreated={sourceCreated}
          />
        ) : (
          <div className="sources-scroll">
            <div className="source-metrics">
              <Metric label="Connected" value={sources.length} />
              <Metric label="Syncing" value={sources.filter((source) => source.state === 'queued' || source.state === 'syncing').length} />
              <Metric label="Needs attention" value={sources.filter((source) => source.state === 'failed').length} />
            </div>

            <section className="source-section">
              <div className="section-heading">
                <div>
                  <h3>Connected sources</h3>
                  <span>{sources.length ? 'Sync activity and indexed scope' : 'No sources connected'}</span>
                </div>
                {sources.length > 0 && (
                  <button className="icon-button" disabled={refreshing} onClick={() => void refreshStatuses()} title="Refresh source status" type="button">
                    <RefreshCw className={refreshing ? 'is-spinning' : ''} />
                  </button>
                )}
              </div>

              {statusError && <div className="source-alert"><AlertCircle /><span>{statusError}</span></div>}

              {sources.length > 0 ? (
                <div className="connected-list">
                  {sources.map((source) => <ConnectedSource key={source.id} source={source} />)}
                </div>
              ) : (
                <div className="sources-empty">
                  <CircleDashed />
                  <strong>Connect your first source</strong>
                  <span>GitHub is ready for this deployment.</span>
                </div>
              )}
            </section>

            <section className="source-section connector-section">
              <div className="section-heading">
                <div>
                  <h3>Add a source</h3>
                  <span>Provider credentials are managed by the worker service</span>
                </div>
              </div>
              <div className="connector-list">
                {connectors.map((connector) => {
                  const Icon = connector.icon
                  return (
                    <button
                      className="connector-row"
                      disabled={!connector.enabled}
                      key={connector.id}
                      onClick={() => connector.id === 'github' && setView('github')}
                      type="button"
                    >
                      <span className="connector-icon"><Icon /></span>
                      <span className="connector-copy"><strong>{connector.label}</strong><small>{connector.detail}</small></span>
                      {connector.enabled ? <span className="connector-action"><Plus /> Connect</span> : <span className="connector-unavailable">Not configured</span>}
                    </button>
                  )
                })}
              </div>
            </section>
          </div>
        )}
      </aside>
    </>
  )
}

function GithubSetup({ connectionStatus, onBack, onCreated }: {
  connectionStatus: ConnectionStatus
  onBack: () => void
  onCreated: (source: KnowledgeSource) => void
}) {
  const [repository, setRepository] = useState('')
  const [scopes, setScopes] = useState(githubScopes.map((scope) => scope.id))
  const [since, setSince] = useState('')
  const [limit, setLimit] = useState(1000)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function submit(event: FormEvent) {
    event.preventDefault()
    const parsed = parseGithubRepository(repository)
    if (!parsed) {
      setError('Enter a GitHub repository URL or owner/repository.')
      return
    }
    if (scopes.length === 0) {
      setError('Select at least one content type.')
      return
    }

    setSubmitting(true)
    setError('')
    try {
      const response = await createGithubIngestion({ ...parsed, scopes, since, limit })
      const now = Date.now()
      onCreated({
        id: crypto.randomUUID(),
        topicId: response.topic_id,
        sourceType: 'github',
        name: `${parsed.owner}/${parsed.repo}`,
        sourceLink: `https://github.com/${parsed.owner}/${parsed.repo}`,
        scopes,
        state: 'queued',
        counts: [{ state: 'available', count: response.jobs_created }],
        createdAt: now,
        updatedAt: now,
      })
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Could not start the GitHub sync.')
    } finally {
      setSubmitting(false)
    }
  }

  function toggleScope(id: string) {
    setScopes((current) => current.includes(id) ? current.filter((scope) => scope !== id) : [...current, id])
  }

  return (
    <form className="github-setup" onSubmit={submit}>
      <div className="setup-scroll">
        <button className="setup-back" onClick={onBack} type="button"><ArrowLeft /> All sources</button>

        <div className="setup-provider">
          <span><GitFork /></span>
          <div><strong>GitHub repository</strong><small>Uses the token configured on the ingestion worker.</small></div>
          <i className={`status-dot status-${connectionStatus}`} />
        </div>

        <label className="field-label" htmlFor="github-repository">Repository</label>
        <div className="input-prefix">
          <GitFork />
          <input
            autoComplete="off"
            autoFocus
            id="github-repository"
            onChange={(event) => setRepository(event.target.value)}
            placeholder="github.com/owner/repository"
            type="text"
            value={repository}
          />
        </div>

        <fieldset className="scope-fieldset">
          <legend>Content to index</legend>
          <div className="scope-options">
            {githubScopes.map((scope) => {
              const checked = scopes.includes(scope.id)
              return (
                <label className={checked ? 'is-checked' : ''} key={scope.id}>
                  <input checked={checked} onChange={() => toggleScope(scope.id)} type="checkbox" />
                  <span className="check-box">{checked && <Check />}</span>
                  <span><strong>{scope.label}</strong><small>{scope.detail}</small></span>
                </label>
              )
            })}
          </div>
        </fieldset>

        <div className="setup-fields">
          <label>
            <span className="field-label">Changed since</span>
            <input max={new Date().toISOString().slice(0, 10)} onChange={(event) => setSince(event.target.value)} type="date" value={since} />
          </label>
          <label>
            <span className="field-label">Item limit</span>
            <input max="10000" min="1" onChange={(event) => setLimit(Number(event.target.value))} type="number" value={limit} />
          </label>
        </div>

        {error && <div className="source-alert is-error"><AlertCircle /><span>{error}</span></div>}
      </div>

      <footer className="setup-footer">
        <button className="secondary-button" onClick={onBack} type="button">Cancel</button>
        <button className="primary-button" disabled={submitting || connectionStatus === 'offline'} type="submit">
          {submitting ? <LoaderCircle className="is-spinning" /> : <RefreshCw />}
          {submitting ? 'Starting sync' : 'Start sync'}
        </button>
      </footer>
    </form>
  )
}

function ConnectedSource({ source }: { source: KnowledgeSource }) {
  const progress = sourceProgress(source)
  return (
    <article className="connected-source">
      <div className="connected-source-top">
        <span className="connector-icon"><GitFork /></span>
        <div className="connected-source-name">
          <strong>{source.name}</strong>
          <span>{source.scopes.map(formatScope).join(' · ')}</span>
        </div>
        <SourceState state={source.state} />
      </div>
      <div className="sync-progress"><i style={{ width: `${progress}%` }} /></div>
      <div className="connected-source-meta">
        <span>{syncDetail(source)}</span>
        <a href={source.sourceLink} rel="noreferrer" target="_blank" title="Open repository"><ExternalLink /></a>
      </div>
      {source.lastError && <p className="sync-error">{source.lastError}</p>}
    </article>
  )
}

function SourceState({ state }: { state: IngestionState }) {
  if (state === 'ready') return <span className="source-state is-ready"><Check /> Ready</span>
  if (state === 'failed') return <span className="source-state is-failed"><AlertCircle /> Failed</span>
  if (state === 'syncing') return <span className="source-state is-syncing"><LoaderCircle className="is-spinning" /> Syncing</span>
  return <span className="source-state"><CircleDashed /> Queued</span>
}

function Metric({ label, value }: { label: string; value: number }) {
  return <div><strong>{value}</strong><span>{label}</span></div>
}

function updateSourceStatus(source: KnowledgeSource, response: JobsStatusResponse): KnowledgeSource {
  const counts = response.counts ?? []
  const count = (state: string) => counts.find((item) => item.state === state)?.count ?? 0
  const failures = count('failed') + count('discarded')
  const active = count('available') + count('running')
  let state: IngestionState = source.state
  if (failures > 0) state = 'failed'
  else if (count('running') > 0) state = 'syncing'
  else if (active > 0) state = 'queued'
  else if (count('completed') > 0) state = 'ready'

  return {
    ...source,
    counts,
    state,
    lastError: response.recent_failures?.[0]?.last_error,
    updatedAt: Date.now(),
  }
}

function sourceProgress(source: KnowledgeSource) {
  const total = source.counts.reduce((sum, item) => sum + item.count, 0)
  const complete = source.counts.find((item) => item.state === 'completed')?.count ?? 0
  if (source.state === 'ready') return 100
  if (source.state === 'failed') return total ? Math.round((complete / total) * 100) : 0
  if (!total) return 8
  return Math.max(8, Math.round((complete / total) * 100))
}

function syncDetail(source: KnowledgeSource) {
  const count = (state: string) => source.counts.find((item) => item.state === state)?.count ?? 0
  const total = source.counts.reduce((sum, item) => sum + item.count, 0)
  if (source.state === 'ready') return `${count('completed')} of ${total} indexing tasks complete`
  if (source.state === 'failed') return `${count('failed') + count('discarded')} indexing tasks need attention`
  if (source.state === 'syncing') return `${count('running')} active · ${count('available')} queued`
  return `${count('available')} indexing tasks queued`
}

function formatScope(scope: string) {
  return scope.split('_').map((part) => part.charAt(0).toUpperCase() + part.slice(1)).join(' ')
}
