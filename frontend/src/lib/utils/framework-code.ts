import type { Framework } from '$lib/state/projects.svelte';
import { isJsFramework } from '$lib/state/projects.svelte';

export function getInstallCommand(framework: Framework): string {
	const base = 'go get go.tracewayapp.com';
	switch (framework) {
		case 'gin':
			return `${base} && go get go.tracewayapp.com/tracewaygin`;
		case 'chi':
			return `${base} && go get go.tracewayapp.com/tracewaychi`;
		case 'fiber':
			return `${base} && go get go.tracewayapp.com/tracewayfiber`;
		case 'fasthttp':
			return `${base} && go get go.tracewayapp.com/tracewayfasthttp`;
		case 'stdlib':
			return `${base} && go get go.tracewayapp.com/tracewayhttp`;
		case 'react':
			return 'npm install @tracewayapp/react';
		case 'svelte':
			return 'npm install @tracewayapp/svelte';
		case 'vuejs':
			return 'npm install @tracewayapp/vue';
		case 'nextjs':
			return 'npm install @tracewayapp/next';
		case 'nestjs':
			return 'npm install @tracewayapp/nest';
		case 'express':
			return 'npm install @tracewayapp/express';
		case 'remix':
			return 'npm install @tracewayapp/remix';
		case 'symfony':
			return 'composer require open-telemetry/sdk open-telemetry/exporter-otlp open-telemetry/opentelemetry-auto-symfony';
		case 'cloudflare':
			return '';
		case 'opentelemetry':
			return '';
		case 'custom':
		default:
			return base;
	}
}

export function getFrameworkCode(framework: Framework, token: string, backendUrl: string): string {
	const connectionString = token
		? `${token}@${backendUrl}/api/report`
		: `YOUR_TOKEN@${backendUrl}/api/report`;

	switch (framework) {
		case 'gin':
			return `package main

import (
    "github.com/gin-gonic/gin"
    tracewaygin "go.tracewayapp.com/tracewaygin"
)

func main() {
    r := gin.Default()
    r.Use(tracewaygin.New("${connectionString}"))
    r.Run(":8080")
}`;

		case 'chi':
			return `package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    tracewaychi "go.tracewayapp.com/tracewaychi"
)

func main() {
    r := chi.NewRouter()
    r.Use(tracewaychi.New("${connectionString}"))

    r.Get("/api/users", getUsers)
    http.ListenAndServe(":8080", r)
}`;

		case 'fiber':
			return `package main

import (
    "github.com/gofiber/fiber/v2"
    tracewayfiber "go.tracewayapp.com/tracewayfiber"
)

func main() {
    app := fiber.New()
    app.Use(tracewayfiber.New("${connectionString}"))

    app.Get("/api/users", getUsers)
    app.Listen(":8080")
}`;

		case 'fasthttp':
			return `package main

import (
    "github.com/valyala/fasthttp"
    tracewayfasthttp "go.tracewayapp.com/tracewayfasthttp"
)

func main() {
    handler := func(ctx *fasthttp.RequestCtx) {
        ctx.SetStatusCode(200)
        ctx.SetBodyString("Hello, World!")
    }

    tracedHandler := tracewayfasthttp.New("${connectionString}")(handler)
    fasthttp.ListenAndServe(":8080", tracedHandler)
}`;

		case 'stdlib':
			return `package main

import (
    "net/http"

    tracewayhttp "go.tracewayapp.com/tracewayhttp"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/users", getUsers)

    handler := tracewayhttp.New("${connectionString}")(mux)
    http.ListenAndServe(":8080", handler)
}`;

		case 'react':
			return `import { TracewayProvider } from "@tracewayapp/react";

function App() {
  return (
    <TracewayProvider connectionString="${connectionString}">
      <YourApp />
    </TracewayProvider>
  );
}

export default App;`;

		case 'svelte':
			return `<script>
  import { setupTraceway } from "@tracewayapp/svelte";

  setupTraceway({
    connectionString: "${connectionString}",
  });
</script>

<slot />`;

		case 'vuejs':
			return `import { createApp } from "vue";
import { createTracewayPlugin } from "@tracewayapp/vue";
import App from "./App.vue";

const app = createApp(App);

app.use(createTracewayPlugin({
  connectionString: "${connectionString}",
}));

app.mount("#app");`;

		case 'nextjs':
			return `import { withTraceway } from "@tracewayapp/next";

export default withTraceway({
    connectionString: "${connectionString}",
});`;

		case 'nestjs':
			return `import { Module } from "@nestjs/common";
import { TracewayModule } from "@tracewayapp/nest";

@Module({
    imports: [
        TracewayModule.forRoot({
            connectionString: "${connectionString}",
        }),
    ],
})
export class AppModule {}`;

		case 'express':
			return `import express from "express";
import { traceway } from "@tracewayapp/express";

const app = express();
app.use(traceway("${connectionString}"));

app.get("/api/users", (req, res) => {
    res.json({ users: [] });
});

app.listen(8080);`;

		case 'remix':
			return `import { withTraceway } from "@tracewayapp/remix";

export default withTraceway({
    connectionString: "${connectionString}",
});`;

		case 'symfony':
			return `<?php
// public/index.php

use App\\Kernel;

require_once dirname(__DIR__) . '/vendor/autoload.php';

\\OpenTelemetry\\SDK\\SdkAutoloader::autoload();

// Fixes for Symfony's OTel auto-instrumentation:
// 1. Corrects http.route from internal route name to URL path template
// 2. Cleans up sub-request scopes so 500 error spans are exported
\\OpenTelemetry\\Instrumentation\\hook(
    \\Symfony\\Component\\HttpKernel\\HttpKernel::class,
    'handle',
    post: static function (
        \\Symfony\\Component\\HttpKernel\\HttpKernel $kernel,
        array $params,
        mixed $returnValue,
        ?\\Throwable $exception
    ): void {
        $request = ($params[0] instanceof \\Symfony\\Component\\HttpFoundation\\Request) ? $params[0] : null;
        if (null === $request) return;

        $type = $params[1] ?? \\Symfony\\Component\\HttpKernel\\HttpKernelInterface::MAIN_REQUEST;

        if ($type === \\Symfony\\Component\\HttpKernel\\HttpKernelInterface::SUB_REQUEST) {
            $scope = \\OpenTelemetry\\Context\\Context::storage()->scope();
            if (null !== $scope) {
                $span = \\OpenTelemetry\\API\\Trace\\Span::fromContext($scope->context());
                $scope->detach();
                $span->end();
            }
            return;
        }

        $routeParams = $request->attributes->get('_route_params', []);
        $path = $request->getPathInfo();
        if (\\is_array($routeParams)) {
            foreach ($routeParams as $name => $value) {
                if (\\is_string($value) && '' !== $value) {
                    $path = str_replace($value, '{' . $name . '}', $path);
                }
            }
        }

        $request->attributes->set('_route', $path);
    }
);

$kernel = new Kernel($_SERVER['APP_ENV'] ?? 'dev', (bool) ($_SERVER['APP_DEBUG'] ?? true));
$request = \\Symfony\\Component\\HttpFoundation\\Request::createFromGlobals();
$response = $kernel->handle($request);
$response->send();
$kernel->terminate($request, $response);`;

		case 'cloudflare':
			return '';

		case 'opentelemetry':
			return '';

		case 'custom':
		default:
			return `package main

import (
    "go.tracewayapp.com"
)

func main() {
    traceway.Init(
        "${connectionString}",
        traceway.WithVersion("1.0.0"),
        traceway.WithServerName("my-server"),
    )
}`;
	}
}

