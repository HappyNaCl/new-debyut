import { writable } from 'svelte/store'

export interface TokenState {
  accessToken: string | null
}

export const tokens = writable<TokenState>({ accessToken: null })

export const authReady = writable<boolean>(false)

export function setAccessToken(accessToken: string) {
  tokens.set({ accessToken })
}

export function clearTokens() {
  tokens.set({ accessToken: null })
}
