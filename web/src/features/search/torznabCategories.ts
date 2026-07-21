import type { TorznabCapsCategory } from 'shared/api/types'

export type TorznabCategoryOption = {
  id: string
  name: string
  indent?: number
}

/** Common Newznab category tree used when indexer caps are unavailable. */
export const STATIC_TORZNAB_CATEGORY_TREE: TorznabCapsCategory[] = [
  {
    id: '2000',
    name: 'Movies',
    subcategories: [
      { id: '2030', name: 'Movies/Foreign' },
      { id: '2040', name: 'Movies/Other' },
      { id: '2045', name: 'Movies/UHD' },
      { id: '2050', name: 'Movies/HD' },
      { id: '2060', name: 'Movies/3D' },
    ],
  },
  {
    id: '5000',
    name: 'TV',
    subcategories: [
      { id: '5030', name: 'TV/Foreign' },
      { id: '5040', name: 'TV/HD' },
      { id: '5045', name: 'TV/UHD' },
      { id: '5050', name: 'TV/Other' },
      { id: '5070', name: 'TV/Anime' },
    ],
  },
  {
    id: '3000',
    name: 'Audio',
    subcategories: [
      { id: '3010', name: 'Audio/MP3' },
      { id: '3040', name: 'Audio/Lossless' },
      { id: '3050', name: 'Audio/Other' },
      { id: '3060', name: 'Audio/Video' },
    ],
  },
  {
    id: '4000',
    name: 'PC',
    subcategories: [
      { id: '4010', name: 'PC/0day' },
      { id: '4020', name: 'PC/ISO' },
      { id: '4030', name: 'PC/Mac' },
      { id: '4040', name: 'PC/Mobile-Other' },
      { id: '4070', name: 'PC/Games' },
    ],
  },
  {
    id: '7000',
    name: 'Books',
    subcategories: [
      { id: '7020', name: 'Books/EBook' },
      { id: '7030', name: 'Books/Comics' },
      { id: '7050', name: 'Books/Mags' },
    ],
  },
  {
    id: '6000',
    name: 'XXX',
    subcategories: [
      { id: '6010', name: 'XXX/DVD' },
      { id: '6020', name: 'XXX/Pack' },
      { id: '6030', name: 'XXX/Other' },
      { id: '6040', name: 'XXX/x264' },
      { id: '6050', name: 'XXX/Picture' },
    ],
  },
]

export const flattenTorznabCategories = (
  categories: TorznabCapsCategory[],
  allLabel: string,
): TorznabCategoryOption[] => {
  const options: TorznabCategoryOption[] = [{ id: '', name: allLabel }]

  const walk = (cats: TorznabCapsCategory[], indent: number) => {
    for (const cat of cats) {
      options.push({ id: cat.id, name: cat.name, indent })
      if (cat.subcategories?.length) walk(cat.subcategories, indent + 1)
    }
  }

  walk(categories, 0)
  return options
}

export const staticTorznabCategoryOptions = (allLabel: string): TorznabCategoryOption[] =>
  flattenTorznabCategories(STATIC_TORZNAB_CATEGORY_TREE, allLabel)
