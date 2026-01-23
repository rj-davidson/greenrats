const effectiveDate = "January 23, 2026";

export default function PrivacyPage() {
  return (
    <main className="mx-auto w-full max-w-3xl px-6 py-12">
      <h1 className="text-3xl font-semibold">Privacy Policy</h1>
      <p className="mt-2 text-sm text-muted-foreground">Effective date: {effectiveDate}</p>

      <section className="mt-8 space-y-3 text-sm leading-6 text-muted-foreground">
        <p>
          This Privacy Policy explains how GreenRats collects, uses, and shares information when you
          use our website, applications, and services (the "Service").
        </p>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Information We Collect</h2>
        <div className="space-y-3 text-muted-foreground">
          <p>
            Account information: name, email address, display name, and authentication identifiers.
          </p>
          <p>
            Subscription and billing information: plan selection, billing status, and payment
            processor identifiers. We do not store full payment card numbers.
          </p>
          <p>League and gameplay data: league names, picks, and league membership details.</p>
          <p>Usage and device data: log data, IP address, browser type, and pages viewed.</p>
        </div>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">How We Use Information</h2>
        <div className="space-y-3 text-muted-foreground">
          <p>Provide, operate, and maintain the Service.</p>
          <p>Process subscriptions and manage billing.</p>
          <p>Respond to support requests and communicate with you.</p>
          <p>Prevent fraud, abuse, and security incidents.</p>
          <p>Improve the Service, including analytics and performance monitoring.</p>
        </div>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">How We Share Information</h2>
        <div className="space-y-3 text-muted-foreground">
          <p>
            Service providers: We share information with vendors who help us operate the Service,
            such as payment processors, hosting providers, analytics, and customer support tools.
          </p>
          <p>
            Legal and safety: We may disclose information if required by law or to protect the
            rights, safety, and security of our users and the Service.
          </p>
          <p>
            Business transfers: If we are involved in a merger, acquisition, or asset sale, your
            information may be transferred as part of that transaction.
          </p>
          <p>With consent: We may share information with your consent or at your direction.</p>
        </div>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Data Retention</h2>
        <p className="text-muted-foreground">
          We retain information for as long as necessary to provide the Service and for legitimate
          business purposes. You may request account deletion by contacting dev@greenrats.com.
        </p>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Security</h2>
        <p className="text-muted-foreground">
          We use reasonable administrative, technical, and physical safeguards to protect your
          information. No method of transmission or storage is completely secure.
        </p>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Your Choices</h2>
        <div className="space-y-3 text-muted-foreground">
          <p>You can update your account information within the Service.</p>
          <p>You can cancel your subscription at any time.</p>
          <p>You can request account deletion by contacting dev@greenrats.com.</p>
        </div>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Children&apos;s Privacy</h2>
        <p className="text-muted-foreground">
          The Service is not intended for children under 13. We do not knowingly collect personal
          information from children under 13.
        </p>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">International Users</h2>
        <p className="text-muted-foreground">
          The Service is operated in the United States. If you access the Service from outside the
          United States, you consent to the transfer and processing of your information in the
          United States.
        </p>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Changes to This Policy</h2>
        <p className="text-muted-foreground">
          We may update this Privacy Policy from time to time. If we make material changes, we will
          provide notice by updating the effective date and, if appropriate, by additional notice
          within the Service.
        </p>
      </section>

      <section className="mt-8 space-y-3 text-sm leading-6">
        <h2 className="text-lg font-semibold text-foreground">Contact</h2>
        <p className="text-muted-foreground">
          Questions about this policy? Contact us at dev@greenrats.com.
        </p>
      </section>
    </main>
  );
}
