<script lang="ts">
  interface Props {
    value: string;
    length?: number;
    disabled?: boolean;
    onComplete?: (code: string) => void;
  }

  let { value = $bindable(""), length = 6, disabled = false, onComplete }: Props = $props();

  let inputs: HTMLInputElement[] = $state([]);

  let digits = $derived(Array.from({ length }, (_, index) => value[index] ?? ""));

  function update(nextValue: string, focusIndex: number) {
    value = nextValue.slice(0, length);
    inputs[Math.min(focusIndex, length - 1)]?.focus();
    if (value.length === length) onComplete?.(value);
  }

  function handleInput(index: number, event: Event) {
    const target = event.target as HTMLInputElement;
    const typed = target.value.replace(/\D/g, "");
    target.value = digits[index];
    if (!typed) return;
    const nextValue = value.slice(0, index) + typed + value.slice(index + typed.length);
    update(nextValue, index + typed.length);
  }

  function handleKeydown(index: number, event: KeyboardEvent) {
    if (event.key === "Backspace") {
      event.preventDefault();
      const removeAt = digits[index] ? index : index - 1;
      if (removeAt < 0) return;
      value = value.slice(0, removeAt);
      inputs[removeAt]?.focus();
    } else if (event.key === "ArrowLeft" && index > 0) {
      inputs[index - 1]?.focus();
    } else if (event.key === "ArrowRight" && index < length - 1) {
      inputs[index + 1]?.focus();
    }
  }

  function handlePaste(event: ClipboardEvent) {
    event.preventDefault();
    const pasted = (event.clipboardData?.getData("text") ?? "").replace(/\D/g, "");
    if (pasted) update(pasted, pasted.length);
  }

  export function focus() {
    inputs[Math.min(value.length, length - 1)]?.focus();
  }
</script>

<div class="otp-input">
  {#each digits as digit, index (index)}
    <input
      bind:this={inputs[index]}
      type="text"
      inputmode="numeric"
      autocomplete={index === 0 ? "one-time-code" : "off"}
      maxlength={length}
      aria-label={`Digit ${index + 1}`}
      value={digit}
      {disabled}
      oninput={(event) => handleInput(index, event)}
      onkeydown={(event) => handleKeydown(index, event)}
      onpaste={handlePaste}
    />
  {/each}
</div>

<style lang="sass">
  .otp-input
    display: flex
    gap: 10px
    justify-content: space-between

  input
    width: 100%
    max-width: 48px
    height: 56px
    border: 1px solid var(--color-border)
    border-radius: 8px
    text-align: center
    font-size: 22px
    font-weight: 600
    color: var(--color-text)
    background: var(--color-surface)
    outline: none
    transition: border-color 0.15s, box-shadow 0.15s

    &:focus
      border-color: var(--color-primary)
      box-shadow: var(--focus-ring)

    &:disabled
      background: var(--color-surface-muted)
</style>
