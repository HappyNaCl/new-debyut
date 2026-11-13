<script lang="ts">
  import "../../../styles/event-detail.sass";
  import { onDestroy } from "svelte";
  import { router } from "svelte-spa-router";
  import {
    CalendarMonth,
    Category,
    Check,
    ConfirmationNumber,
    ContentCopy,
    Gavel,
    LiveTv,
    LocationOn,
    Map as MapIcon,
  } from "@material-symbols-svg/svelte";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import RichText from "../ui/RichText.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { createCategoriesQuery } from "../../api/categories";
  import { extractErrorMessage } from "../../api/client";
  import { createPublicEventQuery } from "../../api/events";
  import { formatRupiah } from "../../format";
  import type { PublicEvent, Ticket } from "../../types/event";

  let eventId = $derived((router.params as Record<string, string> | null | undefined)?.id);
  const event = createPublicEventQuery(() => eventId);
  const categories = createCategoriesQuery();

  type SectionId = "description" | "tickets" | "terms";

  let activeSection = $state<SectionId>("description");
  let copied = $state(false);
  let copyTimer: ReturnType<typeof setTimeout> | undefined;

  const dayFormat = new Intl.DateTimeFormat("en-GB", { day: "numeric", month: "short", year: "numeric" });
  const timeFormat = new Intl.DateTimeFormat("en-GB", { hour: "2-digit", minute: "2-digit" });
  const zoneFormat = new Intl.DateTimeFormat("en-GB", { timeZoneName: "short" });
  const shortDateTime = new Intl.DateTimeFormat("en-GB", {
    day: "numeric",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });

  function zoneName(date: Date) {
    return zoneFormat.formatToParts(date).find((part) => part.type === "timeZoneName")?.value ?? "";
  }

  /** e.g. "14 Nov 2026, 05:00 – 23:50 GMT+7", or a range across days. */
  function eventWhen(source: PublicEvent): string {
    if (!source.startsAt) return "Date to be announced";
    const start = new Date(source.startsAt);
    const end = source.endsAt ? new Date(source.endsAt) : null;
    const zone = zoneName(start);
    if (!end) return `${dayFormat.format(start)}, ${timeFormat.format(start)} ${zone}`;
    if (dayFormat.format(start) === dayFormat.format(end)) {
      return `${dayFormat.format(start)}, ${timeFormat.format(start)} – ${timeFormat.format(end)} ${zone}`;
    }
    return `${shortDateTime.format(start)} – ${shortDateTime.format(end)} ${zone}`;
  }

  function eventWhere(source: PublicEvent): string {
    const venue = [source.venueName, source.city].filter(Boolean).join(", ");
    return venue || "Streamed online on debyut";
  }

  function mapUrl(source: PublicEvent): string | null {
    if (source.latitude === null || source.longitude === null) return null;
    const { latitude: lat, longitude: lon } = source;
    return `https://www.openstreetmap.org/?mlat=${lat}&mlon=${lon}#map=17/${lat}/${lon}`;
  }

  function ticketPrice(ticket: Ticket) {
    return ticket.kind === "FREE" ? "Free" : formatRupiah(ticket.price);
  }

  type SaleState = { label: string; tone: "open" | "soon" | "closed" };

  function saleState(ticket: Ticket, now: Date): SaleState {
    if (ticket.salesStartAt && new Date(ticket.salesStartAt) > now) {
      return { label: `On sale ${shortDateTime.format(new Date(ticket.salesStartAt))}`, tone: "soon" };
    }
    if (ticket.salesEndAt && new Date(ticket.salesEndAt) <= now) return { label: "Sales ended", tone: "closed" };
    return { label: "Available", tone: "open" };
  }

  function startingPrice(tickets: Ticket[]): string {
    if (tickets.length === 0) return "–";
    const lowest = Math.min(...tickets.map((ticket) => (ticket.kind === "FREE" ? 0 : ticket.price)));
    return lowest === 0 ? "Free" : formatRupiah(lowest);
  }

  let categoryName = $derived(
    categories.data?.find((category) => category.id === event.data?.categoryId)?.name ?? "",
  );
  let hasTerms = $derived(!!event.data?.terms);
  let sections = $derived<{ id: SectionId; label: string }[]>([
    { id: "description", label: "Description" },
    { id: "tickets", label: "Tickets" },
    ...(hasTerms ? [{ id: "terms" as const, label: "Terms & agreement" }] : []),
  ]);
  let now = $state(new Date());

  /** In-page links: the router uses the URL hash, so scroll instead of navigating. */
  function goTo(id: SectionId, clickEvent?: MouseEvent) {
    clickEvent?.preventDefault();
    activeSection = id;
    document.getElementById(id)?.scrollIntoView({ behavior: "smooth", block: "start" });
  }

  // Highlight the section currently in view.
  let observer: IntersectionObserver | null = null;
  $effect(() => {
    if (!event.data) return;
    void sections;
    observer?.disconnect();
    observer = new IntersectionObserver(
      (entries) => {
        const visible = entries.filter((entry) => entry.isIntersecting);
        if (visible.length > 0) activeSection = visible[0].target.id as SectionId;
      },
      { rootMargin: "-140px 0px -60% 0px" },
    );
    for (const section of sections) {
      const element = document.getElementById(section.id);
      if (element) observer.observe(element);
    }
    now = new Date();
    return () => observer?.disconnect();
  });

  async function copyLink() {
    const url = window.location.href;
    try {
      await navigator.clipboard.writeText(url);
    } catch {
      // Clipboard API unavailable (e.g. insecure context): fall back to a hidden field.
      const field = document.createElement("textarea");
      field.value = url;
      field.setAttribute("readonly", "");
      field.style.position = "fixed";
      field.style.opacity = "0";
      document.body.append(field);
      field.select();
      document.execCommand("copy");
      field.remove();
    }
    copied = true;
    clearTimeout(copyTimer);
    copyTimer = setTimeout(() => (copied = false), 2000);
  }

  onDestroy(() => clearTimeout(copyTimer));