export function getTestingRouteCode(framework?: Framework): string {
	if (framework === 'symfony') {
		return `<?php
// src/Controller/TestController.php
namespace App\\Controller;

use Symfony\\Component\\HttpFoundation\\Response;
use Symfony\\Component\\Routing\\Attribute\\Route;

class TestController
{
    #[Route('/testing', name: 'testing')]
    public function index(): Response
    {
        throw new \\RuntimeException("Test error from Traceway integration");
    }
}`;
	}
	if (framework && isJsFramework(framework)) {
		return `// Trigger a test error
throw new Error("Test error from Traceway integration");`;
	}
	return `r.GET("/testing", func(c *gin.Context) {
    panic("Test error from Traceway integration")
})`;
}

export function getTestingRouteCode2(framework?: Framework): string {
	if (framework === 'symfony') {
		return '';
	}
	if (framework && isJsFramework(framework)) {
		switch (framework) {
			case 'react':
				return `import { useTraceway } from "@tracewayapp/react";

// In a component using the hook
const { captureException } = useTraceway();
captureException(new Error("Test error"));`;
			case 'svelte':
				return `import { getTraceway } from "@tracewayapp/svelte";

const { captureException } = getTraceway();
captureException(new Error("Test error"));`;
			case 'vuejs':
				return `import { useTraceway } from "@tracewayapp/vue";

const { captureException } = useTraceway();
captureException(new Error("Test error"));`;
			default:
				return `import { captureException } from "@tracewayapp/${getPackageName(framework)}";

captureException(new Error("Test error"));`;
		}
	}
	return `r.GET("/testing", func(c *gin.Context) {
    c.AbortWithError(500, traceway.NewStackTraceErrorf("testing"))
})`;
}

function getPackageName(framework: Framework): string {
	switch (framework) {
		case 'react': return 'react';
		case 'svelte': return 'svelte';
		case 'vuejs': return 'vue';
		case 'nextjs': return 'next';
		case 'nestjs': return 'nest';
		case 'express': return 'express';
		case 'remix': return 'remix';
		default: return 'react';
	}
}

export function getFrameworkLabel(framework: Framework): string {
	const labels: Record<Framework, string> = {
		gin: 'Gin',
		fiber: 'Fiber',
		chi: 'Chi',
		fasthttp: 'FastHTTP',
		stdlib: 'Standard Library (net/http)',
		custom: 'Custom Integration',
		react: 'React',
		svelte: 'Svelte',
		vuejs: 'Vue.js',
		nextjs: 'Next.js',
		nestjs: 'NestJS',
		express: 'Express',
		remix: 'Remix',
		cloudflare: 'Cloudflare',
		opentelemetry: 'OpenTelemetry',
		symfony: 'Symfony',
	};
	return labels[framework] || framework;
}

export function getCodeLanguage(framework: Framework): 'go' | 'javascript' | 'bash' | 'php' {
	if (framework === 'symfony') return 'php';
	if (framework === 'opentelemetry') return 'go';
	if (framework === 'cloudflare') return 'javascript';
	return isJsFramework(framework) ? 'javascript' : 'go';
}
