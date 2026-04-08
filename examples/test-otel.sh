#!/bin/bash
# Test script for OTel example projects
# Usage: ./test-otel.sh [express|hono|nestjs|nextjs|all]

PASS=0
FAIL=0

check() {
  local fw="$1" method="$2" url="$3" expected="$4"
  local status
  status=$(curl -s -o /dev/null -w "%{http_code}" -X "$method" "$url" \
    -H "Content-Type: application/json" \
    ${5:+-d "$5"} 2>/dev/null)

  if [ "$status" = "$expected" ]; then
    printf "  %-8s %-6s %-40s %s -> %s\n" "$fw" "$method" "${url#http://localhost:*}" "$status" "PASS"
    PASS=$((PASS + 1))
  else
    printf "  %-8s %-6s %-40s %s (expected %s) -> %s\n" "$fw" "$method" "${url#http://localhost:*}" "$status" "$expected" "FAIL"
    FAIL=$((FAIL + 1))
  fi
}

test_framework() {
  local fw="$1" port="$2" prefix="$3"
  local base="http://localhost:${port}"

  # Check if server is reachable
  if ! curl -s -o /dev/null --connect-timeout 2 "$base/${prefix}/api/users" 2>/dev/null; then
    printf "  %-8s SKIPPED — server not reachable on port %s\n" "$fw" "$port"
    return
  fi

  echo "  --- $fw (port $port) ---"

  # List users
  check "$fw" GET "${base}/${prefix}/api/users" 200

  # Get user by ID (hit 2 different IDs to test route grouping)
  check "$fw" GET "${base}/${prefix}/api/users/1" 200
  check "$fw" GET "${base}/${prefix}/api/users/2" 200

  # Create user
  check "$fw" POST "${base}/${prefix}/api/users" 201 '{"name":"Test","email":"test@example.com"}'

  # Slow endpoint
  check "$fw" GET "${base}/${prefix}/api/slow" 200

  # Error endpoint
  check "$fw" GET "${base}/${prefix}/api/test-error" 500

  echo ""
}

target="${1:-all}"

echo ""
echo "OTel Example Test Suite"
echo "======================="
echo ""

case "$target" in
  express) test_framework "express" 3001 "express" ;;
  hono)    test_framework "hono"    3002 "hono" ;;
  nestjs)  test_framework "nestjs"  3003 "nestjs" ;;
  nextjs)  test_framework "nextjs"  3004 "nextjs" ;;
  all)
    test_framework "express" 3001 "express"
    test_framework "hono"    3002 "hono"
    test_framework "nestjs"  3003 "nestjs"
    test_framework "nextjs"  3004 "nextjs"
    ;;
  *) echo "Usage: $0 [express|hono|nestjs|nextjs|all]"; exit 1 ;;
esac

echo "Results: $PASS passed, $FAIL failed"
