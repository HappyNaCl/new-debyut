<script lang="ts">
  import "../../../styles/search.sass";
  import { router } from "svelte-spa-router";
  import { EventBusy } from "@material-symbols-svg/svelte";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import { createCategoriesQuery } from "../../api/categories";

  let parameters = $derived(new URLSearchParams(router.querystring ?? ""));
  let query = $derived(parameters.get("q")?.trim() ?? "");
  const categories = createCategoriesQuery();
  let category = $derived(categories.data?.find((candidate) => candidate.slug === parameters.get("category")));
</script>

<div class="search-root">
  <SiteHeader />

  <main class="search-content">
    <h1>
      {#if category}
        {category.name}
      {:else if query}
        Results for “{query}”
      {:else}
        All events
      {/if}
    </h1>

    <div class="empty-state">
      <EventBusy size={48} />
      <p>No events found yet.</p>
    </div>
  </main>
</div>
