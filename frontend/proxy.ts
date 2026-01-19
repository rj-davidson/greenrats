import { authkitMiddleware } from "@workos-inc/authkit-nextjs";

export default authkitMiddleware({
  redirectUri: process.env.WORKOS_REDIRECT_URI,
});

export const config = {
  matcher: [
    // Match all paths except static files and api routes handled by handleAuth
    "/((?!_next/static|_next/image|favicon.ico|api/auth).*)",
  ],
};
