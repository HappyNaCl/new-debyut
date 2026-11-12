<script lang="ts">
  import "../../../styles/auth.sass";
  import { push, router } from "svelte-spa-router";
  import AuthShell from "../auth/AuthShell.svelte";
  import EmailStep from "../auth/EmailStep.svelte";
  import OtpStep from "../auth/OtpStep.svelte";
  import { requestLoginOtp, verifyLoginOtp } from "../../api/auth";
  import { setAccessToken } from "../../stores/tokens";
  import { loadUser } from "../../stores/user";
  import { readRedirect, withRedirect } from "../../redirect";

  const redirectTarget = readRedirect(router.querystring);

  let step = $state<"email" | "otp">("email");
  let email = $state("");

  async function submitEmail(enteredEmail: string) {
    await requestLoginOtp({ email: enteredEmail });
    email = enteredEmail;
    step = "otp";
  }

  async function verifyOtp(otp: string) {
    const response = await verifyLoginOtp({ email, otp });
    setAccessToken(response.data.data.accessToken);
    await loadUser(true).catch(() => {});
    push(redirectTarget ?? "/");
  }
</script>

<div class="auth-root">
  <AuthShell>
    {#if step === "email"}
      <EmailStep title="Log in to your account" submitLabel="Log in" {email} onSubmit={submitEmail}>
        {#snippet subtitle()}
          Don't have an account? <a href={`#${withRedirect("/register", redirectTarget)}`}>Register</a>
        {/snippet}
      </EmailStep>
    {:else}
      <OtpStep
        {email}
        onVerify={verifyOtp}
        onResend={() => requestLoginOtp({ email }).then(() => {})}
        onBack={() => (step = "email")}
      />
    {/if}
  </AuthShell>
</div>
