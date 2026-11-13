<script lang="ts">
  import { CalendarMonth, ConfirmationNumber, LiveTv, LocationOn } from "@material-symbols-svg/svelte";
  import Avatar from "../ui/Avatar.svelte";
  import { formatRupiah } from "../../format";
  import type { EventSummary } from "../../types/event";

  interface Props {
    event: EventSummary;
    categoryName?: string;
  }

  let { event, categoryName = "" }: Props = $props();

  const dateFormat = new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", year: "numeric" });

  let when = $derived.by(() => {
    if (!event.startsAt) return "Date to be announced";
    const start = dateFormat.format(new Date(event.startsAt));
    const end = event.endsAt ? dateFormat.format(new Date(event.endsAt)) : start;
    return start === end ? start : `${start} – ${end}`;
  });
  let where = $derived(
    [event.venueName, event.city].filter(Boolean).join(", ") || (event.streamingOnly ? "Streamed online" : ""),
  );
  let price = $derived(
    event.startingPrice === null ? "" : event.startingPrice === 0 ? "Free" : formatRupiah(event.startingPrice),
  );
</script>

<a href={`#/events/${event.id}`} class="event-card">
  <div class="banner">
    {#if event.bannerUrl}
      <img src={event.bannerUrl} alt="" loading="lazy" />
    {:else}
      <ConfirmationNumber size={36} />
    {/if}
    {#if categoryName}<span class="category">{categoryName}</span>{/if}
  </div>
  <div class="body">
    <h3 class="name">{event.name}</h3>
    <span class="line"><CalendarMonth size={16} /> {when}</span>
    {#if where}
      <span class="line">
        {#if event.streamingOnly}<LiveTv size={16} />{:else}<LocationOn size={16} />{/if}
        {where}
      </span>
    {/if}
    {#if price}
      <span class="price">{event.startingPrice ? `From ${price}` : price}</span>
    {/if}
  </div>
  <div class="organizer">
    <Avatar name={event.organizer.name} src={event.organizer.avatarUrl} size={28} shape="rounded" />
    <span>{event.organizer.name}</span>
  </div>
</a>

<style lang="sass">
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
    transition: transform 0.15s, border-color 0.15s

    &:hover
      transform: translateY(-2px)
      border-color: var(--color-primary)
      text-decoration: none

    &:focus-visible
      outline: 2px solid var(--color-primary)
      outline-offset: 2px

  .banner
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

  .category
    position: absolute
    left: 10px
    top: 10px
    padding: 2px 10px
    border-radius: 999px
    background: rgba(15, 17, 21, 0.72)
    font-size: 12px
    font-weight: 700
    color: #fff

  .body
    display: flex
    flex-direction: column
    gap: 6px
    flex: 1
    padding: 14px 16px

  .name
    display: -webkit-box
    margin: 0 0 2px
    overflow: hidden
    font-size: 16px
    font-weight: 800
    line-height: 1.3
    -webkit-box-orient: vertical
    -webkit-line-clamp: 2
    overflow-wrap: anywhere

  .line
    display: flex
    align-items: center
    gap: 6px
    min-width: 0
    font-size: 13px
    color: var(--color-text-muted)

  .price
    margin-top: auto
    padding-top: 6px
    font-size: 15px
    font-weight: 800

  .organizer
    display: flex
    align-items: center
    gap: 8px
    padding: 10px 16px
    border-top: 1px solid var(--color-border)
    font-size: 13px
    font-weight: 600
    color: var(--color-text-muted)

    span
      overflow: hidden
      text-overflow: ellipsis
      white-space: nowrap
</style>
