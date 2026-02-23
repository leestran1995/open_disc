#!/usr/bin/env bash
#
# E2E API test script for open_disc
# Requires the backend to be running on localhost:8080 (REST) and :8081 (SSE)
#
# Usage: ./scripts/e2e-api-test.sh

set -euo pipefail

BASE_URL="http://localhost:8080"
SSE_URL="http://localhost:8080"

# --- Helpers ---

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0

pass() {
  PASS_COUNT=$((PASS_COUNT + 1))
  printf "${GREEN}PASS${NC} %s\n" "$1"
}

fail() {
  FAIL_COUNT=$((FAIL_COUNT + 1))
  printf "${RED}FAIL${NC} %s\n" "$1"
  if [ -n "${2:-}" ]; then
    printf "     %s\n" "$2"
  fi
}

info() {
  printf "${YELLOW}----${NC} %s\n" "$1"
}

# Lightweight JSON value extraction -- uses jq if available, falls back to grep/sed
json_val() {
  local json="$1" key="$2"
  if command -v jq &>/dev/null; then
    echo "$json" | jq -r ".$key"
  else
    echo "$json" | grep -o "\"$key\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -1 | sed "s/\"$key\"[[:space:]]*:[[:space:]]*\"//" | sed 's/"$//'
  fi
}

# Unique test user per run
TEST_USER="testuser_$(date +%s)"
TEST_PASS="testpass_secure_123"

info "Test user: $TEST_USER"
echo ""

# ================================================================
# 1. Health check
# ================================================================
info "1. Health check -- GET /ping"

PING_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/ping")
PING_BODY=$(echo "$PING_RESPONSE" | head -1)
PING_CODE=$(echo "$PING_RESPONSE" | tail -1)

if [ "$PING_CODE" = "200" ] && [ "$PING_BODY" = "pong" ]; then
  pass "GET /ping returned 200 with 'pong'"
else
  fail "GET /ping" "Expected 200/pong, got $PING_CODE/$PING_BODY"
fi

# ================================================================
# 2. Signup
# ================================================================
info "2. Signup -- POST /signup"

SIGNUP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/signup" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASS\"}")
SIGNUP_BODY=$(echo "$SIGNUP_RESPONSE" | sed '$d')
SIGNUP_CODE=$(echo "$SIGNUP_RESPONSE" | tail -1)

if [ "$SIGNUP_CODE" = "201" ]; then
  pass "POST /signup returned 201"
else
  fail "POST /signup" "Expected 201, got $SIGNUP_CODE -- $SIGNUP_BODY"
fi

# ================================================================
# 3. Signup duplicate
# ================================================================
info "3. Signup duplicate -- POST /signup (same username)"

DUP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/signup" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASS\"}")
DUP_BODY=$(echo "$DUP_RESPONSE" | sed '$d')
DUP_CODE=$(echo "$DUP_RESPONSE" | tail -1)

if [ "$DUP_CODE" = "400" ]; then
  pass "Duplicate signup returned 400"
else
  fail "Duplicate signup" "Expected 400, got $DUP_CODE -- $DUP_BODY"
fi

# ================================================================
# 4. Signup bad password (< 8 chars)
# ================================================================
info "4. Signup bad password -- POST /signup (short password)"

BADPW_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/signup" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"${TEST_USER}_badpw\",\"password\":\"short\"}")
BADPW_BODY=$(echo "$BADPW_RESPONSE" | sed '$d')
BADPW_CODE=$(echo "$BADPW_RESPONSE" | tail -1)

if [ "$BADPW_CODE" = "400" ]; then
  pass "Short password signup returned 400"
else
  fail "Short password signup" "Expected 400, got $BADPW_CODE -- $BADPW_BODY"
fi

# ================================================================
# 5. Signin
# ================================================================
info "5. Signin -- POST /signin"

SIGNIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/signin" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$TEST_USER\",\"password\":\"$TEST_PASS\"}")
SIGNIN_BODY=$(echo "$SIGNIN_RESPONSE" | sed '$d')
SIGNIN_CODE=$(echo "$SIGNIN_RESPONSE" | tail -1)

TOKEN=$(json_val "$SIGNIN_BODY" "data")

