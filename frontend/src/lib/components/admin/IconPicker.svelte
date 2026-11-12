<script lang="ts">
  import { Search } from "@material-symbols-svg/svelte";
  import { categoryIconOptions } from "../../icons/categoryIcons";

  interface Props {
    value: string;
    invalid?: boolean;
    onSelect: (key: string) => void;
  }

  let { value, invalid = false, onSelect }: Props = $props();

  let search = $state("");

  let filteredOptions = $derived.by(() => {
    const term = search.trim().toLowerCase();
    if (!term) return categoryIconOptions;
    return categoryIconOptions.filter(
      (option) => option.label.toLowerCase().includes(term) || option.key.includes(term),
    );
  });
</script>

<div class="icon-picker" class:invalid>
  <label class="search">
    <Search size={18} />
    <input type="text" placeholder="Search icons" aria-label="Search icons" bind:value={search} />
  </label>

  {#if filteredOptions.length === 0}
    <p class="empty">No icons match “{search}”.</p>
  {:else}
    <div class="grid" role="radiogroup" aria-label="Category icon">
      {#each filteredOptions as option (option.key)}
        <button
          type="button"
          class="option"
          class:selected={option.key === value}
          role="radio"
          aria-checked={option.key === value}
          aria-label={option.label}
          title={option.label}
          onclick={() => onSelect(option.key)}
        >
          <option.icon size={22} />
        </button>
      {/each}
    </div>
  {/if}
</div>

<style lang="sass">
  .icon-picker
    display: flex
    flex-direction: column
    gap: 10px
    padding: 12px
    border: 1px solid var(--color-border)
    border-radius: 8px
    background: var(--color-surface)

    &.invalid
      border-color: var(--color-error)

  .search
    display: flex
    align-items: center
    gap: 8px
    height: 38px
    padding: 0 12px
    border-radius: 8px
    background: var(--color-surface-muted)
    color: var(--color-text-subtle)

    &:focus-within
      box-shadow: var(--focus-ring)

    input
      flex: 1
      min-width: 0
      border: none
      background: none
      font-size: 14px
      color: var(--color-text)
      outline: none

  .grid
    display: grid
    grid-template-columns: repeat(auto-fill, minmax(44px, 1fr))
    gap: 6px
    max-height: 232px
    overflow-y: auto

  .option
    display: inline-flex
    align-items: center
    justify-content: center
    aspect-ratio: 1
    border: 1px solid transparent
    border-radius: 8px
    background: none
    color: var(--color-text-muted)
    cursor: pointer
    transition: background 0.15s, color 0.15s, border-color 0.15s

    &:hover
      background: var(--color-surface-muted)
      color: var(--color-text)

    &:focus-visible
      outline: none
      box-shadow: var(--focus-ring)

    &.selected
      border-color: var(--color-primary)
      background: var(--color-primary-soft)
      color: var(--color-primary)

  .empty
    margin: 0
    padding: 16px 0
    text-align: center
    font-size: 14px
    color: var(--color-text-muted)
</style>
