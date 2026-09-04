import axios from 'axios'

import { wafHost } from 'shared/api/hosts'

export interface WAFWarning {
  list: string
  line?: number
  code: string
}

export interface WAFSnapshot {
  whitelist: string
  blacklist: string
  referers: string
  ip_enabled: boolean
  referer_enabled: boolean
  read_only: boolean
  warnings: WAFWarning[]
}

export interface WAFLists {
  whitelist: string
  blacklist: string
  referers: string
}

const emptySnapshot = (): WAFSnapshot => ({
  whitelist: '',
  blacklist: '',
  referers: '',
  ip_enabled: false,
  referer_enabled: false,
  read_only: false,
  warnings: [],
})

export const getWAF = async (signal?: AbortSignal): Promise<WAFSnapshot> => {
  const { data } = await axios.get<WAFSnapshot>(wafHost(), { signal })
  return {
    ...emptySnapshot(),
    ...data,
    whitelist: data.whitelist || '',
    blacklist: data.blacklist || '',
    referers: data.referers || '',
    warnings: data.warnings || [],
  }
}

/** POST /waf requires all three list fields; empty strings clear a list. */
export const setWAF = async (lists: WAFLists): Promise<WAFSnapshot> => {
  const { data } = await axios.post<WAFSnapshot>(wafHost(), {
    whitelist: lists.whitelist,
    blacklist: lists.blacklist,
    referers: lists.referers,
  })
  return {
    ...emptySnapshot(),
    ...data,
    whitelist: data.whitelist || '',
    blacklist: data.blacklist || '',
    referers: data.referers || '',
    warnings: data.warnings || [],
  }
}
