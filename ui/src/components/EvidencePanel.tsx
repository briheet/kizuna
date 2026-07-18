import { ArrowUpRight, BookOpenText, FileText, GitPullRequest, Hash, Network, X } from 'lucide-react'
import type { Evidence } from '../types'

type EvidencePanelProps = {
  evidence: Evidence[]
  open: boolean
  onClose: () => void
}

export function EvidencePanel({ evidence, open, onClose }: EvidencePanelProps) {
  return (
    <aside className={`evidence-panel ${open ? 'is-open' : ''}`}>
      <header>
        <div>
          <span className="eyebrow">Grounded context</span>
          <h2>Evidence</h2>
        </div>
        <button className="icon-button" onClick={onClose} title="Close evidence" type="button">
          <X />
        </button>
      </header>

      <div className="evidence-summary">
        <Network />
        <div>
          <strong>{evidence.length} retrieved {evidence.length === 1 ? 'record' : 'records'}</strong>
          <span>Ranked semantic matches</span>
        </div>
      </div>

      <div className="evidence-list">
        {evidence.map((item, index) => (
          <article className="evidence-card" key={item.id}>
            <div className="evidence-card-meta">
              <span className="source-icon">{sourceIcon(item.sourceType)}</span>
              <span>{item.sourceType}</span>
              <span className="relevance">{Math.round(item.relevance * 100)}%</span>
            </div>
            <h3>{item.title}</h3>
            <p>{item.content}</p>
            {item.sourceLink && (
              <a href={item.sourceLink} rel="noreferrer" target="_blank">
                Open source <ArrowUpRight />
              </a>
            )}
            <span className="evidence-index">0{index + 1}</span>
          </article>
        ))}
      </div>
    </aside>
  )
}

function sourceIcon(sourceType: string) {
  const type = sourceType.toLowerCase()
  if (type.includes('pull') || type.includes('github')) return <GitPullRequest />
  if (type.includes('slack') || type.includes('discord')) return <Hash />
  if (type.includes('adr') || type.includes('runbook')) return <BookOpenText />
  return <FileText />
}
