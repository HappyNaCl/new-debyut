<script lang="ts">
  import { onMount } from "svelte";
  import { ArrowBack } from "@material-symbols-svg/svelte";
  import OtpInput from "../ui/OtpInput.svelte";
  import { extractErrorMessage } from "../../api/client";

  interface Props {
    email: string;
    onVerify: (otp: string) => Promise<void>;
    onResend: () => Promise<void>;
    onBack: () => void;
  }

  let { email, onVerify, onResend, onBack }: Props = $props();

  const RESEND_COOLDOWN_SECONDS = 60;
  const OTP_LENGTH = 6;

  let otp = $state("");
  let verifying = $state(false);
  let resending = $state(false);
  let errorMessage = $state("");
  let infoMessage = $state("");
  let secondsLeft = $state(RESEND_COOLDOWN_SECONDS);
  let otpInput: OtpInput | undefined = $state();

  onMount(() => {
    otpInput?.focus();
    const timer = setInterval(() => {
      if (secondsLeft > 0) secondsLeft -= 1;
    }, 1000);
    return () => clearInterval(timer);
  });

  async function verify() {
    if (otp.length !== OTP_LENGTH || verifying) return;
    errorMessage = "";
    infoMessage = "";
    verifying = true;
    try {
      await onVerify(otp);
    } catch (error) {
      errorMessage = extractErrorMessage(error);
      otp = "";
      otpInput?.focus();
    } finally {
      verifying = false;
    }
  }

  async function resend() {
    errorMessage = "";
    infoMessage = "";
    resending = true;
    try {
      await onResend();
      otp = "";
      secondsLeft = RESEND_COOLDOWN_SECONDS;
      infoMessage = "A new code has been sent.";
      otpInput?.focus();
    } catch (error) {
      errorMessage = extractErrorMessage(error);
    } finally {
      resending = false;
    }
  }

  function handleSubmit(event: SubmitEvent) {
    event.preventDefault();
    verify();
  }
</script>

<div class="auth-heading">
  <h1>Enter verification code</h1>
  <p>We sent a {OTP_LENGTH}-digit code to <strong>{email}</strong></p>
</div>

<form class="auth-card" onsubmit={handleSubmit}>
  <button type="button" class="back-button" onclick={onBack}>
    <ArrowBack size={18} /> Change email
  </button>

  {#if errorMessage}
    <div class="alert-error" role="alert">{errorMessage}</div>
  {/if}
  {#if infoMessage}
    <div class="alert-info" role="status">{infoMessage}</div>
  {/if}

  <OtpInput bind:this={otpInput} bind:value={otp} length={OTP_LENGTH} disabled={verifying} onComplete={verify} />

  <button type="submit" class="button-primary" disabled={verifying || otp.length !== OTP_LENGTH}>
    {verifying ? "Verifying…" : "Verify"}
  </button>

  <p class="resend">
    Didn't receive the code?
    {#if secondsLeft > 0}
      <span class="muted">Resend in {secondsLeft}s</span>
    {:else}
      <button type="button" class="link-button" onclick={resend} disabled={resending}>
        {resending ? "Sending…" : "Resend code"}
      </button>
    {/if}
  </p>
</form>
