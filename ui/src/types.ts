export type Evidence = {
  id: string
  title: string
  content: string
  sourceLink: string
  sourceType: string
  relevance: number
}

export type ChatMessage = {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: number
  evidence?: Evidence[]
  status?: 'loading' | 'error'
}

export type Conversation = {
  id: string
  title: string
  createdAt: number
  updatedAt: number
  messages: ChatMessage[]
}

export type SearchResponse = {
  summary: string
  results?: Array<{
    chunk_id: string
    graph_node_id: string
    content: string
    title: string
    source_link: string
    source_type: string
    node_type: string
    distance: number
  }>
  related_nodes?: Array<{
    id: string
    node_type: string
    title: string
    source_link: string
  }>
}

export type ConnectionStatus = 'checking' | 'online' | 'offline'

export type IngestionState = 'queued' | 'syncing' | 'ready' | 'failed'

export type JobStateCount = {
  state: string
  count: number
}

export type KnowledgeSource = {
  id: string
  topicId: string
  sourceType: 'github'
  name: string
  sourceLink: string
  scopes: string[]
  state: IngestionState
  counts: JobStateCount[]
  lastError?: string
  createdAt: number
  updatedAt: number
}

export type JobsStatusResponse = {
  counts?: JobStateCount[]
  recent_failures?: Array<{
    id: string
    state: string
    kind: string
    last_error: string
    recent_at: string
  }>
}
