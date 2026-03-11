import { authkitMiddleware } from "@workos-inc/authkit-nextjs";

import { env } from "@/lib/env";

export default authkitMiddleware({
  redirectUri: env.WORKOS_REDIRECT_URI,
});

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|api/auth).*)",
  ],
};
