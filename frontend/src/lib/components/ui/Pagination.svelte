<script lang="ts">
  import { ChevronLeft, ChevronRight } from "@material-symbols-svg/svelte";

  interface Props {
    page: number;
    pageSize: number;
    total: number;
    onChange: (page: number) => void;
  }

  let { page, pageSize, total, onChange }: Props = $props();

  let pageCount = $derived(Math.max(1, Math.ceil(total / pageSize)));
</script>

{#if pageCount > 1}
  <nav class="pagination" aria-label="Pagination">
    <button type="button" aria-label="Previous page" disabled={page <= 1} onclick={() => onChange(page - 1)}>
      <ChevronLeft size={20} />
    </button>
    <span>Page {page} of {pageCount}</span>
    <button type="button" aria-label="Next page" disabled={page >= pageCount} onclick={() => onChange(page + 1)}>
      <ChevronRight size={20} />
    </button>
  </nav>
{/if}

<style lang="sass">
  .pagination
    display: flex
    align-items: center
    justify-content: center
    gap: 12px
    font-size: 14px
    font-weight: 600
    color: var(--color-text-muted)

  button
    display: inline-flex
    align-items: center
    justify-content: center
    width: 36px
    height: 36px
    border: 1px solid var(--color-border)
    border-radius: 8px
    background: var(--color-surface)
    color: var(--color-text)
    cursor: pointer

    &:hover:not(:disabled)
      border-color: var(--color-primary)
      color: var(--color-primary)

    &:disabled
      opacity: 0.45
      cursor: not-allowed
</style>
