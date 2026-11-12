<script lang="ts">
  import type { Snippet } from "svelte";
  import { Apartment, AssignmentInd, Category, Dashboard } from "@material-symbols-svg/svelte";
  import DashboardShell from "../layout/DashboardShell.svelte";
  import { createApplicationsQuery } from "../../api/admin";
  import type { DashboardSection } from "../../types/dashboard";

  interface Props {
    title: string;
    children: Snippet;
  }

  let { title, children }: Props = $props();

  const pendingApplications = createApplicationsQuery(() => ({ status: "PENDING", page: 1 }));

  let sections = $derived<DashboardSection[]>([
    {
      label: "Overview",
      links: [{ href: "/admin", label: "Dashboard", icon: Dashboard }],
    },
    {
      label: "Events",
      links: [{ href: "/admin/categories", label: "Categories", icon: Category }],
    },
    {
      label: "Organizers",
      links: [
        {
          href: "/admin/organizer-applications",
          label: "Applications",
          icon: AssignmentInd,
          badge: pendingApplications.data?.total,
        },
        { href: "/admin/organizations", label: "Organizations", icon: Apartment },
      ],
    },
  ]);
</script>

<DashboardShell {title} area="Admin" {sections}>
  {@render children()}
</DashboardShell>