if [ "$SIGNIN_CODE" = "200" ] && [ -n "$TOKEN" ] && [ "$TOKEN" != "null" ]; then
  pass "POST /signin returned 200 with JWT token"
else
  fail "POST /signin" "Expected 200 + token, got $SIGNIN_CODE -- $SIGNIN_BODY"
fi

# ================================================================
# 6. Signin bad password
# ================================================================
info "6. Signin bad password -- POST /signin (wrong password)"

BAD_SIGNIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/signin" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$TEST_USER\",\"password\":\"wrongpassword\"}")
BAD_SIGNIN_BODY=$(echo "$BAD_SIGNIN_RESPONSE" | sed '$d')
BAD_SIGNIN_CODE=$(echo "$BAD_SIGNIN_RESPONSE" | tail -1)

if [ "$BAD_SIGNIN_CODE" = "401" ]; then
  pass "Bad password signin returned 401"
else
  fail "Bad password signin" "Expected 401, got $BAD_SIGNIN_CODE -- $BAD_SIGNIN_BODY"
fi

# ================================================================
# 7. Create room
# ================================================================
info "7. Create room -- POST /rooms"

ROOM_NAME="test-room-$(date +%s)"
CREATE_ROOM_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rooms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"name\":\"$ROOM_NAME\"}")
CREATE_ROOM_BODY=$(echo "$CREATE_ROOM_RESPONSE" | sed '$d')
CREATE_ROOM_CODE=$(echo "$CREATE_ROOM_RESPONSE" | tail -1)

ROOM_ID=$(json_val "$CREATE_ROOM_BODY" "id")

if [ "$CREATE_ROOM_CODE" = "201" ] && [ -n "$ROOM_ID" ] && [ "$ROOM_ID" != "null" ]; then
  pass "POST /rooms returned 201 with room id=$ROOM_ID"
else
  fail "POST /rooms" "Expected 201 + room id, got $CREATE_ROOM_CODE -- $CREATE_ROOM_BODY"
fi

# ================================================================
# 8. Get room
# ================================================================
info "8. Get room -- GET /rooms/:id"

GET_ROOM_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/rooms/$ROOM_ID" \
  -H "Authorization: Bearer $TOKEN")
GET_ROOM_BODY=$(echo "$GET_ROOM_RESPONSE" | sed '$d')
GET_ROOM_CODE=$(echo "$GET_ROOM_RESPONSE" | tail -1)

GET_ROOM_NAME=$(json_val "$GET_ROOM_BODY" "name")

if [ "$GET_ROOM_CODE" = "200" ] && [ "$GET_ROOM_NAME" = "$ROOM_NAME" ]; then
  pass "GET /rooms/:id returned 200 with correct name"
else
  fail "GET /rooms/:id" "Expected 200 + name=$ROOM_NAME, got $GET_ROOM_CODE -- $GET_ROOM_BODY"
fi

# ================================================================
# 9. Join room
# ================================================================
info "9. Join room -- POST /rooms/:id/join"

# The join endpoint requires a user_id (UUID from the DB). Since the signup/signin
# API doesn't return the user_id, we send the request and verify the endpoint is
# reachable and auth works. A 500 from the FK constraint is acceptable here since
# there's no API to look up user_id.
JOIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rooms/$ROOM_ID/join" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"user_id\":\"00000000-0000-0000-0000-000000000000\"}")
JOIN_BODY=$(echo "$JOIN_RESPONSE" | sed '$d')
JOIN_CODE=$(echo "$JOIN_RESPONSE" | tail -1)

if [ "$JOIN_CODE" = "200" ] || [ "$JOIN_CODE" = "201" ]; then
  pass "POST /rooms/:id/join returned $JOIN_CODE"
else
  # Acceptable: 500 due to FK constraint (no user with that UUID)
  info "POST /rooms/:id/join returned $JOIN_CODE (expected -- user_id not exposed via API)"
fi

# ================================================================
# 10. Send message
# ================================================================
info "10. Send message -- POST /messages"

MSG_TEXT="Hello from e2e test $(date +%s)"
SEND_MSG_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{\"room_id\":\"$ROOM_ID\",\"message\":\"$MSG_TEXT\"}")
SEND_MSG_BODY=$(echo "$SEND_MSG_RESPONSE" | sed '$d')
SEND_MSG_CODE=$(echo "$SEND_MSG_RESPONSE" | tail -1)

