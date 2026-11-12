<script lang="ts">
  import "../../../styles/auth.sass";
  import { push, router } from "svelte-spa-router";
  import AuthShell from "../auth/AuthShell.svelte";
  import EmailStep from "../auth/EmailStep.svelte";
  import OtpStep from "../auth/OtpStep.svelte";
  import ProfileForm from "../auth/ProfileForm.svelte";
  import { requestRegisterOtp, verifyRegisterOtp } from "../../api/auth";
  import { setAccessToken } from "../../stores/tokens";
  import { loadUser } from "../../stores/user";
  import { readRedirect, withRedirect } from "../../redirect";
  import type { AuthTokens } from "../../types/auth";

  const PROFILE_LOAD_ATTEMPTS = 5;

  const redirectTarget = readRedirect(router.querystring);

  let step = $state<"email" | "otp" | "profile">("email");
  let email = $state("");
  let registrationToken = $state("");

  async function submitEmail(enteredEmail: string) {
    await requestRegisterOtp({ email: enteredEmail });
    email = enteredEmail;
    step = "otp";
  }

  async function verifyOtp(otp: string) {
    const response = await verifyRegisterOtp({ email, otp });
    registrationToken = response.data.data.registrationToken;
    step = "profile";
  }

  async function finishRegistration(tokens: AuthTokens) {
    setAccessToken(tokens.accessToken);
    await loadUser(true, PROFILE_LOAD_ATTEMPTS).catch(() => {});
    push(redirectTarget ?? "/");
  }
</script>

<div class="auth-root">
  <AuthShell>
    {#if step === "email"}
      <EmailStep title="Create your debyut account" submitLabel="Register" {email} onSubmit={submitEmail}>
        {#snippet subtitle()}
          Already have an account? <a href={`#${withRedirect("/login", redirectTarget)}`}>Log in</a>
        {/snippet}
      </EmailStep>
    {:else if step === "otp"}
      <OtpStep
        {email}
        onVerify={verifyOtp}
        onResend={() => requestRegisterOtp({ email }).then(() => {})}
        onBack={() => (step = "email")}
      />
    {:else}
      <ProfileForm {email} {registrationToken} onComplete={finishRegistration} />
    {/if}
  </AuthShell>
</div>
