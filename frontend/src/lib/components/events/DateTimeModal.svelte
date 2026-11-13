<script lang="ts">
  import { untrack } from "svelte";
  import Modal from "../ui/Modal.svelte";

  interface Props {
    open: boolean;
    startsAt: string | null;
    endsAt: string | null;
    onSave: (startsAt: string | null, endsAt: string | null) => void;
    onClose: () => void;
  }

  let { open, startsAt, endsAt, onSave, onClose }: Props = $props();

  const timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  function split(value: string | null): { date: string; time: string } {
    if (!value) return { date: "", time: "" };
    const date = new Date(value);
    const pad = (part: number) => String(part).padStart(2, "0");
    return {
      date: `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`,
      time: `${pad(date.getHours())}:${pad(date.getMinutes())}`,
    };
  }

  function join(date: string, time: string): string | null {
    if (!date) return null;
    const value = new Date(`${date}T${time || "00:00"}`);
    return Number.isNaN(value.getTime()) ? null : value.toISOString();
  }

  let startDate = $state("");
  let startTime = $state("");
  let endDate = $state("");
  let endTime = $state("");
  let error = $state("");

  $effect(() => {
    if (!open) return;
    untrack(() => {
      ({ date: startDate, time: startTime } = split(startsAt));
      ({ date: endDate, time: endTime } = split(endsAt));
      error = "";
    });
  });

  function handleSave(event: SubmitEvent) {
    event.preventDefault();
    error = "";
    if (!startDate || !startTime) {
      error = "Choose a start date and time.";
      return;
    }
    if (!endDate || !endTime) {
      error = "Choose an end date and time.";
      return;
    }
    const start = join(startDate, startTime);
    const end = join(endDate, endTime);
    if (start && end && end < start) {
      error = "The event must end after it starts.";
      return;
    }
    onSave(start, end);
  }
</script>

<Modal {open} title="Date & time" {onClose}>
  <form class="modal-form" onsubmit={handleSave} novalidate>
    <fieldset class="range">
      <legend>Starts</legend>
      <input type="date" class="text-input" aria-label="Start date" bind:value={startDate} />
      <input type="time" class="text-input" aria-label="Start time" bind:value={startTime} />
    </fieldset>
    <fieldset class="range">
      <legend>Ends</legend>
      <input type="date" class="text-input" aria-label="End date" min={startDate} bind:value={endDate} />
      <input type="time" class="text-input" aria-label="End time" bind:value={endTime} />
    </fieldset>
    <p class="hint">Times are in your time zone ({timeZone}).</p>
    {#if error}<p class="form-error" role="alert">{error}</p>{/if}
    <div class="modal-actions">
      <button type="button" class="button-outline" onclick={onClose}>Cancel</button>
      <button type="submit" class="button-primary">Save</button>
    </div>
  </form>
</Modal>

<style lang="sass">
  .modal-form
    display: flex
    flex-direction: column
    gap: 14px

  .range
    display: grid
    grid-template-columns: 1fr 140px
    gap: 8px
    margin: 0
    padding: 0
    border: none

    legend
      margin-bottom: 6px
      padding: 0
      font-size: 13px
      font-weight: 600
      color: var(--color-text-muted)

  .hint
    margin: 0
    font-size: 12px
    color: var(--color-text-muted)

  .form-error
    margin: 0
    font-size: 13px
    color: var(--color-error)

  .modal-actions
    display: flex
    justify-content: flex-end
    gap: 8px

  @media (max-width: 480px)
    .range
      grid-template-columns: 1fr
</style>
