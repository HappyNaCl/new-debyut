<script lang="ts">
  interface Props {
    checked: boolean;
    label: string;
    disabled?: boolean;
    onChange: (checked: boolean) => void;
  }

  let { checked, label, disabled = false, onChange }: Props = $props();
</script>

<button
  type="button"
  role="switch"
  class="toggle-switch"
  aria-checked={checked}
  aria-label={label}
  {disabled}
  onclick={() => onChange(!checked)}
>
  <span class="thumb"></span>
</button>

<style lang="sass">
  .toggle-switch
    position: relative
    flex-shrink: 0
    width: 44px
    height: 24px
    padding: 0
    border: none
    border-radius: 999px
    background: var(--color-border)
    cursor: pointer
    transition: background 0.15s

    &[aria-checked="true"]
      background: var(--color-primary)

      .thumb
        transform: translateX(20px)

    &:disabled
      opacity: 0.55
      cursor: not-allowed

    &:focus-visible
      outline: none
      box-shadow: var(--focus-ring)

  .thumb
    position: absolute
    top: 3px
    left: 3px
    width: 18px
    height: 18px
    border-radius: 999px
    background: var(--color-switch-thumb)
    box-shadow: var(--card-shadow)
    transition: transform 0.15s
</style>
