import { ClientProviders } from "@/lib/providers/client-providers";
import { AuthKitProvider } from "@workos-inc/authkit-nextjs/components";
import { RatIcon } from "lucide-react";
import type { Metadata } from "next";
import { Geist, Geist_Mono, IM_Fell_French_Canon } from "next/font/google";
import Link from "next/link";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const imFellFrenchCanon = IM_Fell_French_Canon({
  variable: "--font-playfair",
  subsets: ["latin"],
  weight: ["400"],
});

export const metadata: Metadata = {
  title: "greenrats - Golf Pick'em",
  description: "Pick one golfer per tournament. Compete with friends. Track your earnings.",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body
        className={`${geistSans.variable} ${geistMono.variable} ${imFellFrenchCanon.variable} antialiased`}
      >
        <AuthKitProvider>
          <ClientProviders>
            <div className="flex min-h-screen flex-col">
              <div className="flex-1">{children}</div>
              <footer className="border-t py-6">
                <div className="mx-auto flex w-full max-w-6xl flex-col items-center gap-3 px-6 text-sm text-muted-foreground sm:flex-row sm:justify-between">
                  <span />
                  <Link
                    href="/"
                    className="order-1 flex items-center gap-2 font-serif text-base tracking-wide text-foreground transition hover:text-foreground/80 sm:order-2 sm:justify-center"
                  >
                    <RatIcon className="size-4 text-primary" />
                    <span>greenrats</span>
                  </Link>
                  <div className="order-3 flex gap-4">
                    <span className="order-2 sm:order-1">
                      © {new Date().getFullYear()} greenrats
                    </span>
                    <Link className="transition hover:text-foreground" href="/terms">
                      Terms of Service
                    </Link>
                    <Link className="transition hover:text-foreground" href="/privacy">
                      Privacy Policy
                    </Link>
                  </div>
                </div>
              </footer>
            </div>
          </ClientProviders>
        </AuthKitProvider>
      </body>
    </html>
  );
}
