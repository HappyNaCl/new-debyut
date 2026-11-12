<script lang="ts">
  import "../../../styles/admin.sass";
  import { Apartment, AssignmentInd, Category } from "@material-symbols-svg/svelte";
  import AdminLayout from "../admin/AdminLayout.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import StatusBadge from "../ui/StatusBadge.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { createApplicationsQuery, createOrganizationsQuery } from "../../api/admin";
  import { createCategoriesQuery } from "../../api/categories";
  import { formatDate } from "../../format";
  import { user } from "../../stores/user";

  const RECENT_LIMIT = 5;

  const pendingApplications = createApplicationsQuery(() => ({ status: "PENDING", page: 1 }));
  const organizations = createOrganizationsQuery(() => ({ search: "", page: 1 }));
  const categories = createCategoriesQuery();
</script>

<div class="admin-root">
  <AdminLayout title="Dashboard">
    <div class="page-heading">
      <div>
        <p>Welcome back{$user ? `, ${$user.firstName}` : ""}. Here's what needs your attention.</p>
      </div>
    </div>

    <div class="stat-grid">
      <a href="#/admin/organizer-applications" class="stat-card">
        <span class="item-icon"><AssignmentInd size={22} /></span>
        <span class="stat-text">
          <span class="stat-value">{pendingApplications.data?.total ?? "–"}</span>
          <span class="stat-label">Pending applications</span>
        </span>
      </a>
      <a href="#/admin/organizations" class="stat-card">
        <span class="item-icon"><Apartment size={22} /></span>
        <span class="stat-text">
          <span class="stat-value">{organizations.data?.total ?? "–"}</span>
          <span class="stat-label">Organizations</span>
        </span>
      </a>
      <a href="#/admin/categories" class="stat-card">
        <span class="item-icon"><Category size={22} /></span>
        <span class="stat-text">
          <span class="stat-value">{categories.data?.length ?? "–"}</span>
          <span class="stat-label">Event categories</span>
        </span>
      </a>
    </div>

    <div class="section-heading">
      <h2 class="section-title">Waiting for review</h2>
      <a href="#/admin/organizer-applications">Review all</a>
    </div>

    {#if pendingApplications.isPending}
      <div class="state"><Spinner size={32} /></div>
    {:else if pendingApplications.isError}
      <div class="state">
        <p>{extractErrorMessage(pendingApplications.error)}</p>
        <button type="button" class="button-outline" onclick={() => pendingApplications.refetch()}>Try again</button>
      </div>
    {:else if pendingApplications.data.applications.length === 0}
      <div class="state"><p>No applications waiting for review.</p></div>
    {:else}
      <ul class="item-list">
        {#each pendingApplications.data.applications.slice(0, RECENT_LIMIT) as application (application.id)}
          <li class="item-row">
            <span class="item-icon"><AssignmentInd size={22} /></span>
            <div class="item-text">
              <span class="item-title">{application.organizationName}</span>
              <span class="item-detail">
                {application.organizationEmail} · Submitted {formatDate(application.createdAt)}
              </span>
            </div>
            <StatusBadge status={application.status} />
          </li>
        {/each}
      </ul>
    {/if}
  </AdminLayout>
</div>
