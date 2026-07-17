import {
  ArrowUp,
  AtSign,
  BookOpen,
  Check,
  Copy,
  Database,
  ExternalLink,
  GitBranch,
  Menu,
  PanelRightOpen,
  RotateCcw,
  Sparkles,
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import type { FormEvent, KeyboardEvent } from 'react'
import type { ConnectionStatus, Conversation, Evidence } from '../types'
import { BrandMark } from './BrandMark'

type ChatViewProps = {
  connectionStatus: ConnectionStatus
  conversation: Conversation
  evidence: Evidence[]
  evidenceOpen: boolean
  onOpenEvidence: () => void
  onOpenMenu: () => void
  onOpenSources: () => void
  onRegenerate: (messageId: string) => void
  onSend: (prompt: string) => void
  onUsePrompt: (prompt: string) => void
}

const starterPrompts = [
  { icon: <GitBranch />, label: 'Trace a decision', prompt: 'Why did we migrate event delivery to Kafka?' },
  { icon: <AtSign />, label: 'Find an owner', prompt: 'Who owns the checkout service and what changed recently?' },
  { icon: <Sparkles />, label: 'Investigate an incident', prompt: 'What caused the latest checkout latency incident?' },
  { icon: <BookOpen />, label: 'Explain a system', prompt: 'Explain the platform architecture to a new engineer.' },
]

export function ChatView(props: ChatViewProps) {
  const [prompt, setPrompt] = useState('')
  const [copiedId, setCopiedId] = useState('')
  const messagesEnd = useRef<HTMLDivElement>(null)
  const textarea = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    messagesEnd.current?.scrollIntoView({ behavior: 'smooth' })
  }, [props.conversation.messages])

  useEffect(() => {
    setPrompt('')
    textarea.current?.focus()
  }, [props.conversation.id])

  function submit(event?: FormEvent) {
    event?.preventDefault()
    const value = prompt.trim()
    if (!value) return
    props.onSend(value)
    setPrompt('')
    if (textarea.current) textarea.current.style.height = 'auto'
  }

  function keyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      submit()
    }
  }

  async function copyMessage(id: string, content: string) {
    try {
      await navigator.clipboard.writeText(content)
      setCopiedId(id)
      window.setTimeout(() => setCopiedId(''), 1400)
    } catch {
      setCopiedId('')
    }
  }

  const hasMessages = props.conversation.messages.length > 0

  return (
    <main className="chat-view">
      <header className="chat-header">
        <button className="icon-button menu-button" onClick={props.onOpenMenu} title="Open menu" type="button">
          <Menu />
        </button>
        <div className="chat-heading">
          <BrandMark compact />
          <div>
            <strong>{hasMessages ? props.conversation.title : 'Engineering memory'}</strong>
            <span className={`connection-copy status-${props.connectionStatus}`}>
              <i /> {connectionLabel(props.connectionStatus)}
            </span>
          </div>
        </div>
        <div className="header-actions">
          <button className="scope-pill" onClick={props.onOpenSources} title="Manage knowledge sources" type="button">
            <Database />
            <span>All knowledge</span>
          </button>
          {props.evidence.length > 0 && (
            <button className={`icon-button ${props.evidenceOpen ? 'is-active' : ''}`} onClick={props.onOpenEvidence} title="Open evidence" type="button">
              <PanelRightOpen />
            </button>
          )}
        </div>
      </header>

      <section className={`chat-scroll ${hasMessages ? '' : 'empty'}`}>
        {!hasMessages ? (
          <EmptyState onUsePrompt={props.onUsePrompt} />
        ) : (
          <div className="messages">
            {props.conversation.messages.map((message) => (
              <article className={`message ${message.role}`} key={message.id}>
                {message.role === 'assistant' && <BrandMark compact />}
                <div className="message-body">
                  <span className="message-author">{message.role === 'assistant' ? 'Kizuna' : 'You'}</span>
                  {message.status === 'loading' ? (
                    <Thinking />
                  ) : (
                    <>
                      <div className={`message-copy ${message.status === 'error' ? 'error' : ''}`}>
                        {message.content.split('\n').map((paragraph, index) => paragraph && <p key={`${message.id}-${index}`}>{paragraph}</p>)}
                      </div>
                      {message.evidence && message.evidence.length > 0 && (
                        <div className="source-strip">
                          {message.evidence.slice(0, 3).map((item, index) => (
                            <a href={item.sourceLink || undefined} key={item.id} rel="noreferrer" target={item.sourceLink ? '_blank' : undefined}>
                              <span>0{index + 1}</span>
                              <b>{item.title}</b>
                              {item.sourceLink && <ExternalLink />}
                            </a>
                          ))}
                          {message.evidence.length > 3 && (
                            <button onClick={props.onOpenEvidence} type="button">+{message.evidence.length - 3} more</button>
                          )}
                        </div>
                      )}
                      <div className="message-actions">
                        <button onClick={() => copyMessage(message.id, message.content)} title={message.role === 'assistant' ? 'Copy answer' : 'Copy message'} type="button">
                          {copiedId === message.id ? <Check /> : <Copy />}
                        </button>
                        {message.role === 'assistant' && (
                          <button onClick={() => props.onRegenerate(message.id)} title={message.status === 'error' ? 'Retry search' : 'Regenerate answer'} type="button">
                            <RotateCcw />
                          </button>
                        )}
                      </div>
                    </>
                  )}
                </div>
              </article>
            ))}
            <div ref={messagesEnd} />
          </div>
        )}
      </section>

      <div className="composer-wrap">
        <form className="composer" onSubmit={submit}>
          <textarea
            aria-label="Ask Kizuna"
            onChange={(event) => {
              setPrompt(event.target.value)
              event.target.style.height = 'auto'
              event.target.style.height = `${Math.min(event.target.scrollHeight, 160)}px`
            }}
            onKeyDown={keyDown}
            placeholder="Ask about a decision, incident, service, or change..."
            ref={textarea}
            rows={1}
            value={prompt}
          />
          <div className="composer-footer">
            <div className="context-label">
              <Database />
              Entire knowledge graph
            </div>
            <div className="composer-submit">
              <span>Enter to send</span>
              <button aria-label="Send message" disabled={!prompt.trim()} type="submit"><ArrowUp /></button>
            </div>
          </div>
        </form>
        <p className="composer-note">Searches the entire engineering record. Verify critical details in the cited sources.</p>
      </div>
    </main>
  )
}

