/** Resolve a public asset path against Vite `BASE_URL` (supports `./` and subpath deploys). */
export function publicUrl(path: string): string {
  const rel = path.replace(/^\//, '')
  const base = import.meta.env.BASE_URL ?? './'
  if (base === './' || base === '.' || base === '') return `./${rel}`
  const prefix = base.endsWith('/') ? base : `${base}/`
  return `${prefix}${rel}`
}

/** App pathname prefix without trailing slash (empty when served at site root / relative base). */
export function getAppBasePath(): string {
  const base = import.meta.env.BASE_URL ?? './'
  if (base === './' || base === '.' || base === '/' || base === '') return ''
  return base.replace(/\/$/, '')
}
