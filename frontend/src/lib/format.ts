// Helpers de formatação (valores em centavos).

const brl = new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' })

/** centavos -> "R$ 12,90" */
export function money(cents: number): string {
  return brl.format((cents || 0) / 100)
}

/** centavos -> entrada de formulário em reais (ex.: "12.90") sem máscara de moeda. */
export function centsToInput(cents: number): string {
  const v = (cents || 0) / 100
  return v.toFixed(2).replace('.', ',')
}

/** "12,90" ou "12.90" (usuário) -> centavos. */
export function inputToCents(text: string): number {
  const t = String(text ?? '').trim().replace(/\./g, '').replace(',', '.')
  if (!t) return 0
  const n = parseFloat(t)
  if (isNaN(n) || n < 0) return 0
  return Math.round(n * 100)
}

/**
 * Exibe uma data/hora armazenada por SQLite datetime('now','localtime')
 * (ex.: "2026-09-05 14:30:00") no formato pt-BR. O valor já é hora local e
 * não carrega fuso; não tratamos como UTC.
 */
export function dtBR(dt: string): string {
  if (!dt) return ''
  // pega só a parte de data "2026-09-05"
  const [ymd, hms] = dt.split(' ')
  const [y, m, d] = (ymd || '').split('-')
  if (!y || !m || !d) return dt
  const datePart = `${d}/${m}/${y}`
  const [, hh, mm] = (hms || '').split(':')
  if (hh && mm) return `${datePart} ${hh}:${mm}`
  return datePart
}
