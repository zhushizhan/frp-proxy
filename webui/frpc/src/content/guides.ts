import { translate, translateList } from '../i18n'
import type { ProxyType, VisitorType } from '../types'

export interface TypeGuide<TType extends string> {
  type: TType
  label: string
  summary: string
  scenarios: string[]
  caution?: string
  primaryFields: string[]
}

export const PROXY_TYPE_ORDER: ProxyType[] = [
  'tcp',
  'udp',
  'http',
  'https',
  'tcpmux',
  'stcp',
  'sudp',
  'xtcp',
] as const

export const VISITOR_TYPE_ORDER: VisitorType[] = ['stcp', 'sudp', 'xtcp'] as const

const proxyFieldKeyMap: Record<ProxyType, string[]> = {
  tcp: ['common.name', 'form.localIP', 'form.localPort', 'form.remotePort'],
  udp: ['common.name', 'form.localIP', 'form.localPort', 'form.remotePort'],
  http: ['common.name', 'form.localPort', 'form.customDomains', 'form.subdomain'],
  https: ['common.name', 'form.localPort', 'form.customDomains', 'form.subdomain'],
  tcpmux: ['common.name', 'form.localPort', 'form.customDomains', 'form.multiplexer'],
  stcp: ['common.name', 'form.localPort', 'form.secretKey', 'form.allowUsers'],
  sudp: ['common.name', 'form.localPort', 'form.secretKey', 'form.allowUsers'],
  xtcp: ['common.name', 'form.localPort', 'form.secretKey', 'form.allowUsers'],
}

const visitorFieldKeyMap: Record<VisitorType, string[]> = {
  stcp: ['common.name', 'form.serverName', 'form.secretKey', 'form.bindPort'],
  sudp: ['common.name', 'form.serverName', 'form.secretKey', 'form.bindPort'],
  xtcp: ['common.name', 'form.serverName', 'form.secretKey', 'form.bindPort'],
}

export const getProxyGuides = (): TypeGuide<ProxyType>[] =>
  PROXY_TYPE_ORDER.map((type) => ({
    type,
    label: translate(`guides.proxy.${type}.label`),
    summary: translate(`guides.proxy.${type}.summary`),
    scenarios: translateList(`guides.proxy.${type}.scenarios`),
    caution: translate(`guides.proxy.${type}.caution`),
    primaryFields: proxyFieldKeyMap[type].map((key) => translate(key)),
  }))

export const getVisitorGuides = (): TypeGuide<VisitorType>[] =>
  VISITOR_TYPE_ORDER.map((type) => ({
    type,
    label: translate(`guides.visitor.${type}.label`),
    summary: translate(`guides.visitor.${type}.summary`),
    scenarios: translateList(`guides.visitor.${type}.scenarios`),
    caution: translate(`guides.visitor.${type}.caution`),
    primaryFields: visitorFieldKeyMap[type].map((key) => translate(key)),
  }))

export const getProxyGuideMap = (): Record<ProxyType, TypeGuide<ProxyType>> =>
  Object.fromEntries(getProxyGuides().map((guide) => [guide.type, guide])) as Record<
    ProxyType,
    TypeGuide<ProxyType>
  >

export const getVisitorGuideMap = (): Record<VisitorType, TypeGuide<VisitorType>> =>
  Object.fromEntries(getVisitorGuides().map((guide) => [guide.type, guide])) as Record<
    VisitorType,
    TypeGuide<VisitorType>
  >
