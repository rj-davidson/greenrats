import { handleAuth } from "@workos-inc/authkit-nextjs";

import { env } from "@/lib/env";

export const GET = handleAuth({
  returnPathname: "/",
  baseURL: env.BASE_URL,
});
