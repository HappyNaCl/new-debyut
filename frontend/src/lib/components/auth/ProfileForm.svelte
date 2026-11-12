<script lang="ts">
  import { completeRegistration } from "../../api/auth";
  import { extractErrorMessage, extractFieldErrors } from "../../api/client";
  import {
    isValidPhoneNumber,
    latestDateOfBirth as computeLatestDateOfBirth,
    MINIMUM_AGE,
    PHONE_PREFIX,
    toPhoneNumber,
  } from "../../profile";
  import type { AuthTokens, Gender } from "../../types/auth";

  interface Props {
    email: string;
    registrationToken: string;
    onComplete: (tokens: AuthTokens) => Promise<void>;
  }

  let { email, registrationToken, onComplete }: Props = $props();

  const genderOptions: { value: Gender; label: string }[] = [
    { value: "MALE", label: "Male" },
    { value: "FEMALE", label: "Female" },
  ];

  let firstName = $state("");
  let surname = $state("");
  let dateOfBirth = $state("");
  let phoneDigits = $state("");
  let gender = $state<Gender | "">("");
  let agreeToTerms = $state(false);
  let subscribeNewsletter = $state(false);

  let loading = $state(false);
  let errorMessage = $state("");
  let fieldErrors = $state<Record<string, string>>({});

  const latestDateOfBirth = computeLatestDateOfBirth();

  function normalizedPhoneNumber() {
    return toPhoneNumber(phoneDigits);
  }

  function validate(): Record<string, string> {
    const errors: Record<string, string> = {};
    if (!firstName.trim()) errors.firstName = "First name is required.";
    if (!dateOfBirth) errors.dateOfBirth = "Date of birth is required.";
    else if (dateOfBirth > latestDateOfBirth) errors.dateOfBirth = `You must be at least ${MINIMUM_AGE} years old.`;
    if (!isValidPhoneNumber(normalizedPhoneNumber())) errors.phoneNumber = "Enter a valid phone number.";
    if (!gender) errors.gender = "Please select your gender.";
    if (!agreeToTerms) errors.agreeToTerms = "You must agree to the Terms of Service.";
    return errors;
  }

  let attemptedSubmit = $state(false);

  $effect(() => {
    if (attemptedSubmit) fieldErrors = validate();
  });

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    attemptedSubmit = true;
    fieldErrors = validate();
    if (Object.keys(fieldErrors).length > 0) return;

    loading = true;
    try {
      const response = await completeRegistration({
        registrationToken,
        firstName: firstName.trim(),
        surname: surname.trim(),
        dateOfBirth,
        phoneNumber: normalizedPhoneNumber(),
        gender: gender as Gender,
        agreeToTerms,
        subscribeNewsletter,
      });
      await onComplete(response.data.data);
    } catch (error) {
      const serverFieldErrors = extractFieldErrors(error);
      if (Object.keys(serverFieldErrors).length > 0) fieldErrors = serverFieldErrors;
      else errorMessage = extractErrorMessage(error);
    } finally {
      loading = false;
    }
  }
</script>

<div class="auth-heading">
  <h1>Complete your profile</h1>
  <p>Almost done! Tell us a bit about yourself.</p>
</div>

<form class="auth-card" onsubmit={handleSubmit} novalidate>
  {#if errorMessage}
    <div class="alert-error" role="alert">{errorMessage}</div>
  {/if}

  <div class="field">
    <label for="profile-email">Email</label>
    <input id="profile-email" type="email" value={email} disabled />
  </div>

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
      max={latestDateOfBirth}
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

  <div class="checkboxes">
    <label class="checkbox">
      <input type="checkbox" bind:checked={agreeToTerms} />
      <span>I agree to the <strong>Terms of Service</strong> and <strong>Privacy Policy</strong> of debyut.</span>
    </label>
    {#if fieldErrors.agreeToTerms}<span class="field-error">{fieldErrors.agreeToTerms}</span>{/if}

    <label class="checkbox">
      <input type="checkbox" bind:checked={subscribeNewsletter} />
      <span>Send me news about the latest events and promotions.</span>
    </label>
  </div>

  <button type="submit" class="button-primary" disabled={loading}>
    {loading ? "Creating account…" : "Create account"}
  </button>
</form>
