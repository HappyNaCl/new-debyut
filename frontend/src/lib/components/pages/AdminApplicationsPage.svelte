<script lang="ts">
  import "../../../styles/admin.sass";
  import axios from "axios";
  import AdminLayout from "../admin/AdminLayout.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Pagination from "../ui/Pagination.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import StatusBadge from "../ui/StatusBadge.svelte";
  import { extractErrorMessage } from "../../api/client";
  import {
    ADMIN_PAGE_SIZE,
    adminApplicationsQueryKey,
    adminOrganizationsQueryKey,
    approveApplication,
    createApplicationsQuery,
    declineApplication,
  } from "../../api/admin";
  import { formatDate } from "../../format";
  import { queryClient } from "../../queryClient";
  import type { ApplicationStatus, OrganizerApplication } from "../../types/organizerApplication";

  const filters: { value: ApplicationStatus | undefined; label: string }[] = [
    { value: "PENDING", label: "Pending" },
    { value: "APPROVED", label: "Approved" },
    { value: "DECLINED", label: "Declined" },
    { value: undefined, label: "All" },
  ];

  let status = $state<ApplicationStatus | undefined>("PENDING");
  let page = $state(1);
  let reviewingId = $state<string | null>(null);
  let errorMessage = $state("");
  let successMessage = $state("");

  const applications = createApplicationsQuery(() => ({ status, page }));

  function selectStatus(next: ApplicationStatus | undefined) {
    status = next;
    page = 1;
  }

  async function review(application: OrganizerApplication, decision: "approve" | "decline") {
    errorMessage = "";
    successMessage = "";
    reviewingId = application.id;
    try {
      if (decision === "approve") {
        await approveApplication(application.id);
        successMessage = `Approved ${application.organizationName}.`;
      } else {
        await declineApplication(application.id);
        successMessage = `Declined ${application.organizationName}.`;
      }
    } catch (error) {
      const alreadyReviewed = axios.isAxiosError(error) && error.response?.status === 400;
      errorMessage = alreadyReviewed
        ? `${application.organizationName} has already been reviewed.`
        : extractErrorMessage(error);
    } finally {
      reviewingId = null;
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: adminApplicationsQueryKey }),
        queryClient.invalidateQueries({ queryKey: adminOrganizationsQueryKey }),
      ]);
    }
  }
</script>

<div class="admin-root">
  <AdminLayout title="Organizer applications">
    <div class="page-heading">
      <div>
        <p>Approving an application creates the organization and makes the applicant its owner.</p>
      </div>
    </div>

    <div class="filter-tabs" role="tablist">
      {#each filters as filter (filter.label)}
        <button
          type="button"
          role="tab"
          class="filter-tab"
          class:active={status === filter.value}
          aria-selected={status === filter.value}
          onclick={() => selectStatus(filter.value)}
        >
          {filter.label}
        </button>
      {/each}
    </div>

    {#if successMessage}
      <div class="alert alert-success" role="status">{successMessage}</div>
    {/if}
    {#if errorMessage}
      <div class="alert alert-error" role="alert">{errorMessage}</div>
    {/if}

    {#if applications.isPending}
      <div class="state"><Spinner size={32} /></div>
    {:else if applications.isError}
      <div class="state">
        <p>{extractErrorMessage(applications.error)}</p>
        <button type="button" class="button-outline" onclick={() => applications.refetch()}>Try again</button>
      </div>
    {:else if applications.data.applications.length === 0}
      <div class="state"><p>No applications here.</p></div>
    {:else}
      <ul class="item-list">
        {#each applications.data.applications as application (application.id)}
          <li class="item-row application-row">
            <Avatar name={application.organizationName} src={application.avatarUrl} shape="rounded" size={44} />
            <div class="item-text">
              <span class="item-title">{application.organizationName}</span>
              <span class="item-detail">{application.organizationEmail}</span>
              <p class="application-description">{application.description}</p>
              <span class="item-meta">
                <StatusBadge status={application.status} />
                Submitted {formatDate(application.createdAt)}
                {#if application.reviewedAt}· Reviewed {formatDate(application.reviewedAt)}{/if}
              </span>
            </div>
            {#if application.status === "PENDING"}
              <div class="row-actions">
                <button
                  type="button"
                  class="button-outline"
                  onclick={() => review(application, "decline")}
                  disabled={reviewingId !== null}
                >
                  Decline
                </button>
                <button
                  type="button"
                  class="button-primary"
                  onclick={() => review(application, "approve")}
                  disabled={reviewingId !== null}
                >
                  {reviewingId === application.id ? "Saving…" : "Approve"}
                </button>
              </div>
            {/if}
          </li>
        {/each}
      </ul>
      <Pagination {page} pageSize={ADMIN_PAGE_SIZE} total={applications.data.total} onChange={(next) => (page = next)} />
    {/if}
  </AdminLayout>
</div>
