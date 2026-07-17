import { Database, MessageSquareText, PanelLeftClose, PanelLeftOpen, Plus, Search, Trash2, X } from 'lucide-react'
import { useMemo, useState } from 'react'
import type { ConnectionStatus, Conversation } from '../types'
import { BrandMark } from './BrandMark'

type SidebarProps = {
  activeId: string
  collapsed: boolean
  connectionStatus: ConnectionStatus
  conversations: Conversation[]
  mobileOpen: boolean
  onCloseMobile: () => void
  onDelete: (id: string) => void
  onNew: () => void
  onSelect: (id: string) => void
  onToggle: () => void
}

export function Sidebar(props: SidebarProps) {
  const [query, setQuery] = useState('')
  const filtered = useMemo(() => {
    const term = query.trim().toLowerCase()
    return props.conversations
      .filter((item) => !term || item.title.toLowerCase().includes(term))
      .sort((a, b) => b.updatedAt - a.updatedAt)
  }, [props.conversations, query])
  const groups = groupConversations(filtered)

  return (
    <>
      <div className={`sidebar-backdrop ${props.mobileOpen ? 'is-open' : ''}`} onClick={props.onCloseMobile} />
      <aside aria-label="Conversation sidebar" className={`sidebar ${props.mobileOpen ? 'is-open' : ''} ${props.collapsed ? 'is-collapsed' : ''}`}>
        <div className="sidebar-top">
          <BrandMark compact={props.collapsed && !props.mobileOpen} />
          <button className="icon-button sidebar-toggle desktop-only" onClick={props.onToggle} title={props.collapsed ? 'Expand sidebar' : 'Collapse sidebar'} type="button">
            {props.collapsed ? <PanelLeftOpen /> : <PanelLeftClose />}
          </button>
          <button className="icon-button mobile-only" onClick={props.onCloseMobile} title="Close menu" type="button">
            <X />
          </button>
        </div>

        <button className="new-chat-button" onClick={props.onNew} title={props.collapsed ? 'New conversation' : undefined} type="button">
          <Plus />
          <span>New conversation</span>
          <kbd>⌘ K</kbd>
        </button>

        <label className="history-search">
          <Search />
          <input
            aria-label="Search conversations"
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search conversations"
            type="search"
            value={query}
          />
        </label>

        <div className="history-list">
          {groups.map((group) => (
            <section className="history-group" key={group.label}>
              <h2>{group.label}</h2>
              {group.items.map((conversation) => (
                <div className={`history-row ${conversation.id === props.activeId ? 'is-active' : ''}`} key={conversation.id}>
                  <button
                    className="history-item"
                    onClick={() => {
                      props.onSelect(conversation.id)
                      props.onCloseMobile()
                    }}
                    title={props.collapsed ? conversation.title : undefined}
                    type="button"
                  >
                    <MessageSquareText />
                    <span>{conversation.title}</span>
                  </button>
                  <button
                    className="history-delete"
                    onClick={() => props.onDelete(conversation.id)}
                    title="Delete conversation"
                    type="button"
                  >
                    <Trash2 />
                  </button>
                </div>
              ))}
            </section>
          ))}
          {filtered.length === 0 && <p className="history-empty">No conversations found.</p>}
        </div>

        <div className="sidebar-footer">
          <div className="footer-status" title={props.collapsed ? statusLabel(props.connectionStatus) : undefined}>
            <span className="footer-knowledge-icon"><Database /></span>
            <span className="footer-knowledge-copy">
              <b>Knowledge graph</b>
              <small>{statusLabel(props.connectionStatus)}</small>
            </span>
          </div>
          <i className={`status-dot status-${props.connectionStatus}`} />
        </div>
      </aside>
    </>
  )
}

function statusLabel(status: ConnectionStatus) {
  if (status === 'online') return 'API online'
  if (status === 'checking') return 'Checking connection'
  return 'API unavailable'
}

function groupConversations(conversations: Conversation[]) {
  const now = new Date()
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime()
  const yesterdayStart = todayStart - 86_400_000
  const weekStart = todayStart - 7 * 86_400_000
  const buckets = [
    { label: 'Today', items: [] as Conversation[] },
    { label: 'Yesterday', items: [] as Conversation[] },
    { label: 'Previous 7 days', items: [] as Conversation[] },
    { label: 'Older', items: [] as Conversation[] },
  ]

  conversations.forEach((conversation) => {
    if (conversation.updatedAt >= todayStart) buckets[0].items.push(conversation)
    else if (conversation.updatedAt >= yesterdayStart) buckets[1].items.push(conversation)
    else if (conversation.updatedAt >= weekStart) buckets[2].items.push(conversation)
    else buckets[3].items.push(conversation)
  })

  return buckets.filter((bucket) => bucket.items.length > 0)
}
