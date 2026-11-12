<script lang="ts">
  import type { Snippet } from "svelte";
  import { Badge, ConfirmationNumber, Mail, Settings, VideoLibrary } from "@material-symbols-svg/svelte";
  import DashboardShell from "../layout/DashboardShell.svelte";
  import { createMyInvitationsQuery } from "../../api/invitations";
  import { user } from "../../stores/user";
  import type { DashboardSection } from "../../types/dashboard";

  interface Props {
    title: string;
    children: Snippet;
  }

  let { title, children }: Props = $props();

  const invitations = createMyInvitationsQuery(() => !!$user && $user.role !== "ADMIN");

  let sections = $derived<DashboardSection[]>([
    {
      label: "Activity",
      links: [
        { href: "/tickets", label: "My tickets", icon: ConfirmationNumber },
        { href: "/vod", label: "VOD", icon: VideoLibrary },
        { href: "/invitations", label: "Invitations", icon: Mail, badge: invitations.data?.length },
      ],
    },
    {
      label: "Account",
      links: [
        { href: "/profile", label: "Basic information", icon: Badge },
        { href: "/settings", label: "Settings", icon: Settings },
      ],
    },
  ]);
</script>

<DashboardShell {title} area="Account" {sections}>
  {@render children()}
</DashboardShell>
