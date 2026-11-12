<script lang="ts">
  import "../../../styles/organizer.sass";
  import OrganizerLayout from "../organizer/OrganizerLayout.svelte";
  import ToggleSwitch from "../ui/ToggleSwitch.svelte";
  import { theme, toggleTheme } from "../../stores/theme";
</script>

<div class="organizer-root">
  <OrganizerLayout title="Settings">
    {#snippet children()}
      <section class="settings-list">
        <div class="setting-row">
          <div class="setting-text">
            <span class="setting-title">Dark mode</span>
            <span class="setting-description">Use a dark theme on this device.</span>
          </div>
          <ToggleSwitch label="Dark mode" checked={$theme === "dark"} onChange={toggleTheme} />
        </div>
      </section>
    {/snippet}
  </OrganizerLayout>
</div>
