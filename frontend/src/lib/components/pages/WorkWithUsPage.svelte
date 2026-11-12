<script lang="ts">
  import "../../../styles/work-with-us.sass";
  import { HourglassTop } from "@material-symbols-svg/svelte";
  import SiteHeader from "../layout/SiteHeader.svelte";
  import EmailStep from "../auth/EmailStep.svelte";
  import OtpStep from "../auth/OtpStep.svelte";
  import ApplicationForm from "../organizer/ApplicationForm.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import StatusBadge from "../ui/StatusBadge.svelte";
  import { extractErrorMessage } from "../../api/client";
  import {
    createMyApplicationsQuery,
    myApplicationsQueryKey,
    requestOrganizationEmailOtp,
    verifyOrganizationEmailOtp,
  } from "../../api/organizerApplications";
  import { formatDate } from "../../format";
  import { queryClient } from "../../queryClient";
  import { refreshSession, user } from "../../stores/user";
  import type { OrganizerApplication, OrganizerApplicationList } from "../../types/organizerApplication";

  const applications = createMyApplicationsQuery();

  let step = $state<"email" | "otp" | "form">("email");
  let email = $state("");
  let verificationToken = $state("");
  let noticeMessage = $state("");
  let sessionRefreshed = false;

  let pendingApplication = $derived(
    applications.data?.applications.find((application) => application.status === "PENDING"),
  );
  let hasApproval = $derived(
    applications.data?.applications.some((application) => application.status === "APPROVED") ?? false,
  );

  $effect(() => {
    if (hasApproval && $user?.role === "USER" && !sessionRefreshed) {
      sessionRefreshed = true;
      refreshSession();
    }
  });

  function refetchApplications() {
    return queryClient.invalidateQueries({ queryKey: myApplicationsQueryKey });
  }

  async function submitEmail(enteredEmail: string) {
    await requestOrganizationEmailOtp(enteredEmail);
    email = enteredEmail;
    noticeMessage = "";
    step = "otp";
  }

  async function verifyOtp(otp: string) {
    verificationToken = await verifyOrganizationEmailOtp(email, otp);
    step = "form";
  }

  function handleSubmitted(application: OrganizerApplication) {
    queryClient.setQueryData<OrganizerApplicationList>(myApplicationsQueryKey, (current) => ({
      applications: [application, ...(current?.applications ?? [])],
      total: (current?.total ?? 0) + 1,
    }));
    step = "email";
    verificationToken = "";
    refetchApplications();
  }

  function handleExpired(message: string) {
    noticeMessage = `${message} Please verify your organization email again.`;
    verificationToken = "";
    step = "email";
  }
</script>

<div class="work-root">
  <SiteHeader />

  <main class="page-content work-content">
    <div class="work-intro">
      <h1>Work with us</h1>
      <p>
        Sell tickets to your events on debyut. Verify your organization's email, tell us about it, and our team
        will review your application.
      </p>
    </div>

    <div class="work-grid">
      <section class="work-panel">
        {#if applications.isPending}
          <div class="state"><Spinner size={32} /></div>
        {:else if pendingApplication}
          <div class="auth-heading">
            <h1>Application under review</h1>
            <p>We'll email <strong>{pendingApplication.organizationEmail}</strong> once it has been reviewed.</p>
          </div>
          <div class="auth-card pending-card">
            <span class="pending-icon"><HourglassTop size={28} /></span>
            <div class="pending-text">
              <span class="item-title">{pendingApplication.organizationName}</span>
              <span class="item-detail">Submitted {formatDate(pendingApplication.createdAt)}</span>
            </div>
            <StatusBadge status={pendingApplication.status} />
          </div>
        {:else if step === "email"}
          {#if noticeMessage}
            <div class="alert-error notice" role="alert">{noticeMessage}</div>
          {/if}
          <EmailStep
            title="Verify your organization"
            label="Organization email"
            submitLabel="Send code"
            {email}
            onSubmit={submitEmail}
          >
            {#snippet subtitle()}
              We'll send a verification code to your organization's email address.
            {/snippet}
          </EmailStep>
        {:else if step === "otp"}
          <OtpStep
            {email}
            onVerify={verifyOtp}
            onResend={() => requestOrganizationEmailOtp(email)}
            onBack={() => (step = "email")}
          />
        {:else}
          <ApplicationForm
            {email}
            {verificationToken}
            onSubmitted={handleSubmitted}
            onExpired={handleExpired}
            onConflict={refetchApplications}
            onBack={() => (step = "email")}
          />
        {/if}
      </section>

      <aside class="work-history">
        <h2 class="section-title">My applications</h2>
        {#if applications.isPending}
          <div class="state"><Spinner size={24} /></div>
        {:else if applications.isError}
          <div class="state">
            <p>{extractErrorMessage(applications.error)}</p>
            <button type="button" class="button-outline" onclick={() => applications.refetch()}>Try again</button>
          </div>
        {:else if applications.data.applications.length === 0}
          <div class="state"><p>You haven't applied yet.</p></div>
        {:else}
          <ul class="item-list">
            {#each applications.data.applications as application (application.id)}
              <li class="item-row history-row">
                <div class="item-text">
                  <span class="item-title">{application.organizationName}</span>
                  <span class="item-detail">{application.organizationEmail}</span>
                  <span class="item-detail">
                    Submitted {formatDate(application.createdAt)}
                    {#if application.reviewedAt}· Reviewed {formatDate(application.reviewedAt)}{/if}
                  </span>
                  {#if application.status === "APPROVED" && application.organizationId}
                    <a href={`#/organizer/${application.organizationId}`} class="history-link">Open organizer dashboard</a>
                  {/if}
                </div>
                <StatusBadge status={application.status} />
              </li>
            {/each}
          </ul>
        {/if}
      </aside>
    </div>
  </main>
</div>
