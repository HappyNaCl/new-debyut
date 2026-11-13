<script lang="ts">
  import "../../../styles/home.sass";
  import { ConfirmationNumber } from "@material-symbols-svg/svelte";
  import EventCard from "../events/EventCard.svelte";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import { createCategoriesQuery } from "../../api/categories";
  import { extractErrorMessage } from "../../api/client";
  import { createLatestEventsQuery } from "../../api/events";
  import { user } from "../../stores/user";

  const LATEST_COUNT = 8;

  const events = createLatestEventsQuery(LATEST_COUNT);
  const categories = createCategoriesQuery();

  function categoryName(id: string) {
    return categories.data?.find((category) => category.id === id)?.name ?? "";
  }
</script>

<div class="home-root">
  <SiteHeader />

  <main class="home-content">
    <section class="home-intro">
      {#if $user}
        <h1>Hi, {$user.firstName}!</h1>
        <p>Find your next event using the search bar or browse by category.</p>
      {:else}
        <h1>Discover events you'll love</h1>
        <p>Create an account or log in to get started.</p>
      {/if}
    </section>

    <section class="latest" aria-labelledby="latest-title">
      <h2 id="latest-title">Latest events</h2>

      {#if events.isPending}
        <ul class="event-grid" aria-busy="true" aria-label="Loading events">
          {#each Array(LATEST_COUNT) as _, index (index)}
            <li class="skeleton"></li>
          {/each}
        </ul>
      {:else if events.isError}
        <div class="state">
          <p>{extractErrorMessage(events.error)}</p>
          <button type="button" class="button-outline" onclick={() => events.refetch()}>Try again</button>
        </div>
      {:else if events.data.length === 0}
        <div class="state">
          <ConfirmationNumber size={40} />
          <p>No upcoming events yet. Check back soon.</p>
        </div>
      {:else}
        <ul class="event-grid">
          {#each events.data as event (event.id)}
            <li><EventCard {event} categoryName={categoryName(event.categoryId)} /></li>
          {/each}
        </ul>
      {/if}
    </section>
  </main>
</div>
