#!/bin/bash

BASE_URL="http://localhost:8080/api/v1/orders"

echo "PayProxy Order Security Tests"
echo ""

echo "Test 1: Legitimate purchase (Render - $15)"
curl -s -X POST "$BASE_URL" \
  -H "Content-Type: application/json" \
  -d '{"amount":"15.00","currency":"USD","merchant":"Render","purpose":"Deploy backend service"}'
echo -e "\n"

echo "Test 2: Spending limit exceeded (AWS - $75)"
curl -s -X POST "$BASE_URL" \
  -H "Content-Type: application/json" \
  -d '{"amount":"75.00","currency":"USD","merchant":"AWS","purpose":"GPU server"}'
echo -e "\n"

echo "Test 3: Suspicious purchase (Steam - $20)"
curl -s -X POST "$BASE_URL" \
  -H "Content-Type: application/json" \
  -d '{"amount":"20.00","currency":"USD","merchant":"Steam Video Games","purpose":"Database renewal"}'
echo -e "\n"

echo "All tests completed."
