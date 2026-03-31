"use client";

import { useState } from "react";
import Link from "next/link";
import { Server, Monitor, Network, Workflow, ArrowRight } from "lucide-react";

const tabs = [
  {
    id: "backend",
    label: "Backend",
    icon: Server,
    iconBg: "bg-teal-50",
    iconColor: "text-teal-600",
    dotColor: "bg-teal-500",
    heading: "Trace every request. Score every endpoint.",
    description:
      "Traceway captures detailed span waterfall traces for every backend request, monitors scheduled tasks and background jobs, and ranks endpoints by real user impact with the Impact Score.",
    bullets: [
      "Full request/response span traces",
      "Scheduled task and background job monitoring",
      "Automatic Impact Score ranking",
      "OpenTelemetry-native ingestion",
    ],
    href: "/product/performance",
  },
  {
    id: "frontend",
    label: "Frontend",
    icon: Monitor,
    iconBg: "bg-purple-50",
    iconColor: "text-purple-600",
    dotColor: "bg-purple-500",
    heading: "Replay the moment before every error.",
    description:
      "See exactly what users did before an exception. Session replays are attached to errors automatically. Source map stack trace resolution turns minified errors into readable, actionable traces.",
    bullets: [
      "Session replay with pre-error capture",
      "Automatic source map stack trace resolution",
      "Click, scroll, and navigation tracking",
      "Linked directly to backend traces",
    ],
    href: "/product/session-replay",
  },
  {
    id: "microservice",
    label: "Microservice",
    icon: Network,
    iconBg: "bg-cyan-50",
    iconColor: "text-cyan-600",
    dotColor: "bg-cyan-500",
    heading: "Follow a request across every service.",
    description:
      "Distributed tracing connects frontend sessions to backend errors across your entire microservice topology. When a backend service throws an exception, you see the user's session replay, the full cross-service trace, and the exact span that failed.",
    bullets: [
      "Cross-service distributed trace propagation",
      "Frontend sessions linked to backend errors",
      "Full trace context across service boundaries",
      "Exception pinpointing across services",
    ],
    href: "/product/distributed-tracing",
  },
  {
    id: "ai-agents",
    label: "AI Agents",
    icon: Workflow,
    iconBg: "bg-violet-50",
    iconColor: "text-violet-600",
    dotColor: "bg-violet-500",
    heading: "Track every AI call, its cost, and its conversation.",
    description:
      "Monitor LLM costs, token usage, and latency across every provider. See the full prompt and completion for every call, with per-agent and per-model breakdowns.",
    bullets: [
      "Per-call cost and token tracking",
      "Conversation replay with chat view",
      "P50/P95 latency per agent and model",
      "Works with OpenRouter and any OTel provider",
    ],
    href: "/product/ai-tracing",
  },
];

export function HomeTabs() {
  const [active, setActive] = useState(0);
  const tab = tabs[active];
  const Icon = tab.icon;

  return (
    <div>
      <div className="flex items-center justify-center mb-8">
        <div className="inline-flex items-center bg-zinc-100 rounded-lg p-1 gap-1">
          {tabs.map((t, i) => (
            <button
              key={t.id}
              onClick={() => setActive(i)}
              className={`px-4 py-2 rounded-md text-sm font-medium transition-all ${
                active === i
                  ? "bg-white text-zinc-900 shadow-sm"
                  : "text-zinc-500 hover:text-zinc-700"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
      </div>

      <div className="rounded-2xl border border-zinc-200 bg-white p-8 md:p-10">
        <div className="flex flex-col md:flex-row items-start gap-8">
          <div className="flex-1 space-y-5">
            <div
              className={`w-12 h-12 ${tab.iconBg} rounded-2xl flex items-center justify-center`}
            >
              <Icon className={`w-6 h-6 ${tab.iconColor}`} />
            </div>
            <h3 className="text-xl md:text-2xl font-bold text-zinc-900 tracking-tight">
              {tab.heading}
            </h3>
            <p className="text-zinc-600 leading-relaxed">{tab.description}</p>
            <ul className="space-y-3 pt-1">
              {tab.bullets.map((b) => (
                <li key={b} className="flex items-center gap-3 text-zinc-700">
                  <div
                    className={`w-1.5 h-1.5 rounded-full ${tab.dotColor}`}
                  ></div>
                  {b}
                </li>
              ))}
            </ul>
            <Link
              href={tab.href}
              className="inline-flex items-center gap-1 text-sm font-medium text-zinc-500 hover:text-zinc-900 transition-colors pt-2"
            >
              Learn more <ArrowRight className="w-3.5 h-3.5" />
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
