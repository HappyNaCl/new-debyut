<script lang="ts">
  import "../../../styles/vod.sass";
  import { VideoLibrary } from "@material-symbols-svg/svelte";
  import AccountLayout from "../account/AccountLayout.svelte";
</script>

<div class="vod-root">
  <AccountLayout title="VOD">
    <div class="page-heading">
      <div>
        <p>Replays of the events you've bought access to.</p>
      </div>
    </div>

    <div class="state">
      <VideoLibrary size={40} />
      <p>No videos yet.</p>
      <a href="#/search" class="button-primary">Find events</a>
    </div>
  </AccountLayout>
</div>
