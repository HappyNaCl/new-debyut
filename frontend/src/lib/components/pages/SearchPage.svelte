<script lang="ts">
  import "../../../styles/search.sass";
  import { push, router } from "svelte-spa-router";
  import { Close, EventBusy, NearMe } from "@material-symbols-svg/svelte";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import EventCard from "../events/EventCard.svelte";
  import Pagination from "../ui/Pagination.svelte";
  import { createCategoriesQuery } from "../../api/categories";
  import { createEventSearchQuery } from "../../api/events";
  import { extractErrorMessage } from "../../api/client";
  import type { EventSearchParams, SearchFormat, SearchSort } from "../../types/event";

  const PAGE_SIZE = 24;
  const NEAR_RADIUS_KM = 25;

  type When = "" | "today" | "weekend" | "week" | "month";

  const whenOptions: { value: When; label: string }[] = [
    { value: "", label: "Any date" },
    { value: "today", label: "Today" },
    { value: "weekend", label: "This weekend" },
    { value: "week", label: "Next 7 days" },
    { value: "month", label: "This month" },
  ];

  const formatOptions: { value: string; label: string }[] = [
    { value: "", label: "Venue or online" },
    { value: "venue", label: "At a venue" },
    { value: "online", label: "Streamed online" },
  ];

  const sortLabels: Record<SearchSort, string> = {
    RELEVANCE: "Best match",
    SOONEST: "Soonest",
    NEWEST: "Newly added",
    PRICE_LOW: "Lowest price",
  };

  // The URL is the source of truth, so searches can be shared and the back
  // button works.
  let parameters = $derived(new URLSearchParams(router.querystring ?? ""));
  let query = $derived(parameters.get("q")?.trim() ?? "");
  let categorySlug = $derived(parameters.get("category") ?? "");
  let city = $derived(parameters.get("city") ?? "");
  let when = $derived((parameters.get("when") ?? "") as When);
  let format = $derived(parameters.get("format") ?? "");
  let free = $derived(parameters.get("free") === "1");
  let near = $derived(parseNear(parameters.get("near")));
  let sort = $derived.by<SearchSort>(() => {
    const value = parameters.get("sort");
    if (value && value in sortLabels) return value as SearchSort;
    return query ? "RELEVANCE" : "SOONEST";
  });
  let page = $derived(Math.max(1, Number.parseInt(parameters.get("page") ?? "1", 10) || 1));

  let searchParams = $derived.by<EventSearchParams>(() => {
    const params: EventSearchParams = { sort, page, pageSize: PAGE_SIZE };
    if (query) params.q = query;
    if (categorySlug) params.category = categorySlug;
    if (city) params.city = city;
    if (format === "venue" || format === "online") params.format = format.toUpperCase() as SearchFormat;
    if (free) params.free = true;
    if (near) Object.assign(params, { lat: near.lat, lng: near.lng, radiusKm: NEAR_RADIUS_KM });
    Object.assign(params, dateRange(when));
    return params;
  });

  const categories = createCategoriesQuery();
  const results = createEventSearchQuery(() => searchParams);

  let category = $derived(categories.data?.find((candidate) => candidate.slug === categorySlug));
  let cityOptions = $derived.by(() => {
    const counts = Object.entries(results.data?.cities ?? {}).sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0]));
    if (city && !counts.some(([name]) => name === city)) counts.unshift([city, 0]);
    return counts;
  });
  let hasFilters = $derived(!!(categorySlug || city || when || format || free || near));

  let locating = $state(false);
  let locationError = $state("");

  function parseNear(value: string | null) {
    const [lat, lng] = (value ?? "").split(",").map(Number);
    if (!Number.isFinite(lat) || !Number.isFinite(lng) || Math.abs(lat) > 90 || Math.abs(lng) > 180) return null;
    return { lat, lng };
  }

  function endOfDay(date: Date) {
    const end = new Date(date);
    end.setHours(23, 59, 59, 999);
    return end;
  }

  function dateRange(value: When): Pick<EventSearchParams, "from" | "to"> {
    const now = new Date();
    switch (value) {
      case "today":
        return { from: now.toISOString(), to: endOfDay(now).toISOString() };
      case "weekend": {
        // The coming Saturday and Sunday; on a weekend, the rest of this one.
        const sunday = new Date(now);
        sunday.setDate(now.getDate() + ((7 - now.getDay()) % 7));
        const saturday = new Date(sunday);
        saturday.setDate(sunday.getDate() - 1);
        saturday.setHours(0, 0, 0, 0);
        const from = saturday > now ? saturday : now;
        return { from: from.toISOString(), to: endOfDay(sunday).toISOString() };
      }
      case "week": {
        const end = new Date(now);
        end.setDate(now.getDate() + 7);
        return { from: now.toISOString(), to: endOfDay(end).toISOString() };
      }
      case "month":
        return { from: now.toISOString(), to: endOfDay(new Date(now.getFullYear(), now.getMonth() + 1, 0)).toISOString() };
      default:
        return {};
    }
  }

  // update changes some URL parameters; any filter change goes back to page 1.
  function update(changes: Record<string, string | null>) {
    const next = new URLSearchParams(parameters);
    for (const [key, value] of Object.entries(changes)) {
      if (value) next.set(key, value);
      else next.delete(key);
    }
    if (!("page" in changes)) next.delete("page");
    const querystring = next.toString();
    push(querystring ? `/search?${querystring}` : "/search");
  }

  function clearFilters() {
    update({ category: null, city: null, when: null, format: null, free: null, near: null });
  }

  function toggleNear() {
    locationError = "";
    if (near) {
      update({ near: null });
      return;
    }
    if (!("geolocation" in navigator)) {
      locationError = "Your browser can't share your location.";
      return;
    }
    locating = true;
    navigator.geolocation.getCurrentPosition(
      (position) => {
        locating = false;
        // About 100 m of precision is plenty for a 25 km radius.
        const { latitude, longitude } = position.coords;
        update({ near: `${latitude.toFixed(3)},${longitude.toFixed(3)}` });
      },
      () => {
        locating = false;
        locationError = "Couldn't get your location. Check your browser's location permission.";
      },
      { maximumAge: 10 * 60 * 1000, timeout: 10000 },
    );
  }

  function categoryName(id: string) {
    return categories.data?.find((candidate) => candidate.id === id)?.name ?? "";
  }

  function changePage(next: number) {
    update({ page: next > 1 ? String(next) : null });
    window.scrollTo({ top: 0, behavior: "smooth" });
  }
