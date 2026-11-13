<script lang="ts">
  import type { Snippet } from "svelte";
  import { replace, router } from "svelte-spa-router";
  import {
    AccountBalance,
    Apartment,
    Badge,
    Dashboard,
    EventNote,
    Gavel,
    Group,
    Settings,
  } from "@material-symbols-svg/svelte";
  import DashboardShell from "../layout/DashboardShell.svelte";
  import Spinner from "../ui/Spinner.svelte";
  import { extractErrorMessage } from "../../api/client";
  import { createMyOrganizationsQuery, createOrganizationQuery } from "../../api/organizations";
  import type { DashboardLink, DashboardSection, DashboardSwitchOption } from "../../types/dashboard";
  import type { Organization } from "../../types/organization";

  interface Props {
    title: string;
    children: Snippet<[Organization]>;
  }

  let { title, children }: Props = $props();

  const roleLabels = { OWNER: "Owner", STAFF: "Staff" } as const;

  const organizations = createMyOrganizationsQuery();

  let routeId = $derived((router.params as Record<string, string> | null | undefined)?.id);
  let selectedId = $derived(routeId ?? organizations.data?.[0]?.id);
  let subPath = $derived(routeId ? router.location.slice(`/organizer/${routeId}`.length) : "");

  const organization = createOrganizationQuery(() => selectedId);

  $effect(() => {
    if (!routeId && selectedId) replace(`/organizer/${selectedId}`);
  });

  let current = $derived(organizations.data?.find((option) => option.id === selectedId));
  let myRole = $derived(organization.data?.myRole ?? current?.myRole);

  let sections = $derived.by<DashboardSection[]>(() => {
    if (!selectedId) return [];

    const base = `/organizer/${selectedId}`;
    const access: DashboardLink[] = [{ href: `${base}/members`, label: "Members", icon: Group }];

    return [
      {
        label: "Dashboard",
        links: [
          { href: base, label: "Dashboard", icon: Dashboard },
          { href: `${base}/events`, label: "My events", icon: EventNote },
        ],
      },
      { label: "Manage access", links: access },
      {
        label: "Account",
        links: [
          { href: `${base}/profile`, label: "Basic information", icon: Badge },
          { href: `${base}/legal`, label: "Legal information", icon: Gavel },
          { href: `${base}/settings`, label: "Settings", icon: Settings },
          { href: `${base}/bank-account`, label: "Account number", icon: AccountBalance },
        ],
      },
    ];
  });

  let switchOptions = $derived<DashboardSwitchOption[]>(
    (organizations.data ?? []).map((option) => {
      return {
        href: `/organizer/${option.id}${subPath}`,
        label: option.name,
        caption: option.myRole ? roleLabels[option.myRole] : undefined,
        current: option.id === selectedId,
      };
    }),
  );
</script>

<DashboardShell
  {title}
  area="Organizer"
  profileName={organization.data?.name ?? current?.name ?? "Organizer"}
  profileCaption={myRole ? roleLabels[myRole] : "Organizer"}
  profileAvatarUrl={organization.data?.avatarUrl ?? current?.avatarUrl}
  {sections}
  {switchOptions}
>
  {#if organizations.isPending || (selectedId && organization.isPending)}
    <div class="state"><Spinner size={32} /></div>
  {:else if organizations.isError}
    <div class="state">
      <p>{extractErrorMessage(organizations.error)}</p>
      <button type="button" class="button-outline" onclick={() => organizations.refetch()}>Try again</button>
    </div>
  {:else if !selectedId}
    <div class="state">
      <Apartment size={40} />
      <p>You're not part of any organization yet.</p>
      <a href="#/work-with-us" class="button-primary">Work with us</a>
    </div>
  {:else if organization.isError}
    <div class="state">
      <p>{extractErrorMessage(organization.error)}</p>
      <a href="#/organizer" class="button-outline">Back to my organizations</a>
    </div>
  {:else if organization.data}
    {@render children(organization.data)}
  {/if}
</DashboardShell>
