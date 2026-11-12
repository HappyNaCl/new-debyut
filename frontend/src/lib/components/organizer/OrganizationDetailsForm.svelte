<script lang="ts">
  import { untrack } from "svelte";
  import AvatarUploader from "../ui/AvatarUploader.svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import { organizationsQueryKey, updateOrganization, uploadOrganizationAvatar } from "../../api/organizations";
  import { queryClient } from "../../queryClient";
  import type { Organization } from "../../types/organization";

  interface Props {
    organization: Organization;
  }

  let { organization }: Props = $props();

  const NAME_LIMIT = 150;
  const DESCRIPTION_LIMIT = 2000;

  let name = $state(untrack(() => organization.name));
  let description = $state(untrack(() => organization.description));

  let saving = $state(false);
  let errorMessage = $state("");
  let successMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  let isOwner = $derived(organization.myRole === "OWNER");

  async function handleAvatarUpload(file: File) {
    await uploadOrganizationAvatar(organization.id, file);
    await queryClient.invalidateQueries({ queryKey: organizationsQueryKey });
  }

  function validate(): Record<string, string> {
    const errors: Record<string, string> = {};
    if (!name.trim()) errors.name = "Organization name is required.";
    if (!description.trim()) errors.description = "Description is required.";
    return errors;
  }

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    successMessage = "";
    fieldErrors = validate();
    if (Object.keys(fieldErrors).length > 0) return;

    saving = true;
    try {
      await updateOrganization(organization.id, { name: name.trim(), description: description.trim() });
      successMessage = "Organization details saved.";
      await queryClient.invalidateQueries({ queryKey: organizationsQueryKey });
    } catch (error) {
      const serverFieldErrors = extractFieldErrors(error);
      if (Object.keys(serverFieldErrors).length > 0) fieldErrors = serverFieldErrors;
      else errorMessage = extractErrorMessage(error);
    } finally {
      saving = false;
    }
  }
</script>

<form class="card dashboard-form" onsubmit={handleSubmit} novalidate>
  {#if !isOwner}
    <div class="alert alert-success" role="status">Only owners can change these details.</div>
  {/if}
  {#if successMessage}
    <div class="alert alert-success" role="status">{successMessage}</div>
  {/if}
  {#if errorMessage}
    <div class="alert alert-error" role="alert">{errorMessage}</div>
  {/if}

  <AvatarUploader
    name={organization.name}
    src={organization.avatarUrl}
    shape="rounded"
    disabled={!isOwner}
    onUpload={handleAvatarUpload}
  />

  <div class="field">
    <label for="organization-name">Organization name</label>
    <input
      id="organization-name"
      maxlength={NAME_LIMIT}
      disabled={!isOwner}
      class:invalid={fieldErrors.name}
      bind:value={name}
    />
    {#if fieldErrors.name}<span class="field-error">{fieldErrors.name}</span>{/if}
  </div>

  <div class="field">
    <label for="organization-email">Organization email</label>
    <input id="organization-email" type="email" value={organization.email} disabled />
  </div>

  <div class="field">
    <label for="organization-description">Description</label>
    <textarea
      id="organization-description"
      maxlength={DESCRIPTION_LIMIT}
      disabled={!isOwner}
      class:invalid={fieldErrors.description}
      bind:value={description}
    ></textarea>
    {#if fieldErrors.description}<span class="field-error">{fieldErrors.description}</span>{/if}
  </div>

  {#if isOwner}
    <div class="form-actions">
      <button type="submit" class="button-primary" disabled={saving}>
        {saving ? "Saving…" : "Save changes"}
      </button>
    </div>
  {/if}
</form>
