import { QueryClient } from "@tanstack/svelte-query";

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 60_000,
      retry: 1,
    },
  },
});

export function clearSessionQueries() {
  queryClient.removeQueries({ predicate: (query) => query.queryKey[0] !== "categories" });
}
