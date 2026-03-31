"use client";

import Image from "next/image";

const frameworks = [
  { name: "Gin", src: "/images/frameworks/gin.png" },
  { name: "Chi", src: "/images/frameworks/chi.png" },
  { name: "Stdlib", src: "/images/frameworks/stdlib.png" },
  { name: "FastHTTP", src: "/images/frameworks/fasthttp.png" },
  { name: "Express", src: "/images/frameworks/express.png" },
  { name: "NestJS", src: "/images/frameworks/nestjs.png" },
  { name: "Next.js", src: "/images/frameworks/nextjs.png" },
  { name: "Node.js", src: "/images/frameworks/node.png" },
  { name: "Svelte", src: "/images/frameworks/svelte.png" },
  { name: "Remix", src: "/images/frameworks/remix.png" },
  { name: "OpenTelemetry", src: "/images/frameworks/otel.png" },
  { name: "Cloudflare", src: "/images/frameworks/cloudflare.png" },
  { name: "jQuery", src: "/images/frameworks/jquery.png" },
  { name: "Symfony", src: "/images/frameworks/symfony.png" },
  { name: "OpenRouter", src: "/images/frameworks/openrouter.png" },
];

export function FrameworkMarquee() {
  return (
    <div
      className="relative overflow-hidden"
      style={{
        maskImage:
          "linear-gradient(to right, transparent, black 10%, black 90%, transparent)",
        WebkitMaskImage:
          "linear-gradient(to right, transparent, black 10%, black 90%, transparent)",
      }}
    >
      <style>{`
        @keyframes scroll-left {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
      `}</style>
      <div
        className="flex items-center gap-10"
        style={{
          width: "max-content",
          animation: "scroll-left 30s linear infinite",
        }}
      >
        {[...frameworks, ...frameworks].map((fw, i) => (
          <Image
            key={`${fw.name}-${i}`}
            src={fw.src}
            alt={fw.name}
            width={40}
            height={40}
            className="h-8 w-auto shrink-0"
          />
        ))}
      </div>
    </div>
  );
}