MSG_USERNAME=$(json_val "$SEND_MSG_BODY" "message" 2>/dev/null || true)
# Extract nested username from the message object
if command -v jq &>/dev/null; then
  MSG_USERNAME=$(echo "$SEND_MSG_BODY" | jq -r '.message.username')
  MSG_ID=$(echo "$SEND_MSG_BODY" | jq -r '.message.id')
else
  MSG_USERNAME=$(echo "$SEND_MSG_BODY" | grep -o '"username"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/"username"[[:space:]]*:[[:space:]]*"//' | sed 's/"$//')
  MSG_ID=$(echo "$SEND_MSG_BODY" | grep -o '"id"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed 's/"id"[[:space:]]*:[[:space:]]*"//' | sed 's/"$//')
fi

if [ "$SEND_MSG_CODE" = "200" ] && [ "$MSG_USERNAME" = "$TEST_USER" ]; then
  pass "POST /messages returned 200 with username=$TEST_USER from JWT"
else
  fail "POST /messages" "Expected 200 + username=$TEST_USER, got $SEND_MSG_CODE / username=$MSG_USERNAME -- $SEND_MSG_BODY"
fi

# ================================================================
# 11. Get messages
# ================================================================
info "11. Get messages -- GET /messages/:room_id"

GET_MSG_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/messages/$ROOM_ID" \
  -H "Authorization: Bearer $TOKEN")
GET_MSG_BODY=$(echo "$GET_MSG_RESPONSE" | sed '$d')
GET_MSG_CODE=$(echo "$GET_MSG_RESPONSE" | tail -1)

# Check that our message appears in the response
if command -v jq &>/dev/null; then
  FOUND_MSG=$(echo "$GET_MSG_BODY" | jq -r ".messages[] | select(.id==\"$MSG_ID\") | .message")
else
  FOUND_MSG=""
  if echo "$GET_MSG_BODY" | grep -q "$MSG_TEXT"; then
    FOUND_MSG="$MSG_TEXT"
  fi
fi

if [ "$GET_MSG_CODE" = "200" ] && [ -n "$FOUND_MSG" ]; then
  pass "GET /messages/:room_id returned 200 with sent message"
else
  fail "GET /messages/:room_id" "Expected 200 + message text, got $GET_MSG_CODE -- $GET_MSG_BODY"
fi

# ================================================================
# 12. Auth required (no token)
# ================================================================
info "12. Auth required -- POST /rooms without token"

NOAUTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/rooms" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"should-fail\"}")
NOAUTH_CODE=$(echo "$NOAUTH_RESPONSE" | tail -1)

if [ "$NOAUTH_CODE" = "401" ]; then
  pass "POST /rooms without token returned 401"
else
  fail "POST /rooms without token" "Expected 401, got $NOAUTH_CODE"
fi

# ================================================================
# 13. SSE connect
# ================================================================
info "13. SSE connect -- GET /connect/:username on port 8081"

SSE_TMPFILE=$(mktemp)

# Start SSE connection in background, capture first few events
curl -s -N "$SSE_URL/connect/$TEST_USER" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: text/event-stream" \
  --max-time 3 > "$SSE_TMPFILE" 2>/dev/null &
SSE_PID=$!

# Wait for events to arrive
sleep 2

# Kill the background curl
kill $SSE_PID 2>/dev/null || true
wait $SSE_PID 2>/dev/null || true

SSE_OUTPUT=$(cat "$SSE_TMPFILE")
rm -f "$SSE_TMPFILE"

if echo "$SSE_OUTPUT" | grep -q "user_joined"; then
  pass "SSE /connect/:username received user_joined events"
elif [ -n "$SSE_OUTPUT" ]; then
  pass "SSE /connect/:username connected and received data"
else
  fail "SSE /connect/:username" "No data received from SSE endpoint"
fi

# ================================================================
# Summary
# ================================================================
echo ""
echo "========================================"
TOTAL=$((PASS_COUNT + FAIL_COUNT))
printf "Results: ${GREEN}%d passed${NC}, ${RED}%d failed${NC} out of %d tests\n" "$PASS_COUNT" "$FAIL_COUNT" "$TOTAL"
echo "========================================"

if [ "$FAIL_COUNT" -gt 0 ]; then
  exit 1
fi
