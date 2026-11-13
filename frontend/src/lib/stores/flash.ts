import { writable } from "svelte/store";

/**
 * A one-off message carried across a navigation (e.g. "Event published"),
 * shown and cleared by the page that is navigated to.
 */
export const flashMessage = writable<string>("");

export function takeFlash(): string {
  let message = "";
  flashMessage.update((current) => {
    message = current;
    return "";
  });
  return message;
}
