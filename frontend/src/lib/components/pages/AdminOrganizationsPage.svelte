<script lang="ts">
  import "../../../styles/admin.sass";
  import { Group, Search } from "@material-symbols-svg/svelte";
  import AdminLayout from "../admin/AdminLayout.svelte";
  import Avatar from "../ui/Avatar.svelte";
  import Pagination from "../ui/Pagination.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { ADMIN_PAGE_SIZE, createOrganizationsQuery } from "../../api/admin";
  import { formatDate } from "../../format";

  let searchInput = $state("");
  let search = $state("");
  let page = $state(1);

  const organizations = createOrganizationsQuery(() => ({ search, page }));

  function handleSearch(event: SubmitEvent) {
    event.preventDefault();
    search = searchInput.trim();
    page = 1;
  }
</script>

<div class="admin-root">
  <AdminLayout title="Organizations">
    <div class="page-heading">
      <div>
        <p>Every organization on debyut.</p>
      </div>
    </div>

    <form class="search-form" onsubmit={handleSearch} role="search">
      <input
        type="search"
        class="text-input"
        maxlength="150"
        placeholder="Search by name or email"
        aria-label="Search organizations"
        bind:value={searchInput}
      />
      <button type="submit" class="button-outline"><Search size={18} /> Search</button>
    </form>

    {#if organizations.isPending}
      <div class="state"><Spinner size={32} /></div>
    {:else if organizations.isError}
      <div class="state">
        <p>{extractErrorMessage(organizations.error)}</p>
        <button type="button" class="button-outline" onclick={() => organizations.refetch()}>Try again</button>
      </div>
    {:else if organizations.data.organizations.length === 0}
      <div class="state">
        <p>{search ? `No organizations match “${search}”.` : "No organizations yet."}</p>
      </div>
    {:else}
      <ul class="item-list">
        {#each organizations.data.organizations as organization (organization.id)}
          <li class="item-row">
            <Avatar name={organization.name} src={organization.avatarUrl} shape="rounded" />
            <div class="item-text">
              <span class="item-title">{organization.name}</span>
              <span class="item-detail">{organization.email} · Since {formatDate(organization.createdAt)}</span>
            </div>
            <span class="member-count" title="Members">
              <Group size={18} /> {organization.memberCount ?? 0}
            </span>
          </li>
        {/each}
      </ul>
      <Pagination {page} pageSize={ADMIN_PAGE_SIZE} total={organizations.data.total} onChange={(next) => (page = next)} />
    {/if}
  </AdminLayout>
</div>
