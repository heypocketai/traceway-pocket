import Image from "next/image";
import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { ArrowRight, DollarSign, MessageSquareText, BarChart3, Shield } from "lucide-react";

export default function AiTracingPage() {
    return (
        <main className="min-h-screen bg-white text-zinc-950 font-sans selection:bg-zinc-100 selection:text-zinc-900">
            {/* Hero Section */}
            <section className="relative pt-16 pb-20 overflow-hidden">
                <div className="absolute inset-0 -z-10 h-full w-full bg-white bg-[radial-gradient(#e5e7eb_1px,transparent_1px)] [background-size:16px_16px] [mask-image:radial-gradient(ellipse_50%_50%_at_50%_50%,#000_70%,transparent_100%)]"></div>
                <div className="container mx-auto px-4 text-center">
                    <Badge variant="secondary" className="mb-4 bg-violet-50 text-violet-700 hover:bg-violet-100 px-2.5 py-0.5 border border-violet-100 text-xs font-normal rounded-full">
                        AI Tracing
                    </Badge>
                    <h1 className="text-4xl md:text-6xl font-bold tracking-tight mb-6 text-zinc-900">
                        See every AI call, <br /> <span className="text-transparent bg-clip-text bg-gradient-to-r from-violet-600 to-blue-600">its cost, and its conversation</span>
                    </h1>
                    <p className="text-zinc-600 text-lg md:text-xl max-w-2xl mx-auto mb-10 leading-relaxed font-medium">
                        Monitor LLM costs, token usage, latency, and conversations across every AI provider. Works with OpenRouter, OpenAI, Anthropic, and any OpenTelemetry-compatible provider.
                    </p>
                    <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                        <Link href="https://docs.tracewayapp.com/client/openrouter" className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-all cursor-pointer h-10 px-6 bg-zinc-900 text-white hover:bg-zinc-800 shadow-lg shadow-zinc-900/20">
                            Get Started <ArrowRight className="ml-2 h-4 w-4" />
                        </Link>
                        <Link href="http://cloud.tracewayapp.com/register" className="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-all cursor-pointer h-10 px-6 border border-zinc-200 bg-white hover:bg-zinc-50 text-zinc-900 shadow-sm">
                            Try Traceway Cloud
                        </Link>
                    </div>
                </div>
            </section>

            {/* Cost Visibility Section */}
            <section className="py-24 bg-white border-y border-zinc-100">
                <div className="container mx-auto px-4 max-w-5xl">
                    <div className="flex flex-col md:flex-row items-center gap-12 lg:gap-20">
                        <div className="flex-1 space-y-6">
                            <div className="w-12 h-12 bg-green-50 rounded-2xl flex items-center justify-center">
                                <DollarSign className="w-6 h-6 text-green-600" />
                            </div>
                            <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">Know exactly what AI costs you</h2>
                            <p className="text-zinc-600 text-lg leading-relaxed">
                                Every AI call is tracked with input cost, output cost, and total cost. See cost breakdowns per agent, per model, and per time period.
                                Spot cost spikes before they hit your invoice.
                            </p>
                            <ul className="space-y-3 pt-2">
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                    Per-call cost tracking with input/output breakdown
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                    Aggregated cost per agent and model
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div>
                                    Token usage with cached and reasoning token tracking
                                </li>
                            </ul>
                        </div>
                        <div className="flex-1 w-full relative">
                            <div className="absolute inset-0 bg-gradient-to-tr from-green-100/50 to-transparent rounded-3xl transform rotate-3 scale-105 -z-10"></div>
                            <div className="relative rounded-xl overflow-hidden border border-zinc-200 bg-white">
                                <Image
                                    src="/images/ai-traces-cost.png"
                                    alt="AI Traces Cost Dashboard"
                                    width={800}
                                    height={600}
                                    className="w-full h-auto"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Conversation Replay Section */}
            <section className="py-24 bg-zinc-50/50">
                <div className="container mx-auto px-4 max-w-5xl">
                    <div className="flex flex-col md:flex-row-reverse items-center gap-12 lg:gap-20">
                        <div className="flex-1 space-y-6">
                            <div className="w-12 h-12 bg-violet-50 rounded-2xl flex items-center justify-center">
                                <MessageSquareText className="w-6 h-6 text-violet-600" />
                            </div>
                            <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">Replay every conversation</h2>
                            <p className="text-zinc-600 text-lg leading-relaxed">
                                See the exact prompt sent to the model and the full response it generated. Debug unexpected model behavior, catch hallucinations, and understand what your AI agents are actually doing.
                            </p>
                            <ul className="space-y-3 pt-2">
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-violet-500"></div>
                                    Full prompt and completion stored and rendered as chat
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-violet-500"></div>
                                    Raw JSON view for debugging edge cases
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-violet-500"></div>
                                    Privacy mode available to exclude conversation content
                                </li>
                            </ul>
                        </div>
                        <div className="flex-1 w-full relative">
                            <div className="absolute inset-0 bg-gradient-to-tl from-violet-100/50 to-transparent rounded-3xl transform -rotate-3 scale-105 -z-10"></div>
                            <div className="relative rounded-xl overflow-hidden border border-zinc-200 bg-white">
                                <Image
                                    src="/images/ai-traces-conversation.png"
                                    alt="AI Trace Conversation View"
                                    width={800}
                                    height={600}
                                    className="w-full h-auto"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Performance & Latency Section */}
            <section className="py-24 bg-white border-y border-zinc-100">
                <div className="container mx-auto px-4 max-w-5xl">
                    <div className="flex flex-col md:flex-row items-center gap-12 lg:gap-20">
                        <div className="flex-1 space-y-6">
                            <div className="w-12 h-12 bg-blue-50 rounded-2xl flex items-center justify-center">
                                <BarChart3 className="w-6 h-6 text-blue-600" />
                            </div>
                            <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">Understand AI latency at every percentile</h2>
                            <p className="text-zinc-600 text-lg leading-relaxed">
                                AI calls have wildly variable latency. See P50 and P95 duration breakdowns per agent, identify which models or providers are causing slowdowns, and track performance over time.
                            </p>
                            <ul className="space-y-3 pt-2">
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                                    P50/P95 latency per agent and model
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                                    Drill down from agent to individual calls
                                </li>
                                <li className="flex items-center gap-3 text-zinc-700">
                                    <div className="w-1.5 h-1.5 rounded-full bg-blue-500"></div>
                                    Throughput and token usage trends
                                </li>
                            </ul>
                        </div>
                        <div className="flex-1 w-full relative">
                            <div className="absolute inset-0 bg-gradient-to-tr from-blue-100/50 to-transparent rounded-3xl transform rotate-3 scale-105 -z-10"></div>
                            <div className="relative rounded-xl overflow-hidden border border-zinc-200 bg-white">
                                <Image
                                    src="/images/ai-traces-latency.png"
                                    alt="AI Traces Performance Dashboard"
                                    width={800}
                                    height={600}
                                    className="w-full h-auto"
                                />
                            </div>
                        </div>
                    </div>
                </div>
            </section>

            {/* Zero Code Section */}
            <section className="py-24 bg-zinc-50/50">
                <div className="container mx-auto px-4 max-w-5xl">
                    <div className="flex flex-col items-center text-center space-y-6">
                        <div className="w-12 h-12 bg-amber-50 rounded-2xl flex items-center justify-center">
                            <Shield className="w-6 h-6 text-amber-600" />
                        </div>
                        <h2 className="text-2xl md:text-3xl font-bold text-zinc-900 tracking-tight">Zero code changes required</h2>
                        <p className="text-zinc-600 text-lg leading-relaxed max-w-2xl">
                            If you&apos;re using OpenRouter, enable Observability in your settings and point it at Traceway. That&apos;s it.
                            For other providers, any OpenTelemetry-instrumented AI call with <code className="bg-zinc-100 px-1.5 py-0.5 rounded text-sm">gen_ai.*</code> attributes is automatically captured.
                        </p>
                        <div className="flex flex-wrap items-center justify-center gap-8 md:gap-10 pt-6">
                            <Image src="/images/frameworks/openrouter.png" alt="OpenRouter" width={40} height={40} className="h-8 w-auto opacity-80 hover:opacity-100 transition-all duration-200" />
                            <Image src="/images/frameworks/otel.png" alt="OpenTelemetry" width={40} height={40} className="h-8 w-auto opacity-80 hover:opacity-100 transition-all duration-200" />
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
                            Common questions about AI tracing with Traceway.
                        </p>
                    </div>

                    <Accordion type="single" collapsible className="w-full">
                        <AccordionItem value="item-1" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                Why do I need AI observability?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                AI API calls are expensive and unpredictable. A single prompt with a large context window can cost 100x more than average.
                                Without observability, cost spikes go unnoticed until the invoice arrives. AI tracing gives you per-call visibility into costs,
                                token usage, latency, and the actual conversations happening between your app and AI models.
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="item-2" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                Which AI providers are supported?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                Any provider or gateway that exports OTLP traces with <code className="bg-zinc-100 px-1 py-0.5 rounded text-sm">gen_ai.*</code> semantic convention attributes.
                                OpenRouter has built-in support — just enable Observability in your settings. For other providers, use any OpenTelemetry SDK
                                to instrument your AI calls and send them to Traceway.
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="item-3" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                Is the conversation content stored securely?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                Conversation content (prompts and completions) is stored separately from trace metadata in object storage (S3 or local filesystem).
                                If you&apos;re self-hosting, the data never leaves your infrastructure. OpenRouter also offers a Privacy Mode that excludes
                                conversation content entirely, sending only metadata like model, tokens, and costs.
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="item-4" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                How does this work with OpenRouter?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                OpenRouter has a built-in Observability feature that broadcasts OTLP traces for every LLM call.
                                You add Traceway as an &ldquo;OpenTelemetry Collector&rdquo; destination in your OpenRouter settings with your Traceway endpoint
                                and project token. No code changes needed — OpenRouter handles the instrumentation automatically.
                            </AccordionContent>
                        </AccordionItem>
                        <AccordionItem value="item-5" className="border-b-zinc-200">
                            <AccordionTrigger className="text-zinc-900 hover:text-zinc-700 hover:no-underline text-left">
                                Can I track costs across multiple models?
                            </AccordionTrigger>
                            <AccordionContent className="text-zinc-600 leading-relaxed">
                                Yes. AI traces capture the specific model used for each call (e.g., <code className="bg-zinc-100 px-1 py-0.5 rounded text-sm">openai/gpt-4-turbo</code>, <code className="bg-zinc-100 px-1 py-0.5 rounded text-sm">anthropic/claude-3-opus</code>).
                                You can see cost breakdowns per model, compare token efficiency across providers, and identify which models give you the best value.
                            </AccordionContent>
                        </AccordionItem>
                    </Accordion>
                </div>
            </section>
        </main>
    );
}
