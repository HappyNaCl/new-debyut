<script lang="ts">
  import axios from "axios";
  import { ArrowBack } from "@material-symbols-svg/svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import { submitApplication } from "../../api/organizerApplications";
  import type { OrganizerApplication } from "../../types/organizerApplication";

  interface Props {
    email: string;
    verificationToken: string;
    onSubmitted: (application: OrganizerApplication) => void;
    onExpired: (message: string) => void;
    onConflict: () => void;
    onBack: () => void;
  }

  let { email, verificationToken, onSubmitted, onExpired, onConflict, onBack }: Props = $props();

  const NAME_LIMIT = 150;
  const DESCRIPTION_LIMIT = 2000;

  let organizationName = $state("");
  let description = $state("");
  let loading = $state(false);
  let errorMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  function validate(): Record<string, string> {
    const errors: Record<string, string> = {};
    if (!organizationName.trim()) errors.organizationName = "Organization name is required.";
    if (!description.trim()) errors.description = "Tell us a little about your organization.";
    return errors;
  }

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    fieldErrors = validate();
    if (Object.keys(fieldErrors).length > 0) return;

    loading = true;
    try {
      const application = await submitApplication({
        verificationToken,
        organizationName: organizationName.trim(),
        description: description.trim(),
      });
      onSubmitted(application);
    } catch (error) {
      const status = axios.isAxiosError(error) ? error.response?.status : undefined;
      const serverFieldErrors = extractFieldErrors(error);
      if (Object.keys(serverFieldErrors).length > 0) {
        fieldErrors = serverFieldErrors;
      } else if (status === 400) {
        onExpired(extractErrorMessage(error));
      } else {
        errorMessage = extractErrorMessage(error);
        if (status === 409) onConflict();
      }
    } finally {
      loading = false;
    }
  }
</script>

<div class="auth-heading">
  <h1>About your organization</h1>
  <p>Our team reviews every application, usually within a few working days.</p>
</div>

<form class="auth-card" onsubmit={handleSubmit} novalidate>
  <button type="button" class="back-button" onclick={onBack}>
    <ArrowBack size={18} /> Change email
  </button>

  {#if errorMessage}
    <div class="alert-error" role="alert">{errorMessage}</div>
  {/if}

  <div class="field">
    <label for="application-email">Organization email</label>
    <input id="application-email" type="email" value={email} disabled />
  </div>

  <div class="field">
    <label for="application-name">Organization name</label>
    <input
      id="application-name"
      type="text"
      autocomplete="organization"
      maxlength={NAME_LIMIT}
      class:invalid={fieldErrors.organizationName}
      bind:value={organizationName}
    />
    {#if fieldErrors.organizationName}<span class="field-error">{fieldErrors.organizationName}</span>{/if}
  </div>

  <div class="field">
    <label for="application-description">Description</label>
    <textarea
      id="application-description"
      maxlength={DESCRIPTION_LIMIT}
      placeholder="What kind of events do you organize, and where?"
      class:invalid={fieldErrors.description}
      bind:value={description}
    ></textarea>
    <span class="character-count">{description.length}/{DESCRIPTION_LIMIT}</span>
    {#if fieldErrors.description}<span class="field-error">{fieldErrors.description}</span>{/if}
  </div>

  <button type="submit" class="button-primary" disabled={loading}>
    {loading ? "Submitting…" : "Submit application"}
  </button>
</form>

<style lang="sass">
  .character-count
    align-self: flex-end
    font-size: 12px
    color: var(--color-text-subtle)
</style>
