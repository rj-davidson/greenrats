import { handleAuth } from "@workos-inc/authkit-nextjs";

export const GET = handleAuth({
  returnPathname: "/",
  baseURL: process.env.BASE_URL || "https://greenrats.com",
});
