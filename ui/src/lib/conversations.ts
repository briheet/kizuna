import type { Conversation } from '../types'

const CONVERSATIONS_KEY = 'kizuna:conversations:v1'

export function loadConversations(): Conversation[] {
  try {
    const value = localStorage.getItem(CONVERSATIONS_KEY)
    if (!value) return []
    const conversations = JSON.parse(value) as Conversation[]
    if (!Array.isArray(conversations)) return []

    // Remove only threads created by the retired mock evidence mode.
    const cleaned = conversations.filter((conversation) => !conversation.messages.some((message) =>
      message.evidence?.some((item) => item.id.startsWith('demo-')),
    ))
    if (cleaned.length !== conversations.length) saveConversations(cleaned)
    return cleaned
  } catch {
    return []
  }
}

export function saveConversations(conversations: Conversation[]) {
  localStorage.setItem(CONVERSATIONS_KEY, JSON.stringify(conversations.slice(0, 50)))
}

export function makeConversation(): Conversation {
  const now = Date.now()
  return {
    id: crypto.randomUUID(),
    title: 'New conversation',
    createdAt: now,
    updatedAt: now,
    messages: [],
  }
}

export function titleFromPrompt(prompt: string) {
  const clean = prompt.replace(/\s+/g, ' ').trim()
  if (clean.length <= 42) return clean
  return `${clean.slice(0, 42).trim()}...`
}
