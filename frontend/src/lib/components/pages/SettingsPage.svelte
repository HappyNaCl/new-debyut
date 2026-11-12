<script lang="ts">
  import "../../../styles/settings.sass";
  import AccountLayout from "../account/AccountLayout.svelte";
  import ToggleSwitch from "../ui/ToggleSwitch.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { updateMe } from "../../api/user";
  import { theme, toggleTheme } from "../../stores/theme";
  import { loadUser, user } from "../../stores/user";

  loadUser().catch(() => {});

  let savingNewsletter = $state(false);
  let errorMessage = $state("");

  async function changeNewsletter(subscribeNewsletter: boolean) {
    const profile = $user;
    if (!profile) return;

    errorMessage = "";
    savingNewsletter = true;
    try {
      const updated = await updateMe({
        firstName: profile.firstName,
        surname: profile.surname,
        dateOfBirth: profile.dateOfBirth,
        phoneNumber: profile.phoneNumber,
        gender: profile.gender,
        subscribeNewsletter,
      });
      user.set(updated);
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      savingNewsletter = false;
    }
  }
</script>

<div class="settings-root">
  <AccountLayout title="Settings">
    {#if errorMessage}
      <div class="alert alert-error" role="alert">{errorMessage}</div>
    {/if}

    <section class="settings-list">
      <div class="setting-row">
        <div class="setting-text">
          <span class="setting-title">Newsletter</span>
          <span class="setting-description">Get the latest events and promotions in your inbox.</span>
        </div>
        <ToggleSwitch
          label="Newsletter"
          checked={$user?.subscribeNewsletter ?? false}
          disabled={!$user || savingNewsletter}
          onChange={changeNewsletter}
        />
      </div>
      <div class="setting-row">
        <div class="setting-text">
          <span class="setting-title">Dark mode</span>
          <span class="setting-description">Use a dark theme on this device.</span>
        </div>
        <ToggleSwitch label="Dark mode" checked={$theme === "dark"} onChange={toggleTheme} />
      </div>
    </section>
  </AccountLayout>
</div>
