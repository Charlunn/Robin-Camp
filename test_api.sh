#!/bin/bash

# Define the base URL for the API, assuming the default port is 8080
BASE_URL="http://localhost:8080"
MOVIE_TITLE="Inception"
RATER_ID="test-user-123"

# Function to print test headers
print_header() {
    echo ""
    echo "================================================="
    echo "$1"
    echo "================================================="
}

# --- Test Case 1: Create a Movie ---
print_header "Test 1: Create a Movie (POST /movies)"
CREATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/movies" \
-H "Content-Type: application/json" \
-d '{
  "title": "'"$MOVIE_TITLE"'",
  "genre": "Sci-Fi",
  "releaseDate": "2010-07-16"
}')

CREATE_BODY=$(echo "$CREATE_RESPONSE" | sed '$d')
CREATE_STATUS=$(echo "$CREATE_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)

echo "Status Code: $CREATE_STATUS"
echo "Response Body: $CREATE_BODY" | jq .

if [ "$CREATE_STATUS" -ne 201 ]; then
    echo "!!! TEST FAILED: Expected status 201, but got $CREATE_STATUS."
else
    echo "--- TEST PASSED ---"
fi


# --- Test Case 2: List Movies to verify creation ---
print_header "Test 2: List Movies (GET /movies)"
LIST_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET "${BASE_URL}/movies")

LIST_BODY=$(echo "$LIST_RESPONSE" | sed '$d')
LIST_STATUS=$(echo "$LIST_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)

echo "Status Code: $LIST_STATUS"
echo "Response Body: $LIST_BODY" | jq .

if [ "$LIST_STATUS" -ne 200 ]; then
    echo "!!! TEST FAILED: Expected status 200, but got $LIST_STATUS."
else
    # Check if the created movie is in the list
    MOVIE_FOUND=$(echo "$LIST_BODY" | jq '.items[] | select(.title=="'"$MOVIE_TITLE"'")')
    if [ -z "$MOVIE_FOUND" ]; then
        echo "!!! TEST FAILED: Newly created movie was not found in the list."
    else
        echo "--- TEST PASSED ---"
    fi
fi


# --- Test Case 3: Submit a Rating ---
print_header "Test 3: Submit a Rating (POST /movies/{title}/ratings)"
RATE_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X POST "${BASE_URL}/movies/${MOVIE_TITLE}/ratings" \
-H "Content-Type: application/json" \
-H "X-Rater-Id: ${RATER_ID}" \
-d '{
  "rating": 4.5
}')

RATE_BODY=$(echo "$RATE_RESPONSE" | sed '$d')
RATE_STATUS=$(echo "$RATE_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)

echo "Status Code: $RATE_STATUS"
echo "Response Body: $RATE_BODY" | jq .

if [ "$RATE_STATUS" -ne 201 ]; then
    echo "!!! TEST FAILED: Expected status 201, but got $RATE_STATUS."
else
    echo "--- TEST PASSED ---"
fi


# --- Test Case 4: Get Aggregated Rating ---
print_header "Test 4: Get Aggregated Rating (GET /movies/{title}/rating)"
AGG_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" -X GET "${BASE_URL}/movies/${MOVIE_TITLE}/rating")

AGG_BODY=$(echo "$AGG_RESPONSE" | sed '$d')
AGG_STATUS=$(echo "$AGG_RESPONSE" | grep "HTTP_STATUS" | cut -d':' -f2)

echo "Status Code: $AGG_STATUS"
echo "Response Body: $AGG_BODY" | jq .

if [ "$AGG_STATUS" -ne 200 ]; then
    echo "!!! TEST FAILED: Expected status 200, but got $AGG_STATUS."
else
    AVG_RATING=$(echo "$AGG_BODY" | jq '.average')
    COUNT=$(echo "$AGG_BODY" | jq '.count')
    if [ "$AVG_RATING" != "4.5" ] || [ "$COUNT" != "1" ]; then
        echo "!!! TEST FAILED: Aggregated rating is incorrect. Expected average 4.5 and count 1."
    else
        echo "--- TEST PASSED ---"
    fi
fi

echo ""
echo "================================================="
echo "API Validation Script Finished."
echo "================================================="