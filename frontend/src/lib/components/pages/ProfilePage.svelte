<script lang="ts">
  import "../../../styles/profile.sass";
  import AccountLayout from "../account/AccountLayout.svelte";
  import AvatarUploader from "../ui/AvatarUploader.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import { updateMe, uploadMyAvatar } from "../../api/user";
  import {
    isValidPhoneNumber,
    latestDateOfBirth,
    MINIMUM_AGE,
    PHONE_PREFIX,
    toPhoneDigits,
    toPhoneNumber,
  } from "../../profile";
  import { fullName } from "../../format";
  import { loadUser, user } from "../../stores/user";
  import type { Gender } from "../../types/auth";

  const genderOptions: { value: Gender; label: string }[] = [
    { value: "MALE", label: "Male" },
    { value: "FEMALE", label: "Female" },
  ];

  const maximumDateOfBirth = latestDateOfBirth();

  loadUser().catch(() => {});

  let loadedUserId = $state<string | null>(null);
  let firstName = $state("");
  let surname = $state("");
  let dateOfBirth = $state("");
  let phoneDigits = $state("");
  let gender = $state<Gender | "">("");

  let saving = $state(false);
  let errorMessage = $state("");
  let successMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  $effect(() => {
    const profile = $user;
    if (!profile || loadedUserId === profile.id) return;
    loadedUserId = profile.id;
    firstName = profile.firstName;
    surname = profile.surname;
    dateOfBirth = profile.dateOfBirth;
    phoneDigits = toPhoneDigits(profile.phoneNumber);
    gender = profile.gender;
  });

  async function handleAvatarUpload(file: File) {
    user.set(await uploadMyAvatar(file));
  }

  function validate(): Record<string, string> {
    const errors: Record<string, string> = {};
    if (!firstName.trim()) errors.firstName = "First name is required.";
    if (!dateOfBirth) errors.dateOfBirth = "Date of birth is required.";
    else if (dateOfBirth > maximumDateOfBirth) errors.dateOfBirth = `You must be at least ${MINIMUM_AGE} years old.`;
    if (!isValidPhoneNumber(toPhoneNumber(phoneDigits))) errors.phoneNumber = "Enter a valid phone number.";
    if (!gender) errors.gender = "Please select your gender.";
    return errors;
  }

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    successMessage = "";
    fieldErrors = validate();
    if (Object.keys(fieldErrors).length > 0 || !$user) return;

    saving = true;
    try {
      const updated = await updateMe({
        firstName: firstName.trim(),
        surname: surname.trim(),
        dateOfBirth,
        phoneNumber: toPhoneNumber(phoneDigits),
        gender: gender as Gender,
        subscribeNewsletter: $user.subscribeNewsletter,
      });
      user.set(updated);
      successMessage = "Your profile has been saved.";
    } catch (error) {
      const serverFieldErrors = extractFieldErrors(error);
      if (Object.keys(serverFieldErrors).length > 0) fieldErrors = serverFieldErrors;
      else errorMessage = extractErrorMessage(error);
    } finally {
      saving = false;
    }
  }
</script>

<div class="profile-root">
  <AccountLayout title="Basic information">
    <div class="page-heading">
      <div>
        <p>Keep your details up to date so organizers know who's coming.</p>
      </div>
    </div>

    {#if !loadedUserId}
      <div class="state"><Spinner size={32} /></div>
    {:else}
      <form class="card dashboard-form" onsubmit={handleSubmit} novalidate>
        {#if successMessage}
          <div class="alert alert-success" role="status">{successMessage}</div>
        {/if}
        {#if errorMessage}
          <div class="alert alert-error" role="alert">{errorMessage}</div>
        {/if}

        <AvatarUploader
          name={$user ? fullName($user) : firstName}
          src={$user?.avatarUrl}
          onUpload={handleAvatarUpload}
        />

        <div class="field-row">
          <div class="field">
            <label for="profile-first-name">First name</label>
            <input
              id="profile-first-name"
              autocomplete="given-name"
              maxlength="100"
              class:invalid={fieldErrors.firstName}
              bind:value={firstName}
            />
            {#if fieldErrors.firstName}<span class="field-error">{fieldErrors.firstName}</span>{/if}
          </div>

          <div class="field">
            <label for="profile-surname">Surname <span class="optional">(optional)</span></label>
            <input
              id="profile-surname"
              autocomplete="family-name"
              maxlength="100"
              class:invalid={fieldErrors.surname}
              bind:value={surname}
            />
            {#if fieldErrors.surname}<span class="field-error">{fieldErrors.surname}</span>{/if}
          </div>
        </div>

        <div class="field">
          <label for="profile-date-of-birth">Date of birth</label>
          <input
            id="profile-date-of-birth"
            type="date"
            autocomplete="bday"
            max={maximumDateOfBirth}
            class:invalid={fieldErrors.dateOfBirth}
            bind:value={dateOfBirth}
          />
          {#if fieldErrors.dateOfBirth}<span class="field-error">{fieldErrors.dateOfBirth}</span>{/if}
        </div>

        <div class="field">
          <label for="profile-phone">Phone number</label>
          <div class="phone-input" class:invalid={fieldErrors.phoneNumber}>
            <span class="phone-prefix">{PHONE_PREFIX}</span>
            <input
              id="profile-phone"
              type="tel"
              inputmode="numeric"
              autocomplete="tel-national"
              placeholder="81234567890"
              bind:value={phoneDigits}
            />
          </div>
          {#if fieldErrors.phoneNumber}<span class="field-error">{fieldErrors.phoneNumber}</span>{/if}
        </div>

        <fieldset class="field">
          <legend>Gender</legend>
          <div class="choice-group">
            {#each genderOptions as option (option.value)}
              <label class="choice" class:selected={gender === option.value}>
                <input type="radio" name="gender" value={option.value} bind:group={gender} />
                {option.label}
              </label>
            {/each}
          </div>
          {#if fieldErrors.gender}<span class="field-error">{fieldErrors.gender}</span>{/if}
        </fieldset>

        <div class="form-actions">
          <button type="submit" class="button-primary" disabled={saving}>
            {saving ? "Saving…" : "Save changes"}
          </button>
        </div>
      </form>
    {/if}
  </AccountLayout>
</div>