</script>

<svelte:head>
  <title>{event.data ? `${event.data.name} · debyut` : "Event · debyut"}</title>
</svelte:head>

<div class="event-detail-root">
  <SiteHeader />

  {#if event.isPending}
    <div class="page-content"><div class="state"><Spinner size={32} /></div></div>
  {:else if event.isError}
    <div class="page-content">
      <div class="state">
        <ConfirmationNumber size={40} />
        <p>{extractErrorMessage(event.error)}</p>
        <a href="#/search" class="button-primary">Find other events</a>
      </div>
    </div>
  {:else}
    {@const detail = event.data}
    {@const map = mapUrl(detail)}

    {#snippet facts(compact: boolean)}
      <ul class="facts" class:compact>
        <li>
          {#if detail.venueName}<LocationOn size={18} />{:else}<LiveTv size={18} />{/if}
          <span>{eventWhere(detail)}</span>
        </li>
        <li><CalendarMonth size={18} /> <span>{eventWhen(detail)}</span></li>
        {#if categoryName}<li><Category size={18} /> <span>{categoryName}</span></li>{/if}
      </ul>
    {/snippet}

    <header class="hero">
      {#if detail.bannerUrl}
        <div class="hero-backdrop" style:background-image={`url("${detail.bannerUrl}")`}></div>
      {/if}
      <div class="hero-inner">
        <div class="hero-text">
          <h1>{detail.name}</h1>
          {@render facts(false)}
        </div>
      </div>
    </header>

    <nav class="section-nav" aria-label="Event sections">
      <div class="section-nav-inner">
        {#each sections as section (section.id)}
          <a
            href={`#${section.id}`}
            class="section-link"
            class:active={activeSection === section.id}
            aria-current={activeSection === section.id ? "true" : undefined}
            onclick={(clickEvent) => goTo(section.id, clickEvent)}
          >
            {section.label}
          </a>
        {/each}
      </div>
    </nav>

    <div class="layout">
      <main class="main-column">
        <section id="description" class="detail-section" aria-labelledby="description-title">
          <h2 id="description-title" class="visually-hidden">Description</h2>
          {#if detail.description}
            <RichText html={detail.description} />
          {:else}
            <p class="muted">The organizer hasn't added a description.</p>
          {/if}

          {#if detail.galleryUrls.length > 0}
            <div class="gallery" aria-label="Photo gallery">
              {#each detail.galleryUrls as url, index (url)}
                <a href={url} target="_blank" rel="noreferrer" class="gallery-photo">
                  <img src={url} alt={`${detail.name} photo ${index + 1}`} loading="lazy" />
                </a>
              {/each}
            </div>
          {/if}

          {#if detail.venueName}
            <div class="venue">
              <LocationOn size={22} />
              <div>
                <p class="venue-name">{detail.venueName}</p>
                <p class="venue-address">{[detail.address, detail.city].filter(Boolean).join(", ")}</p>
                {#if map}
                  <a href={map} target="_blank" rel="noreferrer" class="venue-map"><MapIcon size={16} /> View on map</a>
                {/if}
              </div>
            </div>
          {/if}
        </section>

        <section id="tickets" class="detail-section" aria-labelledby="tickets-title">
          <h2 id="tickets-title" class="section-title"><ConfirmationNumber size={24} /> Tickets</h2>
          {#if detail.tickets.length === 0}
            <p class="muted">No tickets yet.</p>
          {:else}
            <ul class="ticket-list">
              {#each detail.tickets as ticket (ticket.id)}
                {@const state = saleState(ticket, now)}
                <li class="ticket" class:closed={state.tone === "closed"}>
                  <span class="ticket-price">{ticketPrice(ticket)}</span>
                  <div class="ticket-info">
                    <span class="ticket-name">
                      {ticket.name}
                      {#if ticket.kind === "STREAMING"}<span class="streaming-tag"><LiveTv size={14} /> Streaming</span>{/if}
                    </span>
                    {#if ticket.description}<span class="ticket-description">{ticket.description}</span>{/if}
                    {#if ticket.salesEndAt && state.tone !== "closed"}
                      <span class="ticket-meta">Sales end {shortDateTime.format(new Date(ticket.salesEndAt))}</span>
                    {/if}
                  </div>
                  <span class="sale-state {state.tone}">{state.label}</span>
                </li>
              {/each}
            </ul>
            {#if detail.serviceFeeBearer === "BUYER" && detail.tickets.some((ticket) => ticket.kind !== "FREE")}
              <p class="fee-note">Prices exclude a service fee, added at checkout.</p>
            {/if}
          {/if}
        </section>

        {#if hasTerms}
          <section id="terms" class="detail-section" aria-labelledby="terms-title">
            <h2 id="terms-title" class="section-title"><Gavel size={24} /> Terms & agreement</h2>
            <RichText html={detail.terms} />
          </section>
        {/if}
      </main>

      <aside class="side-column">
        <div class="summary-card">
          {#if detail.coverUrl || detail.bannerUrl}
            <img class="summary-image" class:square={!!detail.coverUrl} src={detail.coverUrl || detail.bannerUrl} alt="" />
          {/if}
          <div class="price-bar">
            <div>
              <span class="price-label">Starts from</span>
              <span class="price-value">{startingPrice(detail.tickets)}</span>
            </div>
            <a href="#tickets" class="button-primary" onclick={(clickEvent) => goTo("tickets", clickEvent)}>
              See tickets
            </a>
          </div>
          <div class="summary-body">
            <h2 class="summary-title">{detail.name}</h2>
            {@render facts(true)}
            <div class="organizer">
              <Avatar name={detail.organizer.name} src={detail.organizer.avatarUrl} size={40} shape="rounded" />
              <div>
                <span class="organizer-label">Organized by</span>
                <span class="organizer-name">{detail.organizer.name}</span>
              </div>
            </div>
          </div>
        </div>

        <div class="share">
          <span class="share-label">Share event</span>
          <button type="button" class="copy-button" class:copied onclick={copyLink} aria-live="polite">
            {#if copied}<Check size={18} /> Link copied{:else}<ContentCopy size={18} /> Copy link{/if}
          </button>
        </div>
      </aside>
    </div>
  {/if}
</div>
