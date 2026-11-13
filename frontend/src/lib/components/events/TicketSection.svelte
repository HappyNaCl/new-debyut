<script lang="ts">
  import { Add, Delete, Edit, Warning } from "@material-symbols-svg/svelte";
  import TicketModal from "./TicketModal.svelte";
  import Modal from "../ui/Modal.svelte";
  import { formatDateTime, formatRupiah } from "../../format";
  import { MAX_TICKET_TYPES, ticketErrors, type FieldErrors } from "../../events/validation";
  import type { TicketInput, TicketKind } from "../../types/event";

  interface Props {
    tickets: TicketInput[];
    eventEnd: string | null;
    errors: FieldErrors;
    onChange: (tickets: TicketInput[]) => void;
  }

  let { tickets, eventEnd, errors, onChange }: Props = $props();

  const kinds: { kind: TicketKind; label: string; caption: string }[] = [
    { kind: "PAID", label: "Paid", caption: "Set a fixed price" },
    { kind: "FREE", label: "Free", caption: "No charge" },
    { kind: "STREAMING", label: "Streaming", caption: "Paid online viewing" },
  ];
  const kindNames: Record<TicketKind, string> = { PAID: "Paid", FREE: "Free", STREAMING: "Streaming" };

  let editing = $state<{ index: number | null; ticket: TicketInput } | null>(null);
  let removingIndex = $state<number | null>(null);

  let atLimit = $derived(tickets.length >= MAX_TICKET_TYPES);

  function startNew(kind: TicketKind) {
    editing = {
      index: null,
      ticket: { kind, name: "", description: "", price: 0, quantity: 0, salesStartAt: null, salesEndAt: null },
    };
  }

  function save(ticket: TicketInput) {
    if (!editing) return;
    const next = [...tickets];
    if (editing.index === null) next.push(ticket);
    else next[editing.index] = ticket;
    onChange(next);
    editing = null;
  }

  function remove() {
    if (removingIndex === null) return;
    onChange(tickets.filter((_, index) => index !== removingIndex));
    removingIndex = null;
  }

  function priceLabel(ticket: TicketInput) {
    if (ticket.kind === "FREE") return "Free";
    return formatRupiah(ticket.price);
  }
</script>

