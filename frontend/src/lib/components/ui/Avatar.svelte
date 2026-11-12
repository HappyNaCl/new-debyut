<script lang="ts">
  import { initials } from "../../format";

  interface Props {
    name: string;
    src?: string;
    size?: number;
    shape?: "circle" | "rounded";
  }

  let { name, src = "", size = 40, shape = "circle" }: Props = $props();

  let failedSrc = $state("");
  let showImage = $derived(!!src && failedSrc !== src);
</script>

<span
  class="avatar"
  class:rounded={shape === "rounded"}
  style:width={`${size}px`}
  style:height={`${size}px`}
  style:font-size={`${Math.round(size * 0.38)}px`}
  aria-hidden="true"
>
  {#if showImage}
    <img {src} alt="" onerror={() => (failedSrc = src)} />
  {:else}
    {initials(name) || "?"}
  {/if}
</span>

<style lang="sass">
  .avatar
    display: inline-flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    overflow: hidden
    border-radius: 999px
    background: var(--color-primary)
    font-weight: 800
    line-height: 1
    color: var(--color-on-primary)

    &.rounded
      border-radius: 25%

  img
    width: 100%
    height: 100%
    object-fit: cover
</style>
