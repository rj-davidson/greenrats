import { withAuth, getSignInUrl, getSignUpUrl, signOut } from "@workos-inc/authkit-nextjs";

/**
 * @deprecated Use `useCurrentUser()` hook from `@/features/users/queries` instead.
 * This returns the WorkOS user, not the database user.
 * The `useCurrentUser()` hook returns the database user with fields like `display_name`.
 */
export async function getCurrentUser() {
  const { user } = await withAuth();
  return user;
}

export async function requireAuth() {
  return withAuth({ ensureSignedIn: true });
}

export { withAuth, getSignInUrl, getSignUpUrl, signOut };
