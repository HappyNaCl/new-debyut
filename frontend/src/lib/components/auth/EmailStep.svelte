<script lang="ts">
  import { untrack, type Snippet } from "svelte";
  import { extractErrorMessage } from "../../api/client";

  interface Props {
    title: string;
    subtitle: Snippet;
    submitLabel: string;
    label?: string;
    email?: string;
    onSubmit: (email: string) => Promise<void>;
  }

  let { title, subtitle, submitLabel, label = "Email", email = "", onSubmit }: Props = $props();

  let emailValue = $state(untrack(() => email));
  let loading = $state(false);
  let errorMessage = $state("");

  async function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    errorMessage = "";
    loading = true;
    try {
      await onSubmit(emailValue.trim());
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      loading = false;
    }
  }
</script>

<div class="auth-heading">
  <h1>{title}</h1>
  <p>{@render subtitle()}</p>
</div>

<form class="auth-card" onsubmit={handleSubmit} novalidate>
  {#if errorMessage}
    <div class="alert-error" role="alert">{errorMessage}</div>
  {/if}

  <div class="field">
    <label for="auth-email">{label}</label>
    <input
      id="auth-email"
      type="email"
      autocomplete="email"
      required
      bind:value={emailValue}
    />
  </div>

  <button type="submit" class="button-primary" disabled={loading || !emailValue.trim()}>
    {loading ? "Please wait…" : submitLabel}
  </button>
</form>
