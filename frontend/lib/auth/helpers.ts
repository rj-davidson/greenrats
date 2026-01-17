import { withAuth, getSignInUrl, getSignUpUrl, signOut } from "@workos-inc/authkit-nextjs";

export async function getCurrentUser() {
  const { user } = await withAuth();
  return user;
}

export async function requireAuth() {
  return withAuth({ ensureSignedIn: true });
}

export { withAuth, getSignInUrl, getSignUpUrl, signOut };