<div class="ticket-section">
  <div class="kind-grid">
    {#each kinds as option (option.kind)}
      <button type="button" class="ticket-stub" onclick={() => startNew(option.kind)} disabled={atLimit}>
        <span class="stub-text">
          <span class="stub-caption">Create ticket</span>
          <span class="stub-label">{option.label}</span>
          <span class="stub-hint">{option.caption}</span>
        </span>
        <span class="stub-add" aria-hidden="true"><Add size={20} /></span>
      </button>
    {/each}
  </div>

  {#if errors.tickets}<p class="section-error" role="alert">{errors.tickets}</p>{/if}

  {#if tickets.length > 0}
    <ul class="ticket-list">
      {#each tickets as ticket, index (index)}
        {@const problems = ticketErrors(errors, index)}
        {@const hasProblems = Object.keys(problems).length > 0}
        <li class="ticket-row" class:has-problems={hasProblems}>
          <div class="ticket-main">
            <span class="ticket-kind {ticket.kind.toLowerCase()}">{kindNames[ticket.kind]}</span>
            <span class="ticket-name">{ticket.name}</span>
            <span class="ticket-meta">
              {ticket.quantity.toLocaleString("en")} available
              {#if ticket.salesStartAt || ticket.salesEndAt}
                · Sales {formatDateTime(ticket.salesStartAt) || "…"} – {formatDateTime(ticket.salesEndAt) || "…"}
              {/if}
            </span>
            {#if hasProblems}
              <span class="ticket-problem"><Warning size={16} /> {Object.values(problems)[0]}</span>
            {/if}
          </div>
          <span class="ticket-price">{priceLabel(ticket)}</span>
          <div class="row-actions">
            <button
              type="button"
              class="icon-button"
              aria-label={`Edit ${ticket.name}`}
              title="Edit"
              onclick={() => (editing = { index, ticket })}
            >
              <Edit size={20} />
            </button>
            <button
              type="button"
              class="icon-button danger"
              aria-label={`Delete ${ticket.name}`}
              title="Delete"
              onclick={() => (removingIndex = index)}
            >
              <Delete size={20} />
            </button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</div>

<TicketModal
  open={editing !== null}
  ticket={editing?.ticket ?? { kind: "PAID", name: "", description: "", price: 0, quantity: 0, salesStartAt: null, salesEndAt: null }}
  isNew={editing?.index === null}
  {eventEnd}
  serverErrors={editing && editing.index !== null ? ticketErrors(errors, editing.index) : {}}
  onSave={save}
  onClose={() => (editing = null)}
/>

<Modal open={removingIndex !== null} title="Delete ticket" size="small" onClose={() => (removingIndex = null)}>
  {#if removingIndex !== null && tickets[removingIndex]}
    <p class="confirm-message">Delete <strong>{tickets[removingIndex].name}</strong>?</p>
    <div class="modal-actions">
      <button type="button" class="button-outline" onclick={() => (removingIndex = null)}>Cancel</button>
      <button type="button" class="button-danger" onclick={remove}>Delete</button>
    </div>
  {/if}
</Modal>

<style lang="sass">
  .ticket-section
    display: flex
    flex-direction: column
    gap: 16px

  .kind-grid
    display: grid
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr))
    gap: 16px

  // A ticket stub: notches cut into both short sides.
  .ticket-stub
    --notch: 9px
    position: relative
    display: flex
    align-items: center
    gap: 12px
    min-height: 104px
    padding: 24px 20px 24px 28px
    border: 1px solid var(--color-border)
    border-radius: 10px
    background: var(--color-surface)
    color: var(--color-text)
    text-align: left
    cursor: pointer
    transition: border-color 0.15s, box-shadow 0.15s

    &::before, &::after
      content: ""
      position: absolute
      top: 50%
      width: calc(var(--notch) * 2)
      height: calc(var(--notch) * 2)
      border: 1px solid var(--color-border)
      border-radius: 50%
      background: var(--color-surface-muted)
      transform: translateY(-50%)

    &::before
      left: calc(var(--notch) * -1 - 1px)
      clip-path: inset(0 0 0 50%)

    &::after
      right: calc(var(--notch) * -1 - 1px)
      clip-path: inset(0 50% 0 0)

    &:hover:not(:disabled)
      border-color: var(--color-primary)
      box-shadow: var(--card-shadow)

      .stub-add
        background: var(--color-primary)
        border-color: var(--color-primary)
        color: var(--color-on-primary)

    &:focus-visible
      outline: 2px solid var(--color-primary)
      outline-offset: 2px

    &:disabled
      opacity: 0.5
      cursor: not-allowed

  .stub-text
    display: flex
    flex-direction: column
    gap: 2px
    flex: 1
    min-width: 0

  .stub-caption
    font-size: 12px
    color: var(--color-text-muted)

  .stub-label
    font-size: 16px
    font-weight: 800

  .stub-hint
    font-size: 12px
    color: var(--color-text-subtle)

  .stub-add
    display: inline-flex
    align-items: center
    justify-content: center
    flex-shrink: 0
    width: 34px
    height: 34px
    border: 1px solid var(--color-border)
    border-radius: 50%
    color: var(--color-text-muted)
    transition: background 0.15s, color 0.15s

  .section-error
    margin: 0
    font-size: 13px
    color: var(--color-error)

  .ticket-list
    display: flex
    flex-direction: column
    gap: 8px
    margin: 0
    padding: 0
    list-style: none

  .ticket-row
    display: flex
    align-items: center
    gap: 16px
    padding: 14px 16px
    border: 1px solid var(--color-border)
    border-radius: 10px
    background: var(--color-surface)

    &.has-problems
      border-color: var(--color-error)

  .ticket-main
    display: flex
    flex-direction: column
    align-items: flex-start
    gap: 4px
    flex: 1
    min-width: 0

  .ticket-kind
    padding: 2px 8px
    border-radius: 999px
    font-size: 11px
    font-weight: 800
    text-transform: uppercase
    letter-spacing: 0.3px

    &.paid
      background: var(--color-primary-soft)
      color: var(--color-primary)

    &.free
      background: var(--color-info-soft)
      color: var(--color-info)

    &.streaming
      background: var(--color-warning-soft)
      color: var(--color-warning)

  .ticket-name
    font-size: 15px
    font-weight: 800
    overflow-wrap: anywhere

  .ticket-meta
    font-size: 12px
    color: var(--color-text-muted)

  .ticket-problem
    display: inline-flex
    align-items: center
    gap: 4px
    font-size: 12px
    color: var(--color-error)

  .ticket-price
    flex-shrink: 0
    font-size: 15px
    font-weight: 800

  .confirm-message
    margin: 0
    font-size: 14px

  .modal-actions
    display: flex
    justify-content: flex-end
    gap: 8px

  @media (max-width: 640px)
    .ticket-row
      flex-wrap: wrap

    .ticket-main
      flex-basis: 100%

    .row-actions
      margin-left: auto
</style>
