import Image from "next/image";
import Link from "next/link";
import { ArrowRight, Bug, Layers, Globe, TrendingUp } from "lucide-react";

import { Chip } from "@/components/chip";
import { Eyebrow } from "@/components/eyebrow";
import { SectionHead } from "@/components/section-head";
import { FeatureRow } from "@/components/feature-row";
import { FaqList } from "@/components/faq-list";
import { FinalCTA } from "@/components/final-cta";
import { AuroraBackground } from "@/components/aurora-background";
import { ImpactScoreVisual } from "@/components/impact-score-visual";

const frameworks = [
  { name: "Gin", src: "/images/frameworks/gin.png" },
  { name: "Chi", src: "/images/frameworks/chi.png" },
  { name: "Express", src: "/images/frameworks/express.png" },
  { name: "NestJS", src: "/images/frameworks/nestjs.png" },
  { name: "Next.js", src: "/images/frameworks/nextjs.png" },
  { name: "Svelte", src: "/images/frameworks/svelte.png" },
  { name: "Remix", src: "/images/frameworks/remix.png" },
  { name: "OpenTelemetry", src: "/images/frameworks/otel.png" },
  { name: "Cloudflare", src: "/images/frameworks/cloudflare.png" },
];

export default function StackTracesPage() {
  return (
    <main className="relative">
      {/* Hero — left aligned */}
      <section className="hero hero-product gridbg relative">
        <AuroraBackground variant="hero" />
        <div className="wrap relative z-10">
          <Chip variant="crit">
            <Bug className="h-3 w-3 inline mr-1" />
            Stack Traces
          </Chip>
          <h1 className="mt-6">
            Find and fix issues <em>before your users notice.</em>
          </h1>
          <p className="hero-sub">
            Every exception, grouped and ranked by automatic Impact Score.
            Traceway normalizes stack traces with a 10-step pipeline and
            SHA-256 hashes them — thousands of duplicates collapse into one
            ranked issue.
          </p>
          <div className="hero-cta-row">
            <Link href="https://docs.tracewayapp.com" className="btn btn-accent">
              Get Started <ArrowRight className="h-4 w-4" />
            </Link>
            <Link href="https://cloud.tracewayapp.com/register" className="btn btn-ghost">
              Try Traceway Cloud
            </Link>
          </div>
        </div>
      </section>

      {/* Impact Score section */}
      <section className="wrap py-20">
        <SectionHead
          eyebrow="Impact Score"
          title={
            <>
              One score. Five signals. <em>Zero guesswork.</em>
            </>
          }
          description="The Impact Score combines five service-level indicators into one automatic priority for every endpoint. It takes the max across all five — if any single signal is bad, the endpoint surfaces immediately."
        />
        <div className="mt-10 max-w-4xl mx-auto">
          <ImpactScoreVisual />
        </div>
      </section>

      {/* Every exception grouped and ranked — absorbed from home */}
      <section className="wrap">
        <FeatureRow
          eyebrow="Grouping"
          title={
            <>
              Every exception, <em>grouped and ranked</em>
            </>
          }
          description="Full stack traces, 10-step normalization, SHA-256 grouping. Thousands of duplicates collapse into one ranked issue so you fix what matters first."
          bullets={[
            "Full stack trace capture with file:line",
            "Intelligent error grouping via SHA-256 hash",
            "User impact analysis across sessions",
            "Source map resolution for minified JS",
          ]}
          image={{ src: "/images/screenshot-2.png", alt: "Exception tracking interface" }}
        />
      </section>

      {/* Issues ranked by what matters */}
      <section className="wrap">
        <FeatureRow
          reverse
          eyebrow="Ranking"
          title="Issues ranked by what matters"
          description="Stop triaging manually. Traceway ranks every issue by frequency, user impact, and recency so your team focuses on the problems that matter most. New regressions surface immediately."
          bullets={[
            "Impact-based ranking across endpoints",
            "Regression detection on new releases",
            "Frequency and recency scoring",
          ]}
          image={{ src: "/images/screenshot-3.png", alt: "Issue ranking dashboard" }}
        />
      </section>

      {/* Intelligent grouping */}
      <section className="wrap">
        <FeatureRow
          eyebrow="Normalization"
          title={
            <>
              Same bug, <em>same group</em> — every time
            </>
          }
          description="Traceway normalizes stack traces before hashing, so the same logical error gets grouped together even when runtime values differ. Memory addresses, UUIDs, timestamps, numeric IDs, and ANSI codes are stripped before the hash."
          bullets={[
            "Stack trace normalization (10-step pipeline)",
            "Cross-service deduplication",
            "Full context preserved on every occurrence",
          ]}
          image={{ src: "/images/screenshot-4.png", alt: "Error grouping interface" }}
        />
      </section>

      {/* Framework showcase */}
      <section className="wrap py-20 text-center">
        <div className="max-w-3xl mx-auto flex flex-col items-center gap-5">
          <div
            className="w-12 h-12 rounded-[10px] flex items-center justify-center"
            style={{
              background: "color-mix(in oklab, var(--a4) 18%, transparent)",
              border: "1px solid color-mix(in oklab, var(--a4) 40%, transparent)",
              color: "var(--a4)",
            }}
          >
            <Globe className="w-5 h-5" />
          </div>
          <Eyebrow>Full stack</Eyebrow>
          <h2>Track issues across your full stack</h2>
          <p style={{ color: "var(--fg-1)", fontSize: 17 }}>
            From Go backends to JavaScript frontends, Traceway captures
            exceptions everywhere your code runs. Unified view of issues across
            services, with full stack traces and contextual tags.
          </p>
          <div className="flex flex-wrap items-center justify-center gap-8 md:gap-10 pt-6 opacity-80">
            {frameworks.map((fw) => (
              <Image
                key={fw.name}
                src={fw.src}
                alt={fw.name}
                width={40}
                height={40}
                className="h-8 w-auto opacity-80 hover:opacity-100 transition-all"
              />
            ))}
          </div>
        </div>
      </section>

      <FinalCTA
        title={
          <>
            Triage <em>faster</em>. Ship <em>safer</em>.
          </>
        }
        description="Connect an SDK, ship an error, see it in Traceway. 5-minute setup."
        primary={{ label: "Get Started", href: "https://docs.tracewayapp.com" }}
        secondary={{
          label: "Try Live Demo",
          href: "https://cloud.tracewayapp.com/login?email=demo@tracewayapp.com&password=demoaccount!",
        }}
      />

      <section className="wrap pt-10 pb-24">
        <div className="max-w-3xl mx-auto">
          <SectionHead align="center" eyebrow="FAQ" title="Questions about stack traces" />
          <div className="mt-4">
            <FaqList
              items={[
                {
                  q: "What is the Impact Score?",
                  a: (
                    <>
                      <p>
                        The Impact Score is Traceway&apos;s automatic prioritization
                        system. It combines five service-level indicators — an
                        inverted apdex variant, error rate floor, P99 latency
                        floor, client error floor, and volume error floor — into
                        a single 0-100 score for every endpoint.
                      </p>
                      <p>
                        It takes the max across all five, so if any single
                        signal is degraded that endpoint surfaces immediately.
                        You can adjust the slow endpoint threshold per endpoint
                        to tune the apdex calculation for routes with different
                        latency profiles.
                      </p>
                      <p>
                        For incident management, notification rules fire when
                        the Impact Score or error rate crosses a threshold, and
                        route alerts to Slack, GitHub Issues, email, or custom
                        webhooks.
                      </p>
                    </>
                  ),
                },
                {
                  q: "How does error grouping work?",
                  a: "Traceway applies a 10-step normalization pipeline to every stack trace: extracting the error type, removing absolute file paths, replacing hex addresses, UUIDs, IPs, timestamps, and numeric IDs with placeholders, normalizing whitespace, and stripping ANSI codes. The result is hashed with SHA-256 so identical logical errors always group together, even if runtime values differ.",
                },
                {
                  q: "How does automatic issue ranking work?",
                  a: "Traceway scores each issue based on how often it occurs, how recently it appeared, and how many users are affected. Issues are continuously re-ranked as new data comes in, so regressions and trending problems surface immediately — no manual triage required.",
                },
                {
                  q: "How does error grouping handle different environments?",
                  a: "Traceway normalizes stack traces by removing runtime-specific values like memory addresses, file paths, UUIDs, and timestamps before hashing. This means the same bug produces the same group regardless of which server or environment it occurred on.",
                },
                {
                  q: "Can I track frontend JavaScript errors?",
                  a: "Yes. Traceway supports frontend frameworks like Next.js, Svelte, and Remix alongside backend frameworks like Express and NestJS. Errors from both your frontend and backend appear in the same dashboard with full stack traces. Source maps upload resolves minified traces back to your original source.",
                },
              ]}
            />
          </div>
        </div>
      </section>
    </main>
  );
}