</script>

<div class="search-root">
  <SiteHeader />

  <main class="search-content">
    <header class="search-heading">
      <h1>
        {#if query}
          Results for “{query}”
        {:else if category}
          {category.name}
        {:else}
          All events
        {/if}
      </h1>
      {#if results.data}
        <p aria-live="polite">
          {results.data.total === 0
            ? "No events found"
            : `${results.data.total.toLocaleString()} ${results.data.total === 1 ? "event" : "events"}`}
          {#if query && category}in {category.name}{/if}
        </p>
      {/if}
    </header>

    <div class="filters" role="group" aria-label="Filters">
      <select
        aria-label="Category"
        value={categorySlug}
        onchange={(event) => update({ category: event.currentTarget.value })}
      >
        <option value="">All categories</option>
        {#each categories.data ?? [] as option (option.id)}
          <option value={option.slug}>{option.name}</option>
        {/each}
      </select>

      <select aria-label="City" value={city} onchange={(event) => update({ city: event.currentTarget.value })}>
        <option value="">All cities</option>
        {#each cityOptions as [name, count] (name)}
          <option value={name}>{name}{count ? ` (${count})` : ""}</option>
        {/each}
      </select>

      <select aria-label="Date" value={when} onchange={(event) => update({ when: event.currentTarget.value })}>
        {#each whenOptions as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>

      <select aria-label="Format" value={format} onchange={(event) => update({ format: event.currentTarget.value })}>
        {#each formatOptions as option (option.value)}
          <option value={option.value}>{option.label}</option>
        {/each}
      </select>

      <button
        type="button"
        class="chip"
        class:active={free}
        aria-pressed={free}
        onclick={() => update({ free: free ? null : "1" })}
      >
        Free
      </button>

      <button
        type="button"
        class="chip"
        class:active={!!near}
        aria-pressed={!!near}
        disabled={locating}
        onclick={toggleNear}
      >
        <NearMe size={16} />
        {locating ? "Locating…" : `Within ${NEAR_RADIUS_KM} km`}
      </button>

      {#if hasFilters}
        <button type="button" class="clear" onclick={clearFilters}>
          <Close size={16} /> Clear filters
        </button>
      {/if}

      <label class="sort">
        <span>Sort</span>
        <select value={sort} onchange={(event) => update({ sort: event.currentTarget.value })}>
          {#each Object.entries(sortLabels) as [value, label] (value)}
            {#if value !== "RELEVANCE" || query}
              <option {value}>{label}</option>
            {/if}
          {/each}
        </select>
      </label>
    </div>

    {#if locationError}
      <p class="alert alert-error" role="alert">{locationError}</p>
    {/if}

    {#if results.isPending}
      <ul class="event-grid" aria-busy="true" aria-label="Loading events">
        {#each Array(8) as _, index (index)}
          <li class="skeleton"></li>
        {/each}
      </ul>
    {:else if results.isError}
      <div class="state">
        <p>{extractErrorMessage(results.error)}</p>
        <button type="button" class="button-outline" onclick={() => results.refetch()}>Try again</button>
      </div>
    {:else if results.data.events.length === 0}
      <div class="state">
        <EventBusy size={48} />
        <p>
          {query ? `No upcoming events match “${query}”.` : "No upcoming events match these filters."}
        </p>
        {#if hasFilters}
          <button type="button" class="button-outline" onclick={clearFilters}>Clear filters</button>
        {/if}
      </div>
    {:else}
      <ul class="event-grid" class:stale={results.isPlaceholderData}>
        {#each results.data.events as event (event.id)}
          <li><EventCard {event} categoryName={categoryName(event.categoryId)} /></li>
        {/each}
      </ul>
      <Pagination page={results.data.page} pageSize={results.data.pageSize} total={results.data.total} onChange={changePage} />
    {/if}
  </main>
</div>
