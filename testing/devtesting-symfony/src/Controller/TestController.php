<?php

namespace App\Controller;

use App\Exception\CustomException;
use OpenTelemetry\API\Globals;
use OpenTelemetry\API\Trace\SpanKind;
use OpenTelemetry\API\Trace\StatusCode;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class TestController
{
    #[Route('/test-ok', methods: ['GET'])]
    public function testOk(): JsonResponse
    {
        usleep(random_int(0, 3000) * 1000);
        return new JsonResponse(['status' => 'ok']);
    }

    #[Route('/test-not-found', methods: ['GET'])]
    public function testNotFound(): JsonResponse
    {
        return new JsonResponse(['status' => 'not-found'], Response::HTTP_NOT_FOUND);
    }

    #[Route('/test-param/{param}', methods: ['GET'])]
    public function testParam(string $param): JsonResponse
    {
        return new JsonResponse(['param' => $param]);
    }

    #[Route('/test-spans', methods: ['GET'])]
    public function testSpans(): JsonResponse
    {
        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');

        $dbAndCacheSpan = $tracer->spanBuilder('db.and.cache')
            ->setSpanKind(SpanKind::KIND_INTERNAL)
            ->startSpan();
        $dbAndCacheScope = $dbAndCacheSpan->activate();

        $dbSpan = $tracer->spanBuilder('db.query')
            ->setSpanKind(SpanKind::KIND_INTERNAL)
            ->startSpan();
        usleep((50 + random_int(0, 100)) * 1000);
        $dbSpan->end();

        $cacheSpan = $tracer->spanBuilder('cache.set')
            ->setSpanKind(SpanKind::KIND_INTERNAL)
            ->startSpan();
        usleep((10 + random_int(0, 30)) * 1000);
        $cacheSpan->end();

        $httpSpan = $tracer->spanBuilder('http.external_api')
            ->setSpanKind(SpanKind::KIND_CLIENT)
            ->startSpan();
        usleep((100 + random_int(0, 200)) * 1000);
        $httpSpan->end();

        $dbAndCacheSpan->end();
        $dbAndCacheScope->detach();

        return new JsonResponse(['status' => 'ok', 'message' => 'Spans captured']);
    }

    #[Route('/test-exception/best', methods: ['GET'])]
    public function testExceptionTest(): never
    {
        usleep(random_int(0, 2000) * 1000);
        throw new \RuntimeException('Damn nice');
    }


    #[Route('/test-metrics', methods: ['GET'])]
    public function testMetrics(): JsonResponse
    {
        $meter = Globals::meterProvider()->getMeter('devtesting-symfony');

        $queues = ['emails', 'notifications', 'exports'];
        $priorities = ['high', 'low'];
        $queueSize = 50 + random_int(0, 450);

        $gauge = $meter->createObservableGauge('app.queue_size', 'items', 'Number of items in queue');
        $gauge->observe(function ($observer) use ($queueSize, $queues, $priorities) {
            $observer->observe($queueSize, [
                'queue' => $queues[array_rand($queues)],
                'priority' => $priorities[array_rand($priorities)],
            ]);
        });

        $caches = ['redis', 'memcached'];
        $operations = ['get', 'set', 'delete'];
        $latency = 5.0 + (random_int(0, 19500) / 100.0);

        $histogram = $meter->createHistogram('app.cache_latency_ms', 'ms', 'Cache operation latency');
        $histogram->record($latency, [
            'cache' => $caches[array_rand($caches)],
            'operation' => $operations[array_rand($operations)],
        ]);

        $hosts = ['web-01', 'web-02', 'web-03'];
        $cpuUsage = 10.0 + (random_int(0, 8500) / 100.0);

        $cpuGauge = $meter->createObservableGauge('app.cpu_usage', '%', 'CPU usage percentage');
        $cpuGauge->observe(function ($observer) use ($cpuUsage, $hosts) {
            $observer->observe($cpuUsage, [
                'host' => $hosts[array_rand($hosts)],
            ]);
        });

        $methods = ['GET', 'POST'];
        $statuses = ['200', '404', '500'];
        $requestCount = random_int(1, 50);

        $counter = $meter->createCounter('app.requests_total', 'requests', 'Total HTTP requests');
        $counter->add($requestCount, [
            'method' => $methods[array_rand($methods)],
            'status' => $statuses[array_rand($statuses)],
        ]);

        return new JsonResponse([
            'queue_size' => $queueSize,
            'cache_latency_ms' => $latency,
            'cpu_usage' => $cpuUsage,
            'requests_total' => $requestCount,
        ]);
    }

    #[Route('/test-message', methods: ['GET'])]
    public function testMessage(): JsonResponse
    {
        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-message-span')
            ->startSpan();
        $scope = $span->activate();

        for ($i = 0; $i < 10; $i++) {
            $span->addEvent("test message $i");
        }

        $span->setStatus(StatusCode::STATUS_ERROR, 'test message exception');
        $span->recordException(new \RuntimeException('test message exception'));
        $span->end();
        $scope->detach();

        return new JsonResponse(['status' => 'ok', 'messages' => 10]);
    }

    #[Route('/test-json', methods: ['GET'])]
    public function testJson(): JsonResponse
    {
        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-json-span')
            ->startSpan();

        $span->setAttribute('json_tag', '{"str": "traceway", "obj": {"id": 1}, "arr": [1, 2, 3]}');
        $span->addEvent('test json');
        $span->end();

        return new JsonResponse(['status' => 'ok']);
    }

    #[Route('/test-self-report-attributes', methods: ['GET'])]
    public function testSelfReportAttributes(): JsonResponse
    {
        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-self-report-attributes')
            ->startSpan();

        $span->setAttribute('Cool', '{"this": "is", "a": "json", "attr": 10}');
        $span->setAttribute('Cool2', 'Pretty cool2');
        $span->setAttribute('Cool3', 'Pretty cool3');
        $span->setStatus(StatusCode::STATUS_ERROR, 'Test');
        $span->recordException(new \RuntimeException('Test'));
        $span->end();

        return new JsonResponse(['status' => 'ok']);
    }

    #[Route('/test-cerror-simple', methods: ['GET'])]
    public function testCerrorSimple(): JsonResponse
    {
        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-cerror-simple')
            ->startSpan();

        $span->setStatus(StatusCode::STATUS_ERROR, 'simple error without stack');
        $span->recordException(new \RuntimeException('simple error without stack'));
        $span->end();

        return new JsonResponse(['error' => 'simple error'], Response::HTTP_INTERNAL_SERVER_ERROR);
    }

    #[Route('/test-cerror-wrapped', methods: ['GET'])]
    public function testCerrorWrapped(): JsonResponse
    {
        $base = new \RuntimeException('base error');
        $layer1 = new \RuntimeException('layer 1', 0, $base);
        $layer2 = new \RuntimeException('layer 2', 0, $layer1);

        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-cerror-wrapped')
            ->startSpan();

        $span->setStatus(StatusCode::STATUS_ERROR, $layer2->getMessage());
        $span->recordException($layer2);
        $span->end();

        return new JsonResponse(['error' => 'wrapped error'], Response::HTTP_INTERNAL_SERVER_ERROR);
    }

    #[Route('/test-cerror-custom', methods: ['GET'])]
    public function testCerrorCustom(): JsonResponse
    {
        $err = new CustomException(500, 'something went wrong');

        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-cerror-custom')
            ->startSpan();

        $span->setStatus(StatusCode::STATUS_ERROR, $err->getMessage());
        $span->recordException($err);
        $span->end();

        return new JsonResponse(['error' => 'custom error'], Response::HTTP_INTERNAL_SERVER_ERROR);
    }

    #[Route('/test-cerror-nested', methods: ['GET'])]
    public function testCerrorNested(): JsonResponse
    {
        $err = $this->outerFunction();

        $tracer = Globals::tracerProvider()->getTracer('devtesting-symfony');
        $span = $tracer->spanBuilder('test-cerror-nested')
            ->startSpan();

        $span->setStatus(StatusCode::STATUS_ERROR, $err->getMessage());
        $span->recordException($err);
        $span->end();

        return new JsonResponse(['error' => 'nested function error'], Response::HTTP_INTERNAL_SERVER_ERROR);
    }

    #[Route('/test-recording/{param}', methods: ['POST'])]
    public function testRecording(string $param, Request $request): JsonResponse
    {
        $data = json_decode($request->getContent(), true);

        if (!$data || !isset($data['name'])) {
            return new JsonResponse(['error' => 'missing name field'], Response::HTTP_BAD_REQUEST);
        }

        if ($data['name'] !== 'good') {
            throw new \RuntimeException('Bad');
        }

        return new JsonResponse(['status' => 'ok', 'param' => $param]);
    }

    private function outerFunction(): \RuntimeException
    {
        return $this->middleFunction();
    }

    private function middleFunction(): \RuntimeException
    {
        return $this->innerFunction();
    }

    private function innerFunction(): \RuntimeException
    {
        return new \RuntimeException('error from inner function');
    }
}
