import { writable } from "svelte/store";

export type Theme = "light" | "dark";

const STORAGE_KEY = "theme";
const darkQuery = window.matchMedia("(prefers-color-scheme: dark)");

function storedTheme(): Theme | null {
  try {
    const value = localStorage.getItem(STORAGE_KEY);
    return value === "light" || value === "dark" ? value : null;
  } catch {
    return null;
  }
}

function systemTheme(): Theme {
  return darkQuery.matches ? "dark" : "light";
}

export const theme = writable<Theme>(storedTheme() ?? systemTheme());

theme.subscribe((value) => {
  document.documentElement.dataset.theme = value;
});

darkQuery.addEventListener("change", () => {
  if (!storedTheme()) theme.set(systemTheme());
});

export function toggleTheme() {
  theme.update((current) => {
    const next: Theme = current === "dark" ? "light" : "dark";
    try {
      localStorage.setItem(STORAGE_KEY, next);
    } catch {}
    return next;
  });
}
