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
  results?: Array<{
    chunk_id: string
    graph_node_id: string
    content: string
    title: string
    source_link: string
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
