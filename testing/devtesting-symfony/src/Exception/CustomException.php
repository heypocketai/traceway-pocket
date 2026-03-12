<?php

namespace App\Exception;

class CustomException extends \RuntimeException
{
    public function __construct(
        int $code,
        string $message,
        ?\Throwable $previous = null,
    ) {
        parent::__construct("CustomException[$code]: $message", $code, $previous);
    }
}
