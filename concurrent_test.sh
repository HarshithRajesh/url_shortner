#!/bin/bash
# filepath: /home/neo/code/projects/url_shortner/concurrent_test.sh

BASE_URL="http://localhost:8080"
CUSTOM_CODE="testrace"
CONCURRENT_REQUESTS=20

echo "🧪 Testing race conditions with $CONCURRENT_REQUESTS concurrent requests..."

# Create temp file for results
RESULTS_FILE=$(mktemp)

# Function to make request and capture result
make_request() {
  local index=$1
  local result=$(curl -s -X POST $BASE_URL/url \
    -H "Content-Type: application/json" \
    -d "{\"url\":\"https://test$index.com\",\"code\":\"$CUSTOM_CODE\"}")

  echo "$result" >>$RESULTS_FILE
}

# Launch concurrent requests
for i in $(seq 1 $CONCURRENT_REQUESTS); do
  make_request $i &
done

# Wait for all requests to complete
wait

echo "📊 Results:"
echo "Total responses: $(wc -l <$RESULTS_FILE)"

# Count successful requests (got the custom code)
SUCCESS_COUNT=$(grep -o "\"short_url\":\"$CUSTOM_CODE\"" $RESULTS_FILE | wc -l)
echo "Requests that got custom code '$CUSTOM_CODE': $SUCCESS_COUNT"

# Count requests that got random codes
RANDOM_COUNT=$(grep -v "\"short_url\":\"$CUSTOM_CODE\"" $RESULTS_FILE | grep "short_url" | wc -l)
echo "Requests that got random codes: $RANDOM_COUNT"

if [ $SUCCESS_COUNT -eq 1 ]; then
  echo "✅ PASS: Only one request got the custom code"
else
  echo "❌ FAIL: $SUCCESS_COUNT requests got the custom code (should be 1)"
fi

# Check for duplicate short URLs in responses
UNIQUE_SHORT_URLS=$(grep -o "\"short_url\":\"[^\"]*\"" $RESULTS_FILE | sort | uniq | wc -l)
TOTAL_SHORT_URLS=$(grep -o "\"short_url\":\"[^\"]*\"" $RESULTS_FILE | wc -l)

if [ $UNIQUE_SHORT_URLS -eq $TOTAL_SHORT_URLS ]; then
  echo "✅ PASS: No duplicate short URLs generated"
else
  echo "❌ FAIL: Found duplicate short URLs"
  echo "Unique: $UNIQUE_SHORT_URLS, Total: $TOTAL_SHORT_URLS"
fi

# Cleanup
rm $RESULTS_FILE
