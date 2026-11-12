<script lang="ts">
  import type { ApplicationStatus } from "../../types/organizerApplication";

  interface Props {
    status: ApplicationStatus;
  }

  let { status }: Props = $props();

  const labels: Record<ApplicationStatus, string> = {
    PENDING: "Pending review",
    APPROVED: "Approved",
    DECLINED: "Declined",
  };
</script>

<span class="status-badge {status.toLowerCase()}">{labels[status]}</span>

<style lang="sass">
  .status-badge
    display: inline-flex
    align-items: center
    flex-shrink: 0
    height: 24px
    padding: 0 10px
    border-radius: 999px
    font-size: 12px
    font-weight: 700
    white-space: nowrap

  .pending
    background: var(--color-warning-soft)
    color: var(--color-warning)

  .approved
    background: var(--color-info-soft)
    color: var(--color-info)

  .declined
    background: var(--color-error-soft)
    color: var(--color-error)
</style>