function EmptyState({ onUsePrompt }: {
  onUsePrompt: (prompt: string) => void
}) {
  return (
    <div className="empty-state">
      <div className="memory-visual" aria-hidden="true">
        <span className="memory-ring ring-one" />
        <span className="memory-ring ring-two" />
        <span className="memory-ring ring-three" />
        <BrandMark compact />
        <i className="node n1" />
        <i className="node n2" />
        <i className="node n3" />
      </div>
      <span className="eyebrow">Your team’s engineering memory</span>
      <h1>Ask what happened.<br />Understand <em>why.</em></h1>
      <p>Trace decisions across code, conversations, incidents, and docs with every answer grounded in the original record.</p>
      <div className="starter-grid">
        {starterPrompts.map((item) => (
          <button key={item.label} onClick={() => onUsePrompt(item.prompt)} type="button">
            <span>{item.icon}</span>
            <div><strong>{item.label}</strong><small>{item.prompt}</small></div>
            <ArrowUp className="starter-arrow" />
          </button>
        ))}
      </div>
    </div>
  )
}

function connectionLabel(status: ConnectionStatus) {
  if (status === 'online') return 'API online'
  if (status === 'checking') return 'Checking API'
  return 'API unavailable'
}

function Thinking() {
  return (
    <div className="thinking">
      <span />
      <span />
      <span />
      <p>Tracing the knowledge graph</p>
    </div>
  )
}
