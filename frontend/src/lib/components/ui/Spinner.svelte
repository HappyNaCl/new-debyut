<script lang="ts">
  import { ProgressActivity } from "@material-symbols-svg/svelte";

  interface Props {
    size?: number;
    fullscreen?: boolean;
  }

  let { size = 40, fullscreen = false }: Props = $props();
</script>

<div class="spinner" class:fullscreen role="status" aria-label="Loading">
  <ProgressActivity {size} />
</div>

<style lang="sass">
  .spinner
    display: inline-flex
    align-items: center
    justify-content: center

  .fullscreen
    position: fixed
    inset: 0
    width: 100vw
    height: 100vh

  .spinner :global(svg)
    color: var(--color-primary)
    animation: spin 0.9s linear infinite

  @keyframes spin
    to
      transform: rotate(360deg)
</style>
