import Image from "next/image";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { ArrowRight, ScrollText, Network, Boxes } from "lucide-react";

export default function LogsPage() {
    return (
        <main className="min-h-screen bg-white text-zinc-950 font-sans selection:bg-zinc-100 selection:text-zinc-900">
            {/* Hero Section */}
            <section className="relative pt-16 pb-20 overflow-hidden">
                <div className="absolute inset-0 -z-10 h-full w-full bg-white bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] [background-size:16px_16px] [mask-image:radial-gradient(ellipse_50%_50%_at_50%_50%,#000_70%,transparent_100%)]"></div>
                <div className="container mx-auto px-4 text-center">
                    <Badge variant="secondary" className="mb-4 bg-amber-50 text-amber-700 hover:bg-amber-100 px-2.5 py-0.5 border border-amber-100 text-xs font-normal rounded-full">
                        Logs
                    </Badge>
                    <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6 text-zinc-900">
                        Every log, <br /> <span className="text-transparent bg-clip-text bg-gradient-to-r from-amber-500 to-orange-600">with its full trace attached</span>
                    </h1>
                    <p className="text-zinc-600 text-lg md:text-xl max-w-2xl mx-auto mb-10 leading-relaxed font-medium">
                        Stop jumping between a log viewer and a trace viewer. Search by severity, service, or attribute, then open the exact request that produced any log line in one click.
                    </p>
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                        <Link href="https://docs.tracewayapp.com/learn/logs" className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-all cursor-pointer h-10 px-6 bg-zinc-900 text-white hover:bg-zinc-800 shadow-lg shadow-zinc-900/20">
                                Get Started <ArrowRight className="ml-2 h-4 w-4" />
                        </Link>
                        <Link href="http://cloud.tracewayapp.com/register" className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-all cursor-pointer h-10 px-6 border border-zinc-200 bg-white hover:bg-zinc-50 text-zinc-900 shadow-sm">
                                Try Traceway Cloud
                        </Link>
                    </div>
                </div>
            </section>

            {/* Search Across Every Log Section */}
            <section className="py-24 bg-white border-y border-zinc-100">
                <div className="container mx-auto px-4 max-w-5xl">
                    <div className="flex flex-col md:flex-row items-center gap-12 lg:gap-20">
                        <div className="flex-1 space-y-6">
                            <div className="w-12 h-12 bg-amber-50 rounded-2xl flex items-center justify-center">
                                <ScrollText className="w-6 h-6 text-amber-600" />
                            </div>
                            <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">Search across every log</h2>
                            <p className="text-zinc-600 text-lg leading-relaxed">
                                Filter by severity, service, or trace. Full-text search over log bodies is powered by a token
                                index, and any resource, scope, or log attribute can be queried for exact matches.
                            </p>
                            <ul className="space-y-3 pt-2">
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-amber-500"></div>
                                    Body search backed by token indexes
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-amber-500"></div>
                                    Six severity levels: TRACE, DEBUG, INFO, WARN, ERROR, FATAL
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-amber-500"></div>
                                    Attribute filters on resource, scope, and log fields
                                </li>
                            </ul>
                        </div>
                        <div className="flex-1 w-full relative">
                            <div className="absolute inset-0 bg-gradient-to-tr from-amber-100/50 to-transparent rounded-3xl transform rotate-3 scale-105 -z-10"></div>
                            <div className="relative rounded-xl overflow-hidden border border-zinc-200 bg-white">
                                <Image
                                    src="/images/logs.png"
                                    alt="Logs search and detail view"
                                    width={1200}
                                    height={700}
                                    className="w-full h-auto"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Linked to Traces + OpenTelemetry Native (2-card grid) */}
            <section className="py-24 bg-zinc-50/50">
                <div className="container mx-auto px-4 max-w-5xl">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                        <div className="rounded-2xl border border-zinc-200 bg-white p-8 space-y-5">
                            <div className="w-12 h-12 bg-orange-50 rounded-2xl flex items-center justify-center">
                                <Network className="w-6 h-6 text-orange-600" />
                            </div>
                            <h2 className="text-xl md:text-2xl font-bold text-zinc-900 tracking-tight">Linked to every trace</h2>
                            <p className="text-zinc-600 leading-relaxed">
                                Every log carries the trace and span ID of the request that emitted it. Open any
                                endpoint and see the exact log lines tied to that invocation &mdash; or follow a distributed trace
                                to see logs from every service it touched.
                            </p>
                            <ul className="space-y-3 pt-1">
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                                    Trace and span context on every log
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                                    Per-trace and per-distributed-trace views
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-orange-500"></div>
                                    Jump from a log to the span that produced it
                                </li>
                            </ul>
                        </div>

                        <div className="rounded-2xl border border-zinc-200 bg-white p-8 space-y-5">
                            <div className="w-12 h-12 bg-green-50 rounded-2xl flex items-center justify-center">
                                <Boxes className="w-6 h-6 text-green-600" />
                            </div>
                            <h2 className="text-xl md:text-2xl font-bold text-zinc-900 tracking-tight">OpenTelemetry-native</h2>
                            <p className="text-zinc-600 leading-relaxed">
                                Send logs from any OTel SDK &mdash; Node.js, Python, Go, Java, .NET, PHP. No vendor client needed.
                                OTLP/HTTP ingestion supports both Protobuf and JSON, and data is stored with a 30-day TTL.
                            </p>
                            <ul className="space-y-3 pt-1">
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                    Standard OTLP endpoint at /api/otel/v1/logs
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                    Works with the OTel Collector or direct SDK export
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                    30-day retention, fully indexed
                                </li>
                            </ul>
                        </div>
                    </div>
                </div>
            </section>

            {/* FAQ Section */}
            <section className="py-24 bg-zinc-50 border-t border-zinc-100">
                <div className="container mx-auto px-4 max-w-3xl">
                    <div className="text-center mb-12">
                        <h2 className="text-3xl font-bold mb-4 text-zinc-900 tracking-tight">Frequently Asked Questions</h2>
                        <p className="text-zinc-600 text-lg">
                            Common questions about logs with Traceway.
                        </p>
                    </div>

                    <Accordion type="single" collapsible className="w-full">
                        <AccordionItem value="item-1" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                How do I send logs to Traceway?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                Point any OpenTelemetry Logs SDK at <code>/api/otel/v1/logs</code> with an
                                {" "}<code>Authorization: Bearer &lt;project_token&gt;</code> header. Protobuf and JSON are
                                both supported. If you already run an OTel Collector, route its <code>logs</code> pipeline to
                                the same endpoint.
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="item-2" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                Are logs linked to my traces automatically?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                Yes. Every log carries the OTel <code>trace_id</code> and <code>span_id</code> from the
                                active context at emission time. That means logs emitted inside a request handler, background
                                job, or child span are automatically associated with the corresponding trace &mdash; no extra
                                plumbing required.
                            </AccordionContent>
                        </AccordionItem>
                    </Accordion>
                </div>
            </section>
        </main>
    );
}
