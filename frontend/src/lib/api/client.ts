import axios, { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { get } from "svelte/store";
import { env } from "../env";
import type { ApiResponse } from "../types/api";
import {
  tokens,
  authReady,
  setAccessToken,
  clearTokens,
} from "../stores/tokens";

export const api = axios.create({
  baseURL: env.VITE_API_URL,
  headers: { "Content-Type": "application/json" },
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const accessToken = get(tokens).accessToken;
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`;
  }
  return config;
});

type RetriableConfig = InternalAxiosRequestConfig & { _retry?: boolean };

let refreshPromise: Promise<string> | null = null;

export async function refreshAccessToken(): Promise<string> {
  const res = await axios.post<ApiResponse<{ accessToken: string }>>(
    `${env.VITE_API_URL}/auth/refresh`,
    null,
    { withCredentials: true },
  );

  setAccessToken(res.data.data.accessToken);
  return res.data.data.accessToken;
}

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const original = error.config as RetriableConfig | undefined;
    if (
      error.response?.status === 401 &&
      original &&
      !original._retry &&
      original.headers.Authorization
    ) {
      original._retry = true;
      try {
        refreshPromise ??= refreshAccessToken().finally(() => {
          refreshPromise = null;
        });
        const accessToken = await refreshPromise;
        original.headers.Authorization = `Bearer ${accessToken}`;
        return api(original);
      } catch (refreshError) {
        clearTokens();
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  },
);

export async function bootstrapAuth(): Promise<void> {
  try {
    await refreshAccessToken();
  } catch {
  } finally {
    authReady.set(true);
  }
}

export function extractErrorMessage(err: unknown): string {
  if (axios.isAxiosError(err)) {
    const data = err.response?.data as { error?: unknown } | undefined;
    const error = data?.error;
    if (typeof error === "string") return error;
    if (error && typeof error === "object")
      return Object.values(error).join(" ");
  }
  return "Something went wrong. Please try again.";
}

export function extractFieldErrors(err: unknown): Record<string, string> {
  if (axios.isAxiosError(err)) {
    const error = (err.response?.data as { error?: unknown } | undefined)?.error;
    if (error && typeof error === "object") return error as Record<string, string>;
  }
  return {};
}
