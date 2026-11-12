export function clickOutside(node: HTMLElement, onOutside: () => void) {
  let callback = onOutside;

  function handlePointerDown(event: PointerEvent) {
    if (!node.contains(event.target as Node)) callback();
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Escape") callback();
  }

  document.addEventListener("pointerdown", handlePointerDown);
  document.addEventListener("keydown", handleKeydown);

  return {
    update(nextOnOutside: () => void) {
      callback = nextOnOutside;
    },
    destroy() {
      document.removeEventListener("pointerdown", handlePointerDown);
      document.removeEventListener("keydown", handleKeydown);
    },
  };
}
