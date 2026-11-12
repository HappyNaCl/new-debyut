import { writable, get } from "svelte/store";
import axios from "axios";
import { getMe } from "../api/user";
import { refreshAccessToken } from "../api/client";
import type { Me } from "../types/user";

export const user = writable<Me | null>(null);

const RETRY_DELAY_MS = 400;

let loadPromise: Promise<Me> | null = null;

async function fetchMe(attempts: number): Promise<Me> {
  for (let attempt = 1; ; attempt++) {
    try {
      const response = await getMe();
      return response.data.data;
    } catch (error) {
      const notFound = axios.isAxiosError(error) && error.response?.status === 404;
      if (!notFound || attempt >= attempts) throw error;
      await new Promise((resolve) => setTimeout(resolve, RETRY_DELAY_MS));
    }
  }
}

export async function loadUser(force = false, attempts = 1): Promise<Me> {
  const cached = get(user);
  if (cached && !force) return cached;

  loadPromise ??= fetchMe(attempts)
    .then((me) => {
      user.set(me);
      return me;
    })
    .finally(() => {
      loadPromise = null;
    });

  return loadPromise;
}

export async function refreshSession(): Promise<Me | null> {
  try {
    await refreshAccessToken();
    return await loadUser(true);
  } catch {
    return null;
  }
}

export function clearUser() {
  user.set(null);
}
