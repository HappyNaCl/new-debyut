const creatorRoutePattern = /^\/(admin|organizer)(\/|$)/;
const accountRoutePattern = /^\/(tickets|vod|invitations|profile|settings)(\/|$)/;

export function isCreatorRoute(location: string): boolean {
  return creatorRoutePattern.test(location);
}

export function isDashboardRoute(location: string): boolean {
  return isCreatorRoute(location) || accountRoutePattern.test(location);
}
