<script lang="ts">
  import "../../../styles/home.sass";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import { user } from "../../stores/user";
</script>

<div class="home-root">
  <SiteHeader />

  <main class="home-content">
    {#if $user}
      <h1>Hi, {$user.firstName}!</h1>
      <p>Find your next event using the search bar or browse by category.</p>
    {:else}
      <h1>Discover events you'll love</h1>
      <p>Create an account or log in to get started.</p>
    {/if}
  </main>
</div>
