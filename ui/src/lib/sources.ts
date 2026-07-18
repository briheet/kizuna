import type { KnowledgeSource } from '../types'

const SOURCES_KEY = 'kizuna:knowledge-sources:v1'

export function loadSources(): KnowledgeSource[] {
  try {
    const parsed = JSON.parse(localStorage.getItem(SOURCES_KEY) ?? '[]')
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function saveSources(sources: KnowledgeSource[]) {
  localStorage.setItem(SOURCES_KEY, JSON.stringify(sources))
}
