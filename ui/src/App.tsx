import { useCallback, useEffect, useMemo, useState } from 'react'
import { ChatView } from './components/ChatView'
import { EvidencePanel } from './components/EvidencePanel'
import { Sidebar } from './components/Sidebar'
import { loadConversations, makeConversation, saveConversations, titleFromPrompt } from './lib/conversations'
import { checkApiHealth, composeAnswer, searchKnowledge } from './lib/search'
import type { ChatMessage, ConnectionStatus, Conversation, Evidence } from './types'

const SIDEBAR_COLLAPSED_KEY = 'kizuna:sidebar-collapsed:v1'

function App() {
  const [conversations, setConversations] = useState<Conversation[]>(() => {
    const saved = loadConversations()
    return saved.length ? saved : [makeConversation()]
  })
  const [activeId, setActiveId] = useState(() => conversations[0].id)
  const [connectionStatus, setConnectionStatus] = useState<ConnectionStatus>('checking')
  const [evidence, setEvidence] = useState<Evidence[]>([])
  const [evidenceOpen, setEvidenceOpen] = useState(false)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [sidebarCollapsed, setSidebarCollapsed] = useState(() => localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true')

  const activeConversation = useMemo(
    () => conversations.find((conversation) => conversation.id === activeId) ?? conversations[0],
    [activeId, conversations],
  )

  useEffect(() => saveConversations(conversations), [conversations])
  useEffect(() => localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(sidebarCollapsed)), [sidebarCollapsed])

  useEffect(() => {
    const controller = new AbortController()
    let active = true
    const timeout = window.setTimeout(() => {
      if (active) setConnectionStatus('offline')
      controller.abort()
    }, 3500)
    setConnectionStatus('checking')
    checkApiHealth(controller.signal)
      .then((online) => {
        if (active) setConnectionStatus(online ? 'online' : 'offline')
      })
      .catch(() => {
        if (active) setConnectionStatus('offline')
      })
      .finally(() => window.clearTimeout(timeout))

    return () => {
      active = false
      window.clearTimeout(timeout)
      controller.abort()
    }
  }, [])

  useEffect(() => {
    const latestAssistant = [...activeConversation.messages].reverse().find((message) => message.role === 'assistant')
    const recentEvidence = latestAssistant?.evidence ?? []
    setEvidence(recentEvidence)
    setEvidenceOpen(false)
  }, [activeConversation.id, activeConversation.messages])

  const createNewConversation = useCallback(() => {
    const current = conversations.find((conversation) => conversation.id === activeId)
    if (current && current.messages.length === 0) {
      setMobileMenuOpen(false)
      return
    }
    const next = makeConversation()
    setConversations((items) => [next, ...items])
    setActiveId(next.id)
    setEvidence([])
    setMobileMenuOpen(false)
  }, [activeId, conversations])

  useEffect(() => {
    function shortcut(event: globalThis.KeyboardEvent) {
      if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
        event.preventDefault()
        createNewConversation()
      }
    }
    window.addEventListener('keydown', shortcut)
    return () => window.removeEventListener('keydown', shortcut)
  }, [createNewConversation])

  function deleteConversation(id: string) {
    setConversations((items) => {
      const remaining = items.filter((conversation) => conversation.id !== id)
      if (remaining.length === 0) {
        const next = makeConversation()
        setActiveId(next.id)
        return [next]
      }
      if (id === activeId) setActiveId(remaining[0].id)
      return remaining
    })
  }

  async function sendMessage(prompt: string) {
    const conversationId = activeConversation.id
    const now = Date.now()
    const userMessage: ChatMessage = { id: crypto.randomUUID(), role: 'user', content: prompt, createdAt: now }
    const loadingId = crypto.randomUUID()
    const loadingMessage: ChatMessage = { id: loadingId, role: 'assistant', content: '', createdAt: now + 1, status: 'loading' }

    setConversations((items) => items.map((conversation) => {
      if (conversation.id !== conversationId) return conversation
      return {
        ...conversation,
        title: conversation.messages.length === 0 ? titleFromPrompt(prompt) : conversation.title,
        updatedAt: now,
        messages: [...conversation.messages, userMessage, loadingMessage],
      }
    }))

    await retrieveAnswer(conversationId, prompt, loadingId)
  }

  async function regenerateAnswer(messageId: string) {
    const messageIndex = activeConversation.messages.findIndex((message) => message.id === messageId)
    if (messageIndex < 0) return

    const userMessage = activeConversation.messages
      .slice(0, messageIndex)
      .reverse()
      .find((message) => message.role === 'user')
    if (!userMessage) return

    replaceMessage(activeConversation.id, messageId, {
      id: messageId,
      role: 'assistant',
      content: '',
      createdAt: Date.now(),
      status: 'loading',
    })
    setEvidence([])
    await retrieveAnswer(activeConversation.id, userMessage.content, messageId)
  }

  async function retrieveAnswer(conversationId: string, prompt: string, messageId: string) {
    try {
      const results = await searchKnowledge(prompt)
      setConnectionStatus('online')
      setEvidence(results)
      const answer: ChatMessage = {
        id: messageId,
        role: 'assistant',
        content: composeAnswer(results),
        createdAt: Date.now(),
        evidence: results,
      }
      replaceMessage(conversationId, messageId, answer)
    } catch (error) {
      const detail = error instanceof Error ? error.message : 'Unknown connection error.'
      setConnectionStatus('offline')
      replaceMessage(conversationId, messageId, {
        id: messageId,
        role: 'assistant',
        content: `I couldn't complete the search. ${detail}`,
        createdAt: Date.now(),
        status: 'error',
      })
    }
  }

  function replaceMessage(conversationId: string, messageId: string, next: ChatMessage) {
    setConversations((items) => items.map((conversation) => (
      conversation.id === conversationId
        ? { ...conversation, updatedAt: Date.now(), messages: conversation.messages.map((message) => message.id === messageId ? next : message) }
        : conversation
    )))
  }

  if (!activeConversation) return null

  return (
    <div className={`app-shell ${evidenceOpen ? 'evidence-is-open' : ''} ${sidebarCollapsed ? 'sidebar-is-collapsed' : ''}`}>
      <Sidebar
        activeId={activeConversation.id}
        collapsed={sidebarCollapsed}
        connectionStatus={connectionStatus}
        conversations={conversations}
        mobileOpen={mobileMenuOpen}
        onCloseMobile={() => setMobileMenuOpen(false)}
        onDelete={deleteConversation}
        onNew={createNewConversation}
        onSelect={setActiveId}
        onToggle={() => setSidebarCollapsed((collapsed) => !collapsed)}
      />
      <ChatView
        connectionStatus={connectionStatus}
        conversation={activeConversation}
        evidence={evidence}
        evidenceOpen={evidenceOpen}
        onOpenEvidence={() => setEvidenceOpen((open) => !open)}
        onOpenMenu={() => setMobileMenuOpen(true)}
        onRegenerate={regenerateAnswer}
        onSend={sendMessage}
        onUsePrompt={sendMessage}
      />
      <EvidencePanel evidence={evidence} open={evidenceOpen} onClose={() => setEvidenceOpen(false)} />
    </div>
  )
}

export default App
