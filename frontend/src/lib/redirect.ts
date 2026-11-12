export function readRedirect(querystring: string | undefined): string | null {
  const target = new URLSearchParams(querystring ?? "").get("redirect");
  if (!target || !target.startsWith("/") || target.startsWith("//")) return null;
  return target;
}

export function withRedirect(path: string, target: string | null): string {
  return target ? `${path}?redirect=${encodeURIComponent(target)}` : path;
}
