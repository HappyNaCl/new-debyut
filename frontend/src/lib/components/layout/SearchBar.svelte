<script lang="ts">
  import { onMount } from "svelte";
  import { Search } from "@material-symbols-svg/svelte";
  import { push, router } from "svelte-spa-router";

  const suggestions = ["concerts", "music festivals", "stand-up comedy", "workshops", "seminars", "sports matches"];
  const ROTATE_INTERVAL_MS = 3000;

  let query = $state("");
  let suggestionIndex = $state(0);

  $effect(() => {
    query = new URLSearchParams(router.querystring ?? "").get("q") ?? "";
  });

  onMount(() => {
    const timer = setInterval(() => {
      suggestionIndex = (suggestionIndex + 1) % suggestions.length;
    }, ROTATE_INTERVAL_MS);
    return () => clearInterval(timer);
  });

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    const trimmed = query.trim() || suggestions[suggestionIndex];
    push(`/search?q=${encodeURIComponent(trimmed)}`);
  }
</script>

<form class="search-bar" role="search" onsubmit={handleSubmit}>
  <input type="search" aria-label="Search events" bind:value={query} />
  {#if !query}
    <span class="placeholder" aria-hidden="true">
      Search for <strong>{suggestions[suggestionIndex]}</strong>
    </span>
  {/if}
  <button type="submit" aria-label="Search">
    <Search size={20} />
  </button>
</form>

<style lang="sass">
  .search-bar
    position: relative
    display: flex
    align-items: center
    flex: 1
    min-width: 0
    height: 44px
    border: 1px solid transparent
    border-radius: 8px
    background: var(--color-surface-muted)
    transition: border-color 0.15s, box-shadow 0.15s

    &:focus-within
      border-color: var(--color-primary)
      box-shadow: var(--focus-ring)

  input
    flex: 1
    min-width: 0
    height: 100%
    padding: 0 16px
    border: none
    background: none
    font-size: 14px
    color: var(--color-text)
    outline: none

    &::-webkit-search-cancel-button
      display: none

  .placeholder
    position: absolute
    left: 17px
    right: 48px
    overflow: hidden
    font-size: 14px
    white-space: nowrap
    text-overflow: ellipsis
    color: var(--color-text-subtle)
    pointer-events: none

    strong
      color: var(--color-strong)

  button
    display: inline-flex
    align-items: center
    justify-content: center
    width: 44px
    height: 100%
    padding: 0
    border: none
    background: none
    color: var(--color-text-muted)
    cursor: pointer

    &:hover
      color: var(--color-primary)
</style>
