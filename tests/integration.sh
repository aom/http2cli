#!/bin/sh

set -e

echo "=== http2cli Integration Tests ==="

# Start server in background
echo "Starting server..."
http2cli --config /etc/http2cli/config.yaml &
SERVER_PID=$!
sleep 2

# Cleanup on exit
cleanup() {
    echo "Stopping server..."
    kill $SERVER_PID 2>/dev/null || true
}
trap cleanup EXIT

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Test helper
test_case() {
    name="$1"
    cmd="$2"

    echo -n "  Testing: $name... "
    if eval "$cmd" > /dev/null 2>&1; then
        echo "OK"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo "FAILED"
        TESTS_FAILED=$((TESTS_FAILED + 1))
    fi
}

echo ""
echo "--- Health Endpoints ---"
test_case "GET /health" \
    'test "$(curl -sf http://localhost:8080/health)" = "{\"status\":\"ok\"}"'

test_case "GET /ready" \
    'test "$(curl -sf http://localhost:8080/ready)" = "{\"status\":\"ready\"}"'

test_case "GET /api/tools" \
    'curl -sf http://localhost:8080/api/tools | grep -q "\"name\":\"cat\"" && \
     curl -sf http://localhost:8080/api/tools | grep -q "\"name\":\"date\"" && \
     curl -sf http://localhost:8080/api/tools | grep -q "\"name\":\"rev\"" && \
     curl -sf http://localhost:8080/api/tools | grep -q "\"name\":\"cat-file\""'

echo ""
echo "--- Date Tool ---"
test_case "GET /date returns date" \
    'curl -sf "http://localhost:8080/date" | grep -qE "^[0-9]{4}-[0-9]{2}-[0-9]{2}"'

test_case "GET /date with format" \
    'curl -sf "http://localhost:8080/date?format=%25Y" | grep -qE "^[0-9]{4}"'

test_case "GET /date with UTC flag" \
    'curl -sf "http://localhost:8080/date?utc=true&format=%25Z" | grep -q "UTC"'

echo ""
echo "--- Cat Tool ---"
test_case "POST /cat echoes input" \
    'echo "hello world" | curl -sf -F "file=@-" http://localhost:8080/cat | grep -q "hello world"'

test_case "POST /cat handles binary" \
    'printf "\x00\x01\x02" | curl -sf -F "file=@-" http://localhost:8080/cat | wc -c | grep -q "3"'

echo ""
echo "--- Rev Tool ---"
test_case "POST /rev reverses text" \
    'echo "hello" | curl -sf -F "text=@-" http://localhost:8080/rev | grep -q "olleh"'

test_case "POST /rev handles multiple lines" \
    'printf "abc\ndef" | curl -sf -F "text=@-" http://localhost:8080/rev | grep -q "cba"'

echo ""
echo "--- Cat-File Tool (file input type) ---"
test_case "POST /cat-file echoes input" \
    'curl -sf -F "file=@/tests/support/text.txt" http://localhost:8080/cat-file | grep -q "Here be text"'

test_case "POST /cat-file handles binary" \
    'printf "\x00\x01\x02" | curl -sf -F "file=@-" http://localhost:8080/cat-file | wc -c | grep -q "3"'

echo ""
echo "--- Error Cases ---"
test_case "wrong method returns 405" \
    'test "$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/date)" = "405"'

test_case "missing file returns 400" \
    'test "$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8080/cat)" = "400"'

echo ""
echo "=== Results ==="
echo "Passed: $TESTS_PASSED"
echo "Failed: $TESTS_FAILED"

if [ $TESTS_FAILED -gt 0 ]; then
    echo "TESTS FAILED!"
    exit 1
fi

echo "ALL TESTS PASSED!"
exit 0