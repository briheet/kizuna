const configuredApiBase = import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, '')

export const apiBase = configuredApiBase || `${window.location.protocol}//${window.location.hostname}:4000`

export function apiUrl(path: string) {
  return `${apiBase}${path}`
}
