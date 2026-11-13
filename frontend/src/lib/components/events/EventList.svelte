<script lang="ts">
  import { Add, CalendarMonth, ConfirmationNumber, LocationOn } from "@material-symbols-svg/svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { createEventsQuery } from "../../api/events";
  import { formatDateTime, formatRupiah } from "../../format";
  import type { EventFilter, OrganizerEvent } from "../../types/event";

  interface Props {
    organizationId: string;
    filter: EventFilter;
    emptyLabel: string;
  }

  let { organizationId, filter, emptyLabel }: Props = $props();

  const events = createEventsQuery(
    () => organizationId,
    () => filter,
  );

  function priceRange(event: OrganizerEvent) {
    if (event.tickets.length === 0) return "No tickets yet";
    const prices = event.tickets.map((ticket) => (ticket.kind === "FREE" ? 0 : ticket.price));
    const lowest = Math.min(...prices);
    const highest = Math.max(...prices);
    if (highest === 0) return "Free";
    return lowest === highest ? formatRupiah(lowest) : `${formatRupiah(lowest)} – ${formatRupiah(highest)}`;
  }

  function place(event: OrganizerEvent) {
    const venue = [event.venueName, event.city].filter(Boolean).join(", ");
    if (venue) return venue;
    return event.tickets.length > 0 && event.tickets.every((ticket) => ticket.kind === "STREAMING") ? "Streamed online" : "";
  }
</script>

{#if events.isPending}
  <div class="state"><Spinner size={28} /></div>
{:else if events.isError}
  <div class="state">
    <p>{extractErrorMessage(events.error)}</p>
    <button type="button" class="button-outline" onclick={() => events.refetch()}>Try again</button>
  </div>
{:else if events.data.length === 0}
  <div class="state events-empty">
    <ConfirmationNumber size={48} />
    <p>No {emptyLabel} yet.</p>
    {#if filter !== "PAST"}
      <a href={`#/organizer/${organizationId}/events/new`} class="button-primary"><Add size={18} /> Create event</a>
    {/if}
  </div>
{:else}
  <ul class="event-grid">
    {#each events.data as event (event.id)}
      {@const href = event.status === "DRAFT" ? `#/organizer/${organizationId}/events/${event.id}/edit` : `#/events/${event.id}`}
      <li>
        <svelte:element this={href ? "a" : "div"} {href} class="event-card">
          <div class="event-banner">
            {#if event.bannerUrl}
              <img src={event.bannerUrl} alt="" />
            {:else}
              <ConfirmationNumber size={32} />
            {/if}
            {#if event.status === "DRAFT"}<span class="draft-badge">Draft</span>{/if}
          </div>
          <div class="event-body">
            <span class="event-name">{event.name}</span>
            <span class="event-line">
              <CalendarMonth size={16} />
              {event.startsAt ? formatDateTime(event.startsAt) : "Date not set"}
            </span>
            <span class="event-line"><LocationOn size={16} /> {place(event) || "Location not set"}</span>
            <span class="event-price">{priceRange(event)}</span>
          </div>
        </svelte:element>
      </li>
    {/each}
  </ul>
{/if}

<style lang="sass">
  .event-grid
    display: grid
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr))
    gap: 16px
    margin: 0
    padding: 0
    list-style: none

  .event-card
    display: flex
    flex-direction: column
    height: 100%
    overflow: hidden
    border: 1px solid var(--color-border)
    border-radius: 12px
    background: var(--color-surface)
    color: var(--color-text)
    box-shadow: var(--card-shadow)
    transition: border-color 0.15s

  a.event-card:hover
    border-color: var(--color-primary)
    text-decoration: none

  .event-banner
    position: relative
    display: flex
    align-items: center
    justify-content: center
    aspect-ratio: 724 / 340
    background: var(--color-surface-muted)
    color: var(--color-text-subtle)

    img
      width: 100%
      height: 100%
      object-fit: cover

  .draft-badge
    position: absolute
    left: 10px
    top: 10px
    padding: 2px 10px
    border-radius: 999px
    background: var(--color-warning-soft)
    font-size: 12px
    font-weight: 800
    color: var(--color-warning)

  .event-body
    display: flex
    flex-direction: column
    gap: 6px
    padding: 14px 16px 16px

  .event-name
    font-size: 16px
    font-weight: 800
    overflow-wrap: anywhere

  .event-line
    display: inline-flex
    align-items: center
    gap: 6px
    font-size: 13px
    color: var(--color-text-muted)

  .event-price
    margin-top: 4px
    font-size: 14px
    font-weight: 800
</style>
