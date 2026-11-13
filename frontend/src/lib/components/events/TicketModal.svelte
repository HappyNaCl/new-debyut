<script lang="ts">
  import { untrack } from "svelte";
  import Modal from "../ui/Modal.svelte";
  import { fromLocalInput, toLocalInput } from "../../format";
  import { validateTicket, type FieldErrors } from "../../events/validation";
  import type { TicketInput, TicketKind } from "../../types/event";

  interface Props {
    open: boolean;
    /** The ticket being edited, or a fresh one of the chosen kind. */
    ticket: TicketInput;
    isNew: boolean;
    eventEnd: string | null;
    /** Server-side errors for this ticket, shown until the user edits it. */
    serverErrors?: FieldErrors;
    onSave: (ticket: TicketInput) => void;
    onClose: () => void;
  }

  let { open, ticket, isNew, eventEnd, serverErrors = {}, onSave, onClose }: Props = $props();

  const kindLabels: Record<TicketKind, string> = {
    PAID: "Paid ticket",
    FREE: "Free ticket",
    STREAMING: "Streaming ticket",
  };

  let name = $state("");
  let description = $state("");
  let price = $state<number | null>(null);
  let quantity = $state<number | null>(null);
  let salesStart = $state("");
  let salesEnd = $state("");
  let errors = $state<FieldErrors>({});

  $effect(() => {
    if (!open) return;
    untrack(() => {
      name = ticket.name;
      description = ticket.description;
      price = isNew ? null : ticket.price;
      quantity = isNew ? null : ticket.quantity;
      salesStart = toLocalInput(ticket.salesStartAt);
      salesEnd = toLocalInput(ticket.salesEndAt);
      errors = { ...serverErrors };
    });
  });

  function handleSave(event: SubmitEvent) {
    event.preventDefault();
    const next: TicketInput = {
      kind: ticket.kind,
      name: name.trim(),
      description: description.trim(),
      price: ticket.kind === "FREE" ? 0 : Math.round(price ?? 0),
      quantity: Math.round(quantity ?? 0),
      salesStartAt: fromLocalInput(salesStart),
      salesEndAt: fromLocalInput(salesEnd),
    };
    errors = validateTicket(next, eventEnd, true);
    if (ticket.kind !== "FREE" && price === null) errors.price = "Enter a price above zero.";
    if (Object.keys(errors).length > 0) return;
    onSave(next);
  }
</script>

<Modal {open} title={`${isNew ? "New" : "Edit"} ${kindLabels[ticket.kind].toLowerCase()}`} {onClose}>
  <form class="modal-form" onsubmit={handleSave} novalidate>
    {#if ticket.kind === "STREAMING"}
      <p class="kind-note">Buyers watch the event online on debyut instead of attending in person.</p>
    {/if}
    <label class="field">
      <span>Ticket name</span>
      <input
        type="text"
        class="text-input"
        maxlength="100"
        placeholder="e.g. Early bird, VIP, Regular"
        class:invalid={errors.name}
        bind:value={name}
      />
      {#if errors.name}<span class="field-error">{errors.name}</span>{/if}
    </label>

    <div class="field-row">
      {#if ticket.kind !== "FREE"}
        <label class="field">
          <span>Price</span>
          <div class="money-input" class:invalid={errors.price}>
            <span class="currency">Rp</span>
            <input type="number" min="0" step="1000" inputmode="numeric" bind:value={price} />
          </div>
          {#if errors.price}<span class="field-error">{errors.price}</span>{/if}
        </label>
      {/if}
      <label class="field">
        <span>Quantity</span>
        <input
          type="number"
          class="text-input"
          min="1"
          step="1"
          inputmode="numeric"
          class:invalid={errors.quantity}
          bind:value={quantity}
        />
        {#if errors.quantity}<span class="field-error">{errors.quantity}</span>{/if}
      </label>
    </div>

    <div class="field-row">
      <label class="field">
        <span>Sales start</span>
        <input type="datetime-local" class="text-input" class:invalid={errors.salesStartAt} bind:value={salesStart} />
        {#if errors.salesStartAt}<span class="field-error">{errors.salesStartAt}</span>{/if}
      </label>
      <label class="field">
        <span>Sales end</span>
        <input
          type="datetime-local"
          class="text-input"
          min={salesStart}
          max={toLocalInput(eventEnd)}
          class:invalid={errors.salesEndAt}
          bind:value={salesEnd}
        />
        {#if errors.salesEndAt}<span class="field-error">{errors.salesEndAt}</span>{/if}
      </label>
    </div>

    <label class="field">
      <span>Description <span class="optional">(optional)</span></span>
      <textarea
        class="text-input textarea"
        maxlength="500"
        rows="3"
        placeholder="What's included with this ticket?"
        class:invalid={errors.description}
        bind:value={description}
      ></textarea>
      <span class="counter">{description.length}/500</span>
      {#if errors.description}<span class="field-error">{errors.description}</span>{/if}
    </label>

    <div class="modal-actions">
      <button type="button" class="button-outline" onclick={onClose}>Cancel</button>
      <button type="submit" class="button-primary">{isNew ? "Add ticket" : "Save ticket"}</button>
    </div>
  </form>
</Modal>

<style lang="sass">
  .modal-form
    display: flex
    flex-direction: column
    gap: 14px

  .kind-note
    margin: 0
    padding: 10px 12px
    border-radius: 8px
    background: var(--color-primary-soft)
    font-size: 13px
    line-height: 1.5
    color: var(--color-text)

  .field-row
    display: grid
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr))
    gap: 12px

  .field
    display: flex
    flex-direction: column
    gap: 6px
    min-width: 0
    font-size: 13px
    font-weight: 600
    color: var(--color-text-muted)

    input
      width: 100%
      min-width: 0

  .optional
    font-weight: 400

  .textarea
    height: auto
    padding: 10px 12px
    resize: vertical
    font-family: inherit

  .money-input
    display: flex
    align-items: center
    height: 40px
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 8px
    background: var(--color-surface)

    &:focus-within
      border-color: var(--color-primary)
      box-shadow: var(--focus-ring)

    &.invalid
      border-color: var(--color-error)

    input
      flex: 1
      height: 100%
      padding: 0 12px 0 6px
      border: none
      background: none
      font-size: 14px
      color: var(--color-text)
      outline: none

  .currency
    padding-left: 12px
    font-size: 14px
    font-weight: 700
    color: var(--color-text-muted)

  .invalid
    border-color: var(--color-error)

  .field-error, .counter
    font-size: 12px
    font-weight: 400

  .counter
    align-self: flex-end
    color: var(--color-text-subtle)

  .field-error
    color: var(--color-error)

  .modal-actions
    display: flex
    justify-content: flex-end
    gap: 8px
</style>
