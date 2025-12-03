#!/bin/bash

# Test rate limiting functionality
# Rate limit is configured as 1000 requests per minute

echo "Testing rate limiting (1000 requests/minute)..."
echo "Making 50 requests quickly to trigger rate limit..."
echo

# Make 50 requests quickly (this should eventually trigger rate limiting)
for i in {1..50}; do
    echo -n "Request $i: "
    response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:9000/get)
    echo "HTTP $response"
    sleep 0.02  # Small delay between requests
done

echo
echo "Waiting 10 seconds for token refill..."
sleep 10

echo "Making another request after waiting (should work):"
response=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:9000/get)
echo "HTTP $response"

echo
echo "Rate limiting test complete!"
