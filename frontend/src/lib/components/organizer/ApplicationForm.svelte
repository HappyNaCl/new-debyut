<script lang="ts">
  import axios from "axios";
  import { onDestroy } from "svelte";
  import { ArrowBack, Close, PhotoCamera } from "@material-symbols-svg/svelte";
  import Avatar from "../ui/Avatar.svelte";
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

  const AVATAR_MAXIMUM_BYTES = 2 * 1024 * 1024;
  const AVATAR_TYPES = ["image/jpeg", "image/png", "image/webp"];

  let avatarInput = $state<HTMLInputElement>();
  let avatarFile = $state<File | null>(null);
  let avatarPreview = $state("");

  let organizationName = $state("");
  let description = $state("");
  let loading = $state(false);
  let errorMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  function setAvatar(file: File | null) {
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
    avatarFile = file;
    avatarPreview = file ? URL.createObjectURL(file) : "";
  }

  function handleAvatarChange() {
    const file = avatarInput?.files?.[0];
    if (avatarInput) avatarInput.value = "";
    if (!file) return;

    const { avatar: _, ...rest } = fieldErrors;
    fieldErrors = rest;
    if (!AVATAR_TYPES.includes(file.type)) {
      fieldErrors = { ...fieldErrors, avatar: "Choose a JPG, PNG or WebP image." };
      return;
    }
    if (file.size > AVATAR_MAXIMUM_BYTES) {
      fieldErrors = { ...fieldErrors, avatar: "The image must be 2 MB or smaller." };
      return;
    }
    setAvatar(file);
  }

  onDestroy(() => {
    if (avatarPreview) URL.revokeObjectURL(avatarPreview);
  });

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
        avatar: avatarFile,
      });
      onSubmitted(application);
    } catch (error) {
      const status = axios.isAxiosError(error) ? error.response?.status : undefined;
      const serverFieldErrors = extractFieldErrors(error);
      if (Object.keys(serverFieldErrors).length > 0) {
        fieldErrors = serverFieldErrors;
        if (fieldErrors.avatar) fieldErrors.avatar = `The image ${fieldErrors.avatar}.`;
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
    <span class="field-label">Organization logo <span class="optional">(optional)</span></span>
    <div class="avatar-picker">
      <Avatar name={organizationName} src={avatarPreview} shape="rounded" size={72} />
      <div class="avatar-actions">
        <div class="avatar-buttons">
          <button type="button" class="button-outline" onclick={() => avatarInput?.click()} disabled={loading}>
            <PhotoCamera size={18} />
            {avatarFile ? "Change image" : "Upload image"}
          </button>
          {#if avatarFile}
            <button
              type="button"
              class="remove-avatar"
              aria-label="Remove image"
              title="Remove image"
              onclick={() => setAvatar(null)}
              disabled={loading}
            >
              <Close size={18} />
            </button>
          {/if}
        </div>
        <span class="hint">JPG, PNG or WebP, up to 2 MB.</span>
      </div>
    </div>
    <input
      bind:this={avatarInput}
      type="file"
      accept={AVATAR_TYPES.join(",")}
      hidden
      onchange={handleAvatarChange}
    />
    {#if fieldErrors.avatar}<span class="field-error">{fieldErrors.avatar}</span>{/if}
  </div>

  <div class="field">
    <label for="application-email">Organization email</label>
    <input id="application-email" type="email" value={email} disabled class:invalid={fieldErrors.organizationEmail} />
    {#if fieldErrors.organizationEmail}
      <span class="field-error">{fieldErrors.organizationEmail} Use "Change email" to pick another.</span>
    {/if}
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
  .field-label
    font-size: 13px
    font-weight: 500
    color: var(--color-text-muted)

  .avatar-picker
    display: flex
    align-items: center
    gap: 16px

  .avatar-actions
    display: flex
    flex-direction: column
    align-items: flex-start
    gap: 6px
    min-width: 0

  .avatar-buttons
    display: flex
    align-items: center
    gap: 6px

  .remove-avatar
    display: inline-flex
    align-items: center
    justify-content: center
    width: 36px
    height: 36px
    padding: 0
    border: 1px solid var(--color-border)
    border-radius: 8px
    background: none
    color: var(--color-text-muted)
    cursor: pointer

    &:hover
      color: var(--color-error)
      border-color: var(--color-error)

  .hint
    font-size: 12px
    color: var(--color-text-muted)

  .character-count
    align-self: flex-end
    font-size: 12px
    color: var(--color-text-subtle)
</style>
