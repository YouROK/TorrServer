import ptt from 'parse-torrent-title'

/** Keep files for the selected season. Paths without a season stay visible. */
export function filterFilesForSeason<T extends { path: string }>(
  files: T[] | undefined,
  selectedSeason: number | undefined,
  seasonAmount: number[] | null | undefined,
): T[] {
  if (!files?.length) return []
  return files.filter(({ path }) => {
    const { season } = ptt.parse(path)
    return season == null || season === selectedSeason || !seasonAmount?.length
  })
}
