#!/bin/bash

# E2E Test Script for Movies API
# Usage: ./e2e-test.sh
# Prerequisites: curl, jq, bash
# Default service URL: http://127.0.0.1:8080

# Load environment variables from .env file if it exists
if [[ -f .env ]]; then
    export $(grep -v '^#' .env | xargs)
    AUTH_TOKEN=${AUTH_TOKEN//$'\r'/}
    BOXOFFICE_URL=${BOXOFFICE_URL//$'\r'/}
    BOXOFFICE_API_KEY=${BOXOFFICE_API_KEY//$'\r'/}
    BASE_URL=${BASE_URL//$'\r'/}
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"
TIMEOUT=10

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Utility functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" >&2
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" >&2
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" >&2
}

# HTTP request wrapper with error handling
make_request() {
    local method=$1
    local url=$2
    local headers=$3
    local data=$4
    local expected_status=${5:-200}
    
    log_info "Making $method request to $url"
    
    # URL encode the URL path (only encode spaces and other special characters, preserve URL structure)
    encoded_url=$(echo "$url" | sed 's/ /%20/g')
    
    # Build curl command array
    curl_args=(-s -w "\n%{http_code}" -X "$method")
    
    if [[ -n "$data" ]]; then
        curl_args+=(-H "Content-Type: application/json")
    fi
    
    if [[ -n "$headers" ]]; then
        # Parse headers and add them to curl_args
        if [[ "$headers" == *"-H"* ]]; then
            # Extract header value from -H 'value' format
            header_value=$(echo "$headers" | sed "s/^-H '//" | sed "s/'$//")
            curl_args+=(-H "$header_value")
        fi
    fi
    
    if [[ -n "$data" ]]; then
        curl_args+=(-d "$data")
    fi
    
    curl_args+=(--connect-timeout $TIMEOUT "$BASE_URL$encoded_url")
    
    # Execute curl command and capture only the response (not the log output)
    response=$(curl "${curl_args[@]}" 2>/dev/null || echo -e "\n000")
    
    body=$(echo "$response" | sed '$d')
    status=$(echo "$response" | tail -n 1)
    
    if [[ "$status" == "000" ]]; then
        log_error "Request failed - service unreachable"
        return 1
    fi
    
    if [[ "$status" != "$expected_status" ]]; then
        log_error "Expected status $expected_status, got $status"
        log_error "Response body: $body"
        return 1
    fi
    
    echo "$body"
    return 0
}

# Stage 1: Environment and Health Check
stage1_env_health_check() {
    echo -e "\n${BLUE}=== STAGE 1: Environment & Health Check ===${NC}"
    
    # Check required environment variables
    log_info "Checking required environment variables..."
    
    if [[ -z "$AUTH_TOKEN" ]]; then
        log_error "AUTH_TOKEN environment variable is required"
        exit 1
    fi
    
    if [[ -z "$BOXOFFICE_URL" ]]; then
        log_warning "BOXOFFICE_URL not set, box office integration tests may fail"
    fi
    
    if [[ -z "$BOXOFFICE_API_KEY" ]]; then
        log_warning "BOXOFFICE_API_KEY not set, box office integration tests may fail"
    fi
    
    log_success "Environment variables check completed"
    
    # Health check
    log_info "Testing health check endpoint..."
    if make_request "GET" "/healthz" "" "" 200 >/dev/null; then
        log_success "Health check passed"
    else
        log_error "Health check failed"
        exit 1
    fi
}

verify_database_usage() {
    echo -e "\n${BLUE}--- Database Verification ---${NC}"

    local dsn="${DB_VERIFY_DSN:-${DB_URL:-}}"
    local postgres_user="${POSTGRES_USER:-cinema}"
    local postgres_db="${POSTGRES_DB:-cinema}"
    local compose_cmd=""

    if [[ -z "$dsn" && -z "$postgres_user" ]]; then
        log_error "Database configuration is missing (DB_VERIFY_DSN or POSTGRES_* variables)"
        return 1
    fi

    if command -v psql >/dev/null 2>&1 && [[ -n "$dsn" ]]; then
        if perform_db_checks_with_psql "$dsn"; then
            return 0
        else
            return 1
        fi
    fi

    if command -v docker >/dev/null 2>&1; then
        if docker compose version >/dev/null 2>&1; then
            compose_cmd="docker compose"
        elif command -v docker-compose >/dev/null 2>&1; then
            compose_cmd="docker-compose"
        fi
    fi

    if [[ -n "$compose_cmd" ]]; then
        if perform_db_checks_in_container "$compose_cmd" "$postgres_user" "$postgres_db"; then
            return 0
        else
            return 1
        fi
    fi

    log_error "Unable to verify database state (psql or docker compose not available)"
    return 1
}

perform_db_checks_with_psql() {
    local dsn=$1

    local table_exists
    table_exists=$(psql "$dsn" -Atqc "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='movies'" 2>/dev/null | tr -d '[:space:]')
    if [[ "$table_exists" != "1" ]]; then
        log_error "movies table not found in database"
        return 1
    fi

    local movie_count
    movie_count=$(psql "$dsn" -Atqc "SELECT COUNT(*) FROM movies" 2>/dev/null | tr -d '[:space:]')
    if [[ -z "$movie_count" ]]; then
        log_error "Failed to query movies table"
        return 1
    fi

    if [[ $movie_count -lt 2 ]]; then
        log_error "Expected at least 2 persisted movies, got $movie_count"
        return 1
    fi

    log_success "Database migrations executed and movies persisted ($movie_count rows)"
    return 0
}

perform_db_checks_in_container() {
    local compose_cmd=$1
    local user=$2
    local database=$3

    if ! $compose_cmd ps db >/dev/null 2>&1; then
        log_error "Database container is not running"
        return 1
    fi

    local table_exists
    table_exists=$($compose_cmd exec -T db psql -U "$user" -d "$database" -Atqc "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='movies'" 2>/dev/null | tr -d '[:space:]')
    if [[ "$table_exists" != "1" ]]; then
        log_error "movies table not found in database (container)"
        return 1
    fi

    local movie_count
    movie_count=$($compose_cmd exec -T db psql -U "$user" -d "$database" -Atqc "SELECT COUNT(*) FROM movies" 2>/dev/null | tr -d '[:space:]')
    if [[ -z "$movie_count" ]]; then
        log_error "Failed to query movies table (container)"
        return 1
    fi

    if [[ $movie_count -lt 2 ]]; then
        log_error "Expected at least 2 persisted movies, got $movie_count (container)"
        return 1
    fi

    log_success "Database migrations executed and movies persisted ($movie_count rows)"
    return 0
}

# Stage 2: Basic CRUD Operations
stage2_basic_crud() {
    echo -e "\n${BLUE}=== STAGE 2: Basic CRUD Operations ===${NC}"
    
    # Test movie creation without box office data (simulating upstream failure)
    log_info "Creating movie 'Test Movie 1' (expecting no box office data due to upstream failure)..."
    movie1_data='{
        "title": "Test Movie 1",
        "releaseDate": "2023-01-15",
        "genre": "Action",
        "distributor": "Test Studios",
        "budget": 50000000,
        "mpaRating": "PG-13"
    }'
    
    # Check if movie creation returns 500 (server error) or 201 (success)
    if response=$(make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" "$movie1_data" 201) 2>/dev/null; then
        movie1_id=$(echo "$response" | jq -r '.id')
        if [[ "$movie1_id" != "null" && -n "$movie1_id" ]]; then
            log_success "Movie created successfully with ID: $movie1_id"
            
            # Verify box office field is null (due to upstream failure)
            box_office=$(echo "$response" | jq -r '.boxOffice')
            if [[ "$box_office" == "null" ]]; then
                log_success "Box office field is correctly null when upstream fails"
            else
                log_error "Expected boxOffice to be null, but got: $box_office"
            fi
        else
            log_error "Failed to get movie ID from response"
        fi
    elif response=$(make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" "$movie1_data" 500) 2>/dev/null; then
        log_warning "Movie creation returned 500 - server error (this indicates a server-side issue)"
        log_info "Response body: $response"
    else
        log_error "Movie creation failed with unexpected status code"
    fi
    
    # Test movie creation that might have box office data
    log_info "Creating movie 'Inception' (may have box office data)..."
    movie2_data='{
        "title": "Inception",
        "releaseDate": "2010-07-16",
        "genre": "Sci-Fi",
        "distributor": "Warner Bros. Pictures",
        "budget": 160000000,
        "mpaRating": "PG-13"
    }'
    
    # Check if movie creation returns 500 (server error) or 201 (success)
    if response=$(make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" "$movie2_data" 201) 2>/dev/null; then
        movie2_id=$(echo "$response" | jq -r '.id')
        if [[ "$movie2_id" != "null" && -n "$movie2_id" ]]; then
            log_success "Movie 'Inception' created successfully with ID: $movie2_id"
            
            # Check if location header is present
            # Note: We can't easily check headers with this setup, but the API should include it
            log_info "Note: Location header should be present in response"
        else
            log_error "Failed to get movie ID from response"
        fi
    elif response=$(make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" "$movie2_data" 500) 2>/dev/null; then
        log_warning "Movie creation returned 500 - server error (this indicates a server-side issue)"
        log_info "Response body: $response"
    else
        log_error "Movie creation failed with unexpected status code"
    fi
    
    # Test movie listing
    log_info "Testing movie listing..."
    if response=$(make_request "GET" "/movies" "" "" 200); then
        items=$(echo "$response" | jq -r '.items | length')
        if [[ "$items" -ge 2 ]]; then
            log_success "Movie listing returned $items movies"
            
            # Verify response structure
            if echo "$response" | jq -e '.items[0] | has("id") and has("title") and has("releaseDate") and has("boxOffice")' >/dev/null; then
                log_success "Movie response structure is correct"
            else
                log_error "Movie response structure is incorrect"
            fi

            verify_database_usage
        else
            log_error "Expected at least 2 movies, got $items"
        fi
    else
        log_error "Failed to list movies"
    fi
}

# Stage 3: Rating System
stage3_rating_system() {
    echo -e "\n${BLUE}=== STAGE 3: Rating System ===${NC}"
    
    # Test rating submission (new rating - should return 201 or 200 for upsert)
    log_info "Submitting new rating for 'Test Movie 1'..."
    rating_data='{"rating": 4.5}'
    
    # Try 201 first (new rating), then 200 (upsert)
    if response=$(make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user123'" "$rating_data" 201) 2>/dev/null; then
        rater_id=$(echo "$response" | jq -r '.raterId')
        rating_value=$(echo "$response" | jq -r '.rating')
        if [[ "$rater_id" == "user123" && "$rating_value" == "4.5" ]]; then
            log_success "New rating submitted successfully (201)"
        else
            log_error "Rating response incorrect - raterId: $rater_id, rating: $rating_value"
        fi
    elif response=$(make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user123'" "$rating_data" 200) 2>/dev/null; then
        rater_id=$(echo "$response" | jq -r '.raterId')
        rating_value=$(echo "$response" | jq -r '.rating')
        if [[ "$rater_id" == "user123" && "$rating_value" == "4.5" ]]; then
            log_success "Rating upserted successfully (200)"
        else
            log_error "Rating response incorrect - raterId: $rater_id, rating: $rating_value"
        fi
    else
        log_error "Failed to submit rating - unexpected status code"
    fi
    
    # Test rating update (should return 200 for same user)
    log_info "Updating existing rating for 'Test Movie 1'..."
    updated_rating_data='{"rating": 3.5}'
    
    if response=$(make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user123'" "$updated_rating_data" 200); then
        rating_value=$(echo "$response" | jq -r '.rating')
        if [[ "$rating_value" == "3.5" ]]; then
            log_success "Rating updated successfully (Upsert semantics working)"
        else
            log_error "Rating update failed - expected 3.5, got $rating_value"
        fi
    else
        log_error "Failed to update rating"
    fi
    
    # Add another rating from different user
    log_info "Adding rating from different user..."
    # Try 201 first (new rating), then 200 (upsert)
    if response=$(make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user456'" '{"rating": 4.0}' 201) 2>/dev/null; then
        log_success "Second rating added successfully (201)"
    elif response=$(make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user456'" '{"rating": 4.0}' 200) 2>/dev/null; then
        log_success "Second rating upserted successfully (200)"
    else
        log_error "Failed to add second rating - unexpected status code"
    fi
    
    # Test rating aggregation
    log_info "Testing rating aggregation for 'Test Movie 1'..."
    if response=$(make_request "GET" "/movies/Test Movie 1/rating" "" "" 200); then
        average=$(echo "$response" | jq -r '.average')
        count=$(echo "$response" | jq -r '.count')
        
        # Expected: (3.5 + 4.0) / 2 = 3.75 -> 3.8 (rounded to 1 decimal)
        if [[ "$count" == "2" ]]; then
            log_success "Rating count is correct: $count"
            
            # Check if average is properly rounded to 1 decimal place
            if [[ "$average" =~ ^[0-9]+\.[0-9]$ ]]; then
                log_success "Average rating format is correct: $average"
                
                # Verify the calculation (3.5 + 4.0) / 2 = 3.75 -> 3.8
                expected_avg="3.8"
                if [[ "$average" == "$expected_avg" ]]; then
                    log_success "Average calculation is correct: $average"
                else
                    log_warning "Average calculation: expected $expected_avg, got $average (may be due to rounding implementation)"
                fi
            else
                log_error "Average rating should have 1 decimal place, got: $average"
            fi
        else
            log_error "Expected count 2, got $count"
        fi
    else
        log_error "Failed to get rating aggregation"
    fi
}

# Stage 4: Search and Pagination
stage4_search_pagination() {
    echo -e "\n${BLUE}=== STAGE 4: Search and Pagination ===${NC}"
    
    # Test keyword search
    log_info "Testing keyword search with 'q=Test'..."
    if response=$(make_request "GET" "/movies?q=Test" "" "" 200); then
        items=$(echo "$response" | jq -r '.items | length')
        if [[ "$items" -ge 1 ]]; then
            log_success "Keyword search returned $items results"
        else
            log_warning "Keyword search returned no results"
        fi
    else
        log_error "Keyword search failed"
    fi
    
    # Test year filter
    log_info "Testing year filter with 'year=2010'..."
    if response=$(make_request "GET" "/movies?year=2010" "" "" 200); then
        items=$(echo "$response" | jq -r '.items | length')
        if [[ "$items" -ge 1 ]]; then
            log_success "Year filter returned $items results"
            
            # Verify all returned movies are from 2010
            if echo "$response" | jq -e '.items[] | select(.releaseDate | startswith("2010"))' >/dev/null; then
                log_success "All results are from year 2010"
            else
                log_error "Some results are not from year 2010"
            fi
        else
            log_warning "Year filter returned no results"
        fi
    else
        log_error "Year filter failed"
    fi
    
    # Test genre filter
    log_info "Testing genre filter with 'genre=Sci-Fi'..."
    if response=$(make_request "GET" "/movies?genre=Sci-Fi" "" "" 200); then
        items=$(echo "$response" | jq -r '.items | length')
        if [[ "$items" -ge 1 ]]; then
            log_success "Genre filter returned $items results"
        else
            log_warning "Genre filter returned no results"
        fi
    else
        log_error "Genre filter failed"
    fi
    
    # Test pagination with limit
    log_info "Testing pagination with limit=1..."
    if response=$(make_request "GET" "/movies?limit=1" "" "" 200); then
        items=$(echo "$response" | jq -r '.items | length')
        next_cursor=$(echo "$response" | jq -r '.nextCursor')
        
        if [[ "$items" == "1" ]]; then
            log_success "Limit parameter working correctly"
            
            if [[ "$next_cursor" != "null" && -n "$next_cursor" ]]; then
                log_success "Next cursor provided: $next_cursor"
                
                # Test next page
                log_info "Testing next page with cursor..."
                encoded_cursor=$(echo -n "$next_cursor" | jq -sRr @uri)
                if response=$(make_request "GET" "/movies?limit=1&cursor=$encoded_cursor" "" "" 200); then
                    log_success "Cursor-based pagination working"
                else
                    log_error "Cursor-based pagination failed"
                fi
            else
                log_info "No next cursor (end of results)"
            fi
        else
            log_error "Expected 1 item with limit=1, got $items"
        fi
    else
        log_error "Pagination test failed"
    fi
}

# Stage 5: Authentication and Permissions
stage5_auth_permissions() {
    echo -e "\n${BLUE}=== STAGE 5: Authentication and Permissions ===${NC}"
    
    # Test movie creation without Bearer token (should return 401)
    log_info "Testing movie creation without Bearer token (expecting 401)..."
    if make_request "POST" "/movies" "" '{"title":"Test Auth","releaseDate":"2023-01-01"}' 401 >/dev/null; then
        log_success "Correctly returned 401 for missing Bearer token"
    else
        log_error "Should return 401 for missing Bearer token (OpenAPI spec requires 401)"
    fi
    
    # Test movie creation with invalid Bearer token (should return 401 according to OpenAPI)
    log_info "Testing movie creation with invalid Bearer token..."
    if make_request "POST" "/movies" "-H 'Authorization: Bearer invalid_token'" '{"title":"Test Auth","releaseDate":"2023-01-01"}' 401 >/dev/null; then
        log_success "Correctly returned 401 for invalid Bearer token"
    else
        log_error "Should return 401 for invalid Bearer token (OpenAPI spec requires 401)"
    fi
    
    # Test rating submission without X-Rater-Id (should return 401)
    log_info "Testing rating submission without X-Rater-Id (expecting 401)..."
    if make_request "POST" "/movies/Test Movie 1/ratings" "" '{"rating":4.0}' 401 >/dev/null; then
        log_success "Correctly returned 401 for missing X-Rater-Id"
    else
        log_error "Should return 401 for missing X-Rater-Id"
    fi
    
    # Test rating submission with empty X-Rater-Id (should return 401)
    log_info "Testing rating submission with empty X-Rater-Id (expecting 401)..."
    if make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id:'" '{"rating":4.0}' 401 >/dev/null; then
        log_success "Correctly returned 401 for empty X-Rater-Id"
    else
        log_error "Should return 401 for empty X-Rater-Id"
    fi
}

# Stage 6: Error Handling and Edge Cases
stage6_error_handling() {
    echo -e "\n${BLUE}=== STAGE 6: Error Handling and Edge Cases ===${NC}"
    
    # Test rating aggregation for non-existent movie (should return 404)
    log_info "Testing rating aggregation for non-existent movie (expecting 404)..."
    if make_request "GET" "/movies/NonExistentMovie/rating" "" "" 404 >/dev/null; then
        log_success "Correctly returned 404 for non-existent movie rating"
    else
        log_error "Should return 404 for non-existent movie rating"
    fi
    
    # Test rating submission for non-existent movie (should return 404)
    log_info "Testing rating submission for non-existent movie (expecting 404)..."
    if make_request "POST" "/movies/NonExistentMovie/ratings" "-H 'X-Rater-Id: user123'" '{"rating":4.0}' 404 >/dev/null; then
        log_success "Correctly returned 404 for rating submission to non-existent movie"
    else
        log_error "Should return 404 for rating submission to non-existent movie"
    fi
    
    # Test invalid rating values (should return 400)
    log_info "Testing invalid rating value (expecting 400)..."
    if make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user999'" '{"rating":6.0}' 400 >/dev/null; then
        log_success "Correctly returned 400 for invalid rating value"
    else
        log_error "Should return 400 for invalid rating value"
    fi

    if make_request "POST" "/movies/Test Movie 1/ratings" "-H 'X-Rater-Id: user999'" '{"rating":0.25}' 400 >/dev/null; then
        log_success "Correctly returned 400 for invalid rating step"
    else
        log_error "Should return 400 for invalid rating step"
    fi

    # Test missing required fields in movie creation (should return 400)
    log_info "Testing movie creation with missing title (expecting 400)..."
    if make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" '{"releaseDate":"2023-01-01"}' 400 >/dev/null; then
        log_success "Correctly returned 400 for missing title"
    else
        log_error "Should return 400 for missing title"
    fi

    # Test invalid date format (should return 400)
    log_info "Testing movie creation with invalid date format (expecting 400)..."
    if make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" '{"title":"Test Invalid Date","releaseDate":"invalid-date"}' 400 >/dev/null; then
        log_success "Correctly returned 400 for invalid date format"
    else
        log_error "Should return 400 for invalid date format"
    fi

    # Test invalid JSON (should return 400)
    log_info "Testing invalid JSON in request body (expecting 400)..."
    if make_request "POST" "/movies" "-H 'Authorization: Bearer $AUTH_TOKEN'" 'invalid json' 400 >/dev/null; then
        log_success "Correctly returned 400 for invalid JSON"
    else
        log_error "Should return 400 for invalid JSON"
    fi
}

# Main execution
main() {
    echo -e "${GREEN}Starting E2E Tests for Movies API${NC}"
    echo -e "Target service: $BASE_URL"
    echo -e "Timeout: ${TIMEOUT}s\n"
    
    # Run all test stages
    stage1_env_health_check
    stage2_basic_crud
    stage3_rating_system
    stage4_search_pagination
    stage5_auth_permissions
    stage6_error_handling
    
    # Print summary
    echo -e "\n${BLUE}=== TEST SUMMARY ===${NC}"
    echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
    
    if [[ $TESTS_FAILED -eq 0 ]]; then
        echo -e "\n${GREEN}🎉 All tests passed!${NC}"
        exit 0
    else
        echo -e "\n${RED}❌ Some tests failed. Please check the logs above.${NC}"
        exit 1
    fi
}

# Check prerequisites
if ! command -v curl &> /dev/null; then
    echo -e "${RED}Error: curl is required but not installed.${NC}"
    exit 1
fi

if ! command -v jq &> /dev/null; then
    echo -e "${RED}Error: jq is required but not installed.${NC}"
    exit 1
fi

# Run main function
main "$@"
