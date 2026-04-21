import Link from "next/link";
import {
  ArrowRight,
  BarChart3,
  Cpu,
  LayoutDashboard,
  Network,
  Package,
  Radio,
  Ruler,
} from "lucide-react";

import { Chip } from "@/components/chip";
import { SectionHead } from "@/components/section-head";
import { FeatureRow } from "@/components/feature-row";
import { BentoGrid, BentoCell } from "@/components/bento-grid";
import { FaqList } from "@/components/faq-list";
import { FinalCTA } from "@/components/final-cta";
import { AuroraBackground } from "@/components/aurora-background";
import { Terminal } from "@/components/terminal";

export default function MetricsPage() {
  return (
    <main className="relative">
      <section className="hero hero-product gridbg relative">
        <AuroraBackground variant="hero" />
        <div className="wrap relative z-10">
          <Chip variant="ok">
            <BarChart3 className="h-3 w-3 inline mr-1" />
            Metrics
          </Chip>
          <h1 className="mt-6">
            Measure what matters, <em>without the bill shock.</em>
          </h1>
          <p className="hero-sub">
            Custom application metrics, automatic server metrics, flexible
            widget dashboards, and OpenTelemetry ingestion — all included in
            every plan. No per-metric billing, no surprise overages.
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

      {/* Application metrics */}
      <section className="wrap">
        <FeatureRow
          eyebrow="Application metrics"
          title={
            <>
              Custom metrics with <em>one function call</em>
            </>
          }
          description="Counter, Gauge, and Histogram types with dimensional tags. Emit metrics from anywhere in your code — the SDK batches, compresses, and ships them to Traceway without blocking your hot path."
          bullets={[
            "Counter / Gauge / Histogram types",
            "Dimensional tags for facet breakdowns",
            "Single-call SDK API",
            "Zero impact on request latency",
          ]}
          image={{ src: "/images/screenshot-4.png", alt: "Custom metrics dashboard" }}
        />
      </section>

      {/* Code example */}
      <section className="wrap py-16">
        <div className="max-w-3xl mx-auto">
          <Terminal
            title="capture custom metrics"
            lines={[
              { type: "mute", content: "// Track business metrics alongside system metrics" },
              { type: "cmd", content: 'traceway.Metric.Counter("signups", 1, tags)' },
              { type: "cmd", content: 'traceway.Metric.Gauge("queue.depth", 42, tags)' },
              { type: "cmd", content: 'traceway.Metric.Histogram("checkout.ms", 312, tags)' },
              { type: "mute", content: "// Tags can be plan, region, tenant — anything you want to slice by" },
              { type: "ok", content: '✓ metrics indexed by type, tag, and time — query anywhere' },
            ]}
            showCursor
          />
        </div>
      </section>

      {/* Server metrics */}
      <section className="wrap">
        <FeatureRow
          reverse
          eyebrow="Server metrics"
          title="CPU, memory, goroutines, GC — automatic"
          description="The Traceway SDK emits runtime metrics every 10 seconds without any configuration. See host health alongside application metrics in a single view."
          bullets={[
            "CPU usage percentage",
            "Memory (allocated, heap, used%)",
            "Goroutine count + heap object count",
            "GC cycles and pause time",
            "Zero-config, always on",
          ]}
          image={{ src: "/images/screenshot-4.png", alt: "Server metrics dashboard" }}
        />
      </section>

      {/* Widget groups */}
      <section className="wrap">
        <FeatureRow
          eyebrow="Widget groups"
          title={
            <>
              Dashboards that match <em>your team&apos;s mental model</em>
            </>
          }
          description="Pick metrics, pick charts, group them into widget pages. No query language required; filters, tag breakdowns, and rollups are all declarative."
          bullets={[
            "Drag-to-add charts",
            "Group widgets by feature, service, or team",
            "Per-metric filters and rollups",
            "Set default dashboards per organization",
          ]}
          image={{ src: "/images/screenshot-3.png", alt: "Widget groups dashboard" }}
        />
      </section>

      {/* Ingest bento */}
      <section className="wrap py-10">
        <SectionHead
          eyebrow="Ingest"
          title={
            <>
              OpenTelemetry metrics — <em>first-class</em>
            </>
          }
          description="Bring metrics from anywhere. Traceway speaks OTLP natively and registers every metric with type, unit, and cardinality."
        />
        <BentoGrid>
          <BentoCell
            size="wide"
            icon={Network}
            title="OTLP/HTTP + OTLP/gRPC"
            iconColor="var(--a2)"
          >
            <p>
              Point any OpenTelemetry SDK at Traceway. Sum, Gauge, and
              Histogram metric types are all understood natively — no
              translation layer, no metric-name mangling.
            </p>
          </BentoCell>
          <BentoCell
            size="med"
            icon={Package}
            title="Metric Registry"
            iconColor="var(--a1)"
          >
            <p>
              Every metric shows up with its type, unit, description,
              last-seen timestamp, and cardinality. Spot runaway tag explosions
              before they hit your bill.
            </p>
          </BentoCell>
          <BentoCell
            size="med"
            icon={Radio}
            title="Prometheus scrape"
            iconColor="var(--ok)"
          >
            <p>
              Traceway can pull <code>/metrics</code> endpoints on a
              configurable interval. Your existing Prometheus exporters work
              unchanged.
            </p>
          </BentoCell>
          <BentoCell size="sm" icon={Cpu} title="StatsD compat">
            <p>UDP ingestion for legacy StatsD metric streams.</p>
          </BentoCell>
          <BentoCell size="sm" icon={Ruler} title="Custom units">
            <p>Bytes, seconds, percentages — declare once, render consistently.</p>
          </BentoCell>
          <BentoCell size="sm" icon={LayoutDashboard} title="Always in context">
            <p>Metrics link back to the traces, logs, and issues around them.</p>
          </BentoCell>
        </BentoGrid>
      </section>

      <FinalCTA
        title={
          <>
            Ship metrics <em>in 5 minutes</em>
          </>
        }
        description="Application + server metrics. Included on every plan. No per-metric billing."
        primary={{
          label: "Read the Metrics docs",
          href: "https://docs.tracewayapp.com",
        }}
        secondary={{
          label: "Try Live Demo",
          href: "https://cloud.tracewayapp.com/login?email=demo@tracewayapp.com&password=demoaccount!",
        }}
      />

      <section className="wrap pt-10 pb-24">
        <div className="max-w-3xl mx-auto">
          <SectionHead align="center" eyebrow="FAQ" title="Questions about metrics" />
          <div className="mt-4">
            <FaqList
              items={[
                {
                  q: "How do I emit a custom metric?",
                  a: (
                    <>
                      <p>
                        With the Traceway SDK, a single call:{" "}
                        <code>traceway.Metric.Counter(&quot;signups&quot;, 1, tags)</code>
                        {" "}— that&apos;s it. Full docs cover Gauge and Histogram.
                      </p>
                      <p>
                        Or send metrics via OpenTelemetry at{" "}
                        <code>/api/otel/v1/metrics</code> — OTLP/HTTP and
                        OTLP/gRPC are both supported natively.
                      </p>
                    </>
                  ),
                },
                {
                  q: "What server metrics are collected automatically?",
                  a: "The Traceway SDK automatically emits CPU usage, memory usage (allocated, heap, used%), goroutine count, heap object count, GC cycle count, and GC pause time every 10 seconds. No configuration required — they show up in the Metrics dashboard out of the box.",
                },
                {
                  q: "Do custom metrics count toward my event limit?",
                  a: "No. Metrics are included at no additional event cost — only issues, HTTP requests, and background tasks count toward your event limit. This means you can emit thousands of custom metrics without worrying about billing.",
                },
                {
                  q: "Can I query metrics by tag or dimension?",
                  a: "Yes. Every tag becomes a facet you can filter on; widget groups let you build per-dimension chart panels. For example, a `plan` tag on a signups metric lets you chart signups broken down by plan, region, or tenant.",
                },
                {
                  q: "Does Traceway support Prometheus?",
                  a: "Yes. Traceway can scrape /metrics endpoints on a configurable interval. Your existing Prometheus exporters work unchanged — just point Traceway at their address and choose a scrape interval.",
                },
              ]}
            />
          </div>
        </div>
      </section>
    </main>
  );
}
