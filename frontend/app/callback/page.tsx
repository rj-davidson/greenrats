// This page exists as a placeholder for the OAuth callback
// The actual callback is handled by the /api/auth/[...auth] route
// WorkOS AuthKit middleware handles the redirect after authentication

export default function CallbackPage() {
  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-muted-foreground">Processing authentication...</p>
    </main>
  );
}
