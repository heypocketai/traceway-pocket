<?php

use App\Kernel;
use Symfony\Component\Dotenv\Dotenv;

require_once dirname(__DIR__) . '/vendor/autoload.php';

(new Dotenv())->loadEnv(dirname(__DIR__) . '/.env');
\OpenTelemetry\SDK\SdkAutoloader::autoload();

// Fix 1: Symfony's OTel auto-instrumentation sets http.route to the
// internal route name (e.g., "app_test_testparam") instead of the
// path template. This hook modifies _route on the request so Symfony's
// own terminate hook reads the corrected path.
//
// Fix 2: When a 500 occurs, Symfony renders the error page via a
// sub-request, which creates an extra OTel scope. The terminate hook
// then picks up the sub-request scope instead of the main request's,
// leaving the main span orphaned (never ended, never exported).
// We clean up sub-request scopes immediately to prevent this.
\OpenTelemetry\Instrumentation\hook(
    \Symfony\Component\HttpKernel\HttpKernel::class,
    'handle',
    post: static function (
        \Symfony\Component\HttpKernel\HttpKernel $kernel,
        array $params,
        mixed $returnValue,
        ?\Throwable $exception
    ): void {
        $request = ($params[0] instanceof \Symfony\Component\HttpFoundation\Request) ? $params[0] : null;
        if (null === $request) return;

        $type = $params[1] ?? \Symfony\Component\HttpKernel\HttpKernelInterface::MAIN_REQUEST;

        // Sub-requests (e.g., error page rendering) push an extra scope
        // that prevents the main request's span from being exported.
        // Clean up sub-request scopes immediately.
        if ($type === \Symfony\Component\HttpKernel\HttpKernelInterface::SUB_REQUEST) {
            $scope = \OpenTelemetry\Context\Context::storage()->scope();
            if (null !== $scope) {
                $span = \OpenTelemetry\API\Trace\Span::fromContext($scope->context());
                $scope->detach();
                $span->end();
            }
            return;
        }

        // For main requests: fix route name from internal name to path template
        $routeParams = $request->attributes->get('_route_params', []);
        $path = $request->getPathInfo();
        if (\is_array($routeParams)) {
            foreach ($routeParams as $name => $value) {
                if (\is_string($value) && '' !== $value) {
                    $path = str_replace($value, '{' . $name . '}', $path);
                }
            }
        }

        $request->attributes->set('_route', $path);
    }
);

$kernel = new Kernel('dev', true);
$request = \Symfony\Component\HttpFoundation\Request::createFromGlobals();
$response = $kernel->handle($request);
$response->send();
$kernel->terminate($request, $response);
