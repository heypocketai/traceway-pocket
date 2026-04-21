import Link from "next/link";
import { Smartphone, Github, ArrowRight } from "lucide-react";

import { Chip } from "@/components/chip";
import { Eyebrow } from "@/components/eyebrow";
import { SectionHead } from "@/components/section-head";
import { StatsStrip } from "@/components/stats-strip";
import { FinalCTA } from "@/components/final-cta";
import { Terminal } from "@/components/terminal";
import { AuroraBackground } from "@/components/aurora-background";

export default function FlutterSessionReplayPage() {
  return (
    <main className="relative">
      {/* 1. HERO */}
      <section className="hero hero-product gridbg relative">
        <AuroraBackground variant="hero" />
        <div className="wrap relative z-10">
          <Chip>
            <Smartphone className="h-3 w-3 inline mr-1" />
            Flutter Session Replay
          </Chip>
          <h1 className="mt-6">
            See the crash. <em>Not just the trace.</em>
          </h1>
          <p className="hero-sub">
            Full-screen session recording, pinned to every stack trace. Four
            lines of setup. Zero frame drops on your app.
          </p>
          <div className="hero-cta-row">
            <Link
              href="https://cloud.tracewayapp.com/register"
              className="btn btn-accent"
            >
              Start free
              <ArrowRight className="h-4 w-4" />
            </Link>
            <Link
              href="https://github.com/tracewayapp/traceway-flutter"
              target="_blank"
              rel="noopener noreferrer"
              className="btn btn-ghost"
            >
              <Github className="h-4 w-4" />
              View on GitHub
            </Link>
          </div>

          {/* Free-tier chip — the page's sharpest offer, front-loaded */}
          <div className="mt-5">
            <Chip variant="ok">
              <span
                style={{
                  fontFamily: "var(--font-mono)",
                  letterSpacing: "0.04em",
                }}
              >
                10,000 recordings / month — free, forever
              </span>
            </Chip>
          </div>
        </div>
      </section>

      {/* 2. THE 4 LINES */}
      <section className="wrap py-20">
        <div className="grid gap-12 md:grid-cols-[1fr_1.1fr] items-center">
          <div>
            <Eyebrow>Setup</Eyebrow>
            <h2 className="mt-3">
              Paste four lines. <em>Ship.</em>
            </h2>
            <p className="muted mt-4 max-w-[460px]">
              Recording starts on the first frame. Every exception carries the
              replay ID automatically — open any stack trace in the Traceway
              dashboard, press play.
            </p>
            <p
              className="mt-4 text-[13px]"
              style={{
                color: "var(--fg-3)",
                fontFamily: "var(--font-mono)",
              }}
            >
              // <code>flutter pub add traceway</code> to install.
            </p>
          </div>
          <Terminal
            title="main.dart"
            lines={[
              { type: "cmd", content: "Traceway.run(" },
              {
                type: "tx",
                content:
                  "  connectionString: 'token@cloud.tracewayapp.com/api/report',",
              },
              { type: "tx", content: "  options: TracewayOptions(screenCapture: true)," },
              { type: "cmd", content: "  child: MyApp());" },
            ]}
            showCursor
          />
        </div>
      </section>

      {/* 3. BENCHMARKS */}
      <section className="wrap py-20">
        <SectionHead
          eyebrow="Measured on real hardware"
          title={
            <>
              No frame drops. <em>No battery spike.</em>
            </>
          }
          description="Benchmarked on Pixel 5, Pixel 6, and Pixel 8 via Firebase Test Lab. Full harness open source — run it yourself on your own device tier."
        />
        <StatsStrip
          stats={[
            { num: "<em>0%</em>", label: "Frame-time regression at p50" },
            { num: "<em>5–12</em>MB", label: "Steady-state RAM footprint" },
            { num: "<em>15</em> FPS", label: "Default capture rate" },
            { num: "<em>~250</em>KB", label: "per 10-second recording" },
          ]}
        />
      </section>

      {/* 4. FINAL CTA */}
      <FinalCTA
        title={
          <>
            Ship it <em>before your next release.</em>
          </>
        }
        description="10,000 recordings every month. Free. Forever. No card required."
        primary={{
          label: "Create your project",
          href: "https://cloud.tracewayapp.com/register",
        }}
        secondary={{
          label: "View on GitHub",
          href: "https://github.com/tracewayapp/traceway-flutter",
          external: true,
        }}
      />
    </main>
  );
}
