import { logout } from "./api/auth";
import { clearSessionQueries } from "./queryClient";
import { clearTokens } from "./stores/tokens";
import { clearUser } from "./stores/user";

export async function signOut() {
  try {
    await logout();
  } catch {
  } finally {
    clearTokens();
    clearUser();
    clearSessionQueries();
  }
}
