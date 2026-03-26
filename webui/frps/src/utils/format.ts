import { getCurrentLocale, translate } from '../i18n'

export function formatDistanceToNow(date: Date): string {
  const seconds = Math.floor((new Date().getTime() - date.getTime()) / 1000)
  if (seconds <= 0) return translate('time.justNow')

  let interval = seconds / 31536000
  if (interval >= 1) return formatRelativeUnit(Math.floor(interval), 'year', 'years')

  interval = seconds / 2592000
  if (interval >= 1) return formatRelativeUnit(Math.floor(interval), 'month', 'months')

  interval = seconds / 86400
  if (interval >= 1) return formatRelativeUnit(Math.floor(interval), 'day', 'days')

  interval = seconds / 3600
  if (interval >= 1) return formatRelativeUnit(Math.floor(interval), 'hour', 'hours')

  interval = seconds / 60
  if (interval >= 1) return formatRelativeUnit(Math.floor(interval), 'minute', 'minutes')

  return formatRelativeUnit(Math.floor(seconds), 'second', 'seconds')
}

export function formatFileSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return '0 B'
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  // Prevent index out of bounds for extremely large numbers
  const unit = sizes[i] || sizes[sizes.length - 1]
  const val = bytes / Math.pow(k, i)

  return parseFloat(val.toFixed(2)) + ' ' + unit
}

function formatRelativeUnit(
  count: number,
  singularKey: 'year' | 'month' | 'day' | 'hour' | 'minute' | 'second',
  pluralKey: 'years' | 'months' | 'days' | 'hours' | 'minutes' | 'seconds',
): string {
  const locale = getCurrentLocale()
  if (locale === 'zh-CN') {
    return translate(`time.${pluralKey}`, { count })
  }
  return translate(`time.${count === 1 ? singularKey : pluralKey}`, { count })
}
