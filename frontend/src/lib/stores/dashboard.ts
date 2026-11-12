import { writable } from "svelte/store";

const STORAGE_KEY = "dashboard-sidebar-collapsed";

function storedCollapsed(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

export const sidebarCollapsed = writable<boolean>(storedCollapsed());

export function toggleSidebar() {
  sidebarCollapsed.update((current) => {
    const next = !current;
    try {
      localStorage.setItem(STORAGE_KEY, String(next));
    } catch {}
    return next;
  });
}
