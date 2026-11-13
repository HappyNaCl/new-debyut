<script lang="ts">
  import { onMount } from "svelte";
  import { get } from "svelte/store";
  import Home from "./lib/components/pages/HomePage.svelte";
  import Login from "./lib/components/pages/LoginPage.svelte";
  import Register from "./lib/components/pages/RegisterPage.svelte";
  import Search from "./lib/components/pages/SearchPage.svelte";
  import AdminDashboard from "./lib/components/pages/AdminDashboardPage.svelte";
  import AdminCategories from "./lib/components/pages/AdminCategoriesPage.svelte";
  import AdminApplications from "./lib/components/pages/AdminApplicationsPage.svelte";
  import AdminOrganizations from "./lib/components/pages/AdminOrganizationsPage.svelte";
  import WorkWithUs from "./lib/components/pages/WorkWithUsPage.svelte";
  import OrganizerDashboard from "./lib/components/pages/OrganizerDashboardPage.svelte";
  import OrganizerEvents from "./lib/components/pages/OrganizerEventsPage.svelte";
  import OrganizerMembers from "./lib/components/pages/OrganizerMembersPage.svelte";
  import Invitations from "./lib/components/pages/InvitationsPage.svelte";
  import Tickets from "./lib/components/pages/TicketsPage.svelte";
  import Vod from "./lib/components/pages/VodPage.svelte";
  import Profile from "./lib/components/pages/ProfilePage.svelte";
  import Settings from "./lib/components/pages/SettingsPage.svelte";
  import OrganizerProfile from "./lib/components/pages/OrganizerProfilePage.svelte";
  import OrganizerLegal from "./lib/components/pages/OrganizerLegalPage.svelte";
  import OrganizerSettings from "./lib/components/pages/OrganizerSettingsPage.svelte";
  import OrganizerBankAccount from "./lib/components/pages/OrganizerBankAccountPage.svelte";
  import Spinner from "./lib/components/ui/Spinner.svelte";
  import ThemeToggle from "./lib/components/ui/ThemeToggle.svelte";
  import Router, { replace, router, type RouteDetail } from "svelte-spa-router";
  import wrap from "svelte-spa-router/wrap";
  import { QueryClientProvider } from "@tanstack/svelte-query";
  import { queryClient } from "./lib/queryClient";
  import { bootstrapAuth } from "./lib/api/client";
  import { authReady, tokens } from "./lib/stores/tokens";
  import { loadUser } from "./lib/stores/user";
  import { withRedirect } from "./lib/redirect";
  import { isDashboardRoute } from "./lib/dashboardRoutes";
  import type { Role } from "./lib/types/user";

  const signedOut = () => get(tokens).accessToken === null;
  const signedIn = () => get(tokens).accessToken !== null;

  async function currentRole(): Promise<Role | null> {
    if (get(tokens).accessToken === null) return null;
    const me = await loadUser().catch(() => null);
    return me?.role ?? null;
  }

  const isAdmin = async () => (await currentRole()) === "ADMIN";
  const isOrganizer = async () => (await currentRole()) === "ORGANIZER";
  const isNotAdmin = async () => (await currentRole()) !== "ADMIN";

  const signInRequired = { signInRequired: true };

  const routes = {
    "/": Home,
    "/search": Search,
    "/login": wrap({ component: Login, conditions: [signedOut] }),
    "/register": wrap({ component: Register, conditions: [signedOut] }),
    "/work-with-us": wrap({ component: WorkWithUs, conditions: [signedIn, isNotAdmin], userData: signInRequired }),
    "/invitations": wrap({ component: Invitations, conditions: [signedIn], userData: signInRequired }),
    "/tickets": wrap({ component: Tickets, conditions: [signedIn], userData: signInRequired }),
    "/vod": wrap({ component: Vod, conditions: [signedIn], userData: signInRequired }),
    "/profile": wrap({ component: Profile, conditions: [signedIn], userData: signInRequired }),
    "/settings": wrap({ component: Settings, conditions: [signedIn], userData: signInRequired }),
    "/organizer": wrap({ component: OrganizerDashboard, conditions: [isOrganizer] }),
    "/organizer/:id": wrap({ component: OrganizerDashboard, conditions: [isOrganizer] }),
    "/organizer/:id/events": wrap({ component: OrganizerEvents, conditions: [isOrganizer] }),
    "/organizer/:id/members": wrap({ component: OrganizerMembers, conditions: [isOrganizer] }),
    "/organizer/:id/invitations": wrap({ component: OrganizerMembers, conditions: [isOrganizer] }),
    "/organizer/:id/profile": wrap({ component: OrganizerProfile, conditions: [isOrganizer] }),
    "/organizer/:id/legal": wrap({ component: OrganizerLegal, conditions: [isOrganizer] }),
    "/organizer/:id/settings": wrap({ component: OrganizerSettings, conditions: [isOrganizer] }),
    "/organizer/:id/bank-account": wrap({ component: OrganizerBankAccount, conditions: [isOrganizer] }),
    "/admin": wrap({ component: AdminDashboard, conditions: [isAdmin] }),
    "/admin/categories": wrap({ component: AdminCategories, conditions: [isAdmin] }),
    "/admin/organizer-applications": wrap({ component: AdminApplications, conditions: [isAdmin] }),
    "/admin/organizations": wrap({ component: AdminOrganizations, conditions: [isAdmin] }),
  };

  function handleConditionsFailed(detail: RouteDetail) {
    const needsSignIn = (detail.userData as typeof signInRequired | undefined)?.signInRequired;
    if (needsSignIn && signedOut()) {
      const target = detail.querystring ? `${detail.location}?${detail.querystring}` : detail.location;
      replace(withRedirect("/login", target));
      return;
    }
    replace("/");
  }

  let inDashboard = $derived(isDashboardRoute(router.location));

  const SPINNER_DELAY_MS = 150;
  let showSpinner = $state(false);

  onMount(() => {
    const timer = setTimeout(() => {
      showSpinner = true;
    }, SPINNER_DELAY_MS);

    bootstrapAuth()
      .then(() => {
        if (get(tokens).accessToken) {
          return loadUser().catch(() => {});
        }
      })
      .finally(() => {
        clearTimeout(timer);
      });
  });
</script>

<svelte:head>
  <link rel="preconnect" href="https://fonts.googleapis.com" />
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
  <link
    href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700;800&display=swap"
    rel="stylesheet"
  />
</svelte:head>

<QueryClientProvider client={queryClient}>
  <main>
    {#if $authReady}
      <Router {routes} onConditionsFailed={handleConditionsFailed} />
    {:else if showSpinner}
      <Spinner fullscreen size={44} />
    {/if}
    {#if !inDashboard}
      <ThemeToggle />
    {/if}
  </main>
</QueryClientProvider>
