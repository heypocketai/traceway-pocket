import { useRouter } from "next/router";
import { useSdk } from "./SdkContext";

const FRAMEWORKS = [
  {
    value: "openrouter",
    label: "OpenRouter",
    description: "AI observability for OpenRouter with automatic OTLP trace export.",
    icon: "/openrouter.png",
    href: "/client/openrouter",
    badge: "production",
  },
  {
    value: "otel",
    label: "OpenTelemetry (otel)",
    description: "Send traces and metrics from any OTel-instrumented app to Traceway.",
    icon: "/otel.png",
    href: "/client/otel",
    badge: "production",
  },
  {
    value: "cloudflare",
    label: "Cloudflare Workers",
    description: "Cloudflare Workers with automatic request tracing via OTLP.",
    icon: "/cloudflare.png",
    href: "/client/cloudflare",
    badge: "production",
  },
  {
    value: "php-symfony",
    label: "Symfony",
    description: "Symfony framework with OpenTelemetry auto-instrumentation.",
    icon: "/symfony.png",
    href: "/client/symfony",
    badge: "production",
  },
  {
    value: "php-laravel",
    label: "Laravel",
    description: "Laravel framework with OpenTelemetry auto-instrumentation.",
    icon: "/laravel.png",
    href: "/client/laravel",
    badge: "production",
  },
  {
    value: "python-django",
    label: "Django",
    description: "Django framework with OpenTelemetry auto-instrumentation.",
    icon: "/django.png",
    href: "/client/django",
    badge: "new",
  },
  {
    value: "go-gin",
    label: "Go Gin",
    description:
      "Gin Gonic web framework with automatic request tracing and panic recovery.",
    icon: "/gin.png",
    href: "/client/gin-middleware",
    badge: "production",
  },
  {
    value: "go-chi",
    label: "Go Chi",
    description:
      "Lightweight Chi router with automatic request tracing and panic recovery.",
    icon: "/chi.png",
    href: "/client/chi-middleware",
    badge: "new",
  },
  {
    value: "go-fiber",
    label: "Go Fiber",
    description:
      "Express-inspired Fiber framework with request tracing and error capture.",
    icon: "/fiber.svg",
    href: "/client/fiber-middleware",
    badge: "new",
  },
  {
    value: "go-fasthttp",
    label: "Go FastHTTP",
    description:
      "High-performance FastHTTP server with request tracing and panic recovery.",
    icon: "/fasthttp.png",
    href: "/client/fasthttp-middleware",
    badge: "new",
  },
  {
    value: "go-http",
    label: "Go net/http",
    description:
      "Standard library HTTP middleware for request tracing and error capture.",
    icon: "/stdlib.png",
    href: "/client/http-middleware",
    badge: "new",
  },
  {
    value: "go-generic",
    label: "Go Generic",
    description:
      "Framework-agnostic SDK for manual instrumentation of any Go application.",
    icon: "/custom.png",
    href: "/client/sdk",
    badge: "new",
  },
  {
    value: "js-nextjs",
    label: "Next.js",
    description: "Next.js applications with OpenTelemetry auto-instrumentation.",
    icon: "/nextjs.png",
    href: "/client/nextjs",
    badge: "new",
  },
  {
    value: "js-node",
    label: "Node.js",
    description: "Node.js backend with OpenTelemetry traces and metrics.",
    icon: "/node.png",
    href: "/client/node-sdk",
    badge: "production",
  },
  {
    value: "js-nestjs",
    label: "NestJS",
    description: "NestJS framework with OpenTelemetry auto-instrumentation.",
    icon: "/nestjs.png",
    href: "/client/nestjs",
    badge: "production",
  },
  {
    value: "js-hono",
    label: "Hono",
    description: "Lightweight multi-runtime framework with OpenTelemetry.",
    icon: "/hono.png",
    href: "/client/hono",
    badge: "production",
  },
  {
    value: "js-react",
    label: "React",
    description: "React applications with error boundaries and hooks.",
    icon: "/react.png",
    href: "/client/react",
    badge: "production",
  },
  {
    value: "js-vue",
    label: "Vue.js",
    description: "Vue 3 applications with plugin and composables.",
    icon: "/vue.png",
    href: "/client/vue",
    badge: "new",
  },
  {
    value: "js-svelte",
    label: "Svelte",
    description: "Svelte/SvelteKit applications with context API.",
    icon: "/svelte.png",
    href: "/client/svelte",
    badge: "production",
  },
  {
    value: "js-jquery",
    label: "jQuery",
    description: "jQuery applications with automatic AJAX error capture.",
    icon: "/jquery.png",
    href: "/client/jquery",
    badge: "new",
  },
  {
    value: "js-generic",
    label: "JS Generic",
    description: "Framework-agnostic JavaScript SDK for browsers.",
    icon: "/javascript.png",
    href: "/client/js-sdk",
    badge: "production",
  },
  {
    value: "flutter",
    label: "Flutter",
    description: "Flutter mobile apps with automatic error capture and screen recording.",
    icon: "/flutter.png",
    href: "/client/flutter",
    badge: "production",
  },
  {
    value: "android",
    label: "Android",
    description: "Native Android (Kotlin/Java) apps with automatic exception capture, logs, HTTP, and navigation breadcrumbs.",
    icon: "/android.png",
    href: "/client/android",
    badge: "new",
  },
  {
    value: "react-native",
    label: "React Native",
    description: "React Native and Expo apps with automatic exception, fetch / XHR, and console capture. Works in Expo Go.",
    icon: "/react.png",
    href: "/client/react-native",
    badge: "new",
  },
];

export default function FrameworkPicker() {
  const router = useRouter();
  const { setSdk } = useSdk();

  function handleSelect(fw) {
    setSdk(fw.value);
    router.push(`${fw.href}?sdk=${fw.value}`);
  }

  return (
    <div className="framework-picker">
      <h2 className="framework-picker-heading">Choose your framework</h2>
      <p className="framework-picker-subheading">
        Select the framework you're using to get started with Traceway.
      </p>
      <div className="framework-picker-grid">
        {FRAMEWORKS.map((fw) => (
          <button
            key={fw.value}
            className="framework-picker-card"
            onClick={() => handleSelect(fw)}
          >
            <div className="framework-picker-top">
              <img
                src={fw.icon}
                alt={fw.label}
                className="framework-picker-icon"
              />
              <span className={`framework-picker-badge ${fw.badge === "production" ? "badge-production" : "badge-new"}`}>
                {fw.badge === "production" ? "Used in Production" : "New"}
              </span>
            </div>
            <span className="framework-picker-label">{fw.label}</span>
            <span className="framework-picker-desc">{fw.description}</span>
          </button>
        ))}
      </div>
    </div>
  );
}
