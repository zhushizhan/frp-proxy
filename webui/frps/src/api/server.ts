import { http } from './http'
import type { ServerInfo, ServerSettings } from '../types/server'

export const getServerInfo = () => {
  return http.get<ServerInfo>('../api/serverinfo')
}

export const getServerSettings = () => {
  return http.get<ServerSettings>('../api/settings')
}

export const updateServerSettings = (settings: ServerSettings) => {
  return http.put<void>('../api/settings', settings)
}
