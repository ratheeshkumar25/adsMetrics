#!/bin/bash

# Ads Metric Tracker API Test Script with NATS
# Production-ready testing for all endpoints

set -e

API_BASE="http://localhost:8080"
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"
NATS_URL="http://localhost:8222"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=================================================${NC}"
echo -e "${BLUE}  Ads Metric Tracker API Testing Suite (NATS)   ${NC}"
echo -e "${BLUE}=================================================${NC}"

# Function to check service health
check_service() {
    local service_name=$1
    local url=$2
    local expected_code=${3:-200}
    
    echo -e "\n${YELLOW}Checking $service_name health...${NC}"
    if curl -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_code"; then
        echo -e "${GREEN}✓ $service_name is healthy${NC}"
        return 0
    else
        echo -e "${RED}✗ $service_name is not responding${NC}"
        return 1
    fi
}

# Function to test API endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local expected_code=${4:-200}
    local description=$5
    
    echo -e "\n${YELLOW}Testing: $description${NC}"
    echo -e "Endpoint: $method $endpoint"
    
    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" "$API_BASE$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" -H "Content-Type: application/json" -d "$data" "$API_BASE$endpoint")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n -1)
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ Success ($http_code)${NC}"
        echo -e "Response: $body" | head -c 200
        echo -e "\n"
        return 0
    else
        echo -e "${RED}✗ Failed ($http_code, expected $expected_code)${NC}"
        echo -e "Response: $body"
        return 1
    fi
}

# Function to test load
test_load() {
    local endpoint=$1
    local data=$2
    local count=${3:-100}
    
    echo -e "\n${YELLOW}Load testing $endpoint with $count requests...${NC}"
    
    start_time=$(date +%s)
    success_count=0
    
    for i in $(seq 1 $count); do
        if curl -s -o /dev/null -w "%{http_code}" -X POST -H "Content-Type: application/json" -d "$data" "$API_BASE$endpoint" | grep -q "202"; then
            ((success_count++))
        fi
        
        # Show progress every 20 requests
        if [ $((i % 20)) -eq 0 ]; then
            echo -e "${BLUE}Progress: $i/$count${NC}"
        fi
    done
    
    end_time=$(date +%s)
    duration=$((end_time - start_time))
    rps=$((count / duration))
    
    echo -e "${GREEN}✓ Load test completed${NC}"
    echo -e "  Successful requests: $success_count/$count"
    echo -e "  Duration: ${duration}s"
    echo -e "  RPS: $rps"
}

# Main testing flow
main() {
    echo -e "\n${BLUE}Step 1: Health Checks${NC}"
    
    # Check all services
    check_service "Ads Tracker API" "$API_BASE/health"
    check_service "Prometheus" "$PROMETHEUS_URL/-/healthy"
    check_service "Grafana" "$GRAFANA_URL/api/health"
    check_service "NATS" "$NATS_URL/healthz"
    
    echo -e "\n${BLUE}Step 2: API Functionality Tests${NC}"
    
    # Test GET /ads
    test_endpoint "GET" "/ads" "" "200" "Get all ads"
    
    # Test POST /ads/click with various scenarios
    test_endpoint "POST" "/ads/click" \
        '{"ad_id":"tech-001","ip":"192.168.1.100","video_play_time":30}' \
        "202" "Record click for tech-001"
    
    test_endpoint "POST" "/ads/click" \
        '{"ad_id":"tech-002","ip":"192.168.1.101","video_play_time":45}' \
        "202" "Record click for tech-002"
    
    test_endpoint "POST" "/ads/click" \
        '{"ad_id":"fashion-001","ip":"192.168.1.102","video_play_time":15}' \
        "202" "Record click for fashion-001"
    
    # Test invalid ad ID
    test_endpoint "POST" "/ads/click" \
        '{"ad_id":"invalid-ad","ip":"192.168.1.103","video_play_time":30}' \
        "400" "Record click for invalid ad (should fail)"
    
    # Test GET /ads/analytics
    sleep 2 # Give time for async processing
    
    test_endpoint "GET" "/ads/analytics" "" "200" "Get analytics for all ads"
    test_endpoint "GET" "/ads/analytics?ad_id=tech-001" "" "200" "Get analytics for tech-001"
    test_endpoint "GET" "/ads/analytics?ad_id=tech-002&timeframe=5m" "" "200" "Get analytics for tech-002 (5m timeframe)"
    
    echo -e "\n${BLUE}Step 3: Load Testing${NC}"
    
    # Perform load test
    test_load "/ads/click" \
        '{"ad_id":"tech-001","ip":"192.168.1.200","video_play_time":25}' \
        50
    
    echo -e "\n${BLUE}Step 4: Monitoring Tests${NC}"
    
    # Check metrics endpoint
    test_endpoint "GET" "/metrics" "" "200" "Prometheus metrics endpoint"
    
    # Wait for metrics to be processed
    sleep 5
    
    # Test analytics after load
    test_endpoint "GET" "/ads/analytics?ad_id=tech-001" "" "200" "Analytics after load test"
    
    echo -e "\n${BLUE}Step 5: Service URLs${NC}"
    echo -e "${GREEN}✓ API Documentation: $API_BASE/swagger/index.html${NC}"
    echo -e "${GREEN}✓ Prometheus UI: $PROMETHEUS_URL${NC}"
    echo -e "${GREEN}✓ Grafana Dashboard: $GRAFANA_URL (admin/admin123)${NC}"
    echo -e "${GREEN}✓ NATS Monitoring: $NATS_URL${NC}"
    
    echo -e "\n${BLUE}Step 6: Production Readiness Checks${NC}"
    
    # Check for key metrics
    echo -e "${YELLOW}Checking key metrics availability...${NC}"
    
    if curl -s "$API_BASE/metrics" | grep -q "http_requests_total"; then
        echo -e "${GREEN}✓ HTTP request metrics available${NC}"
    else
        echo -e "${RED}✗ HTTP request metrics missing${NC}"
    fi
    
    if curl -s "$API_BASE/metrics" | grep -q "ad_clicks_total"; then
        echo -e "${GREEN}✓ Ad click metrics available${NC}"
    else
        echo -e "${RED}✗ Ad click metrics missing${NC}"
    fi
    
    echo -e "\n${GREEN}=================================================${NC}"
    echo -e "${GREEN}           Testing Complete!                     ${NC}"
    echo -e "${GREEN}=================================================${NC}"
    
    echo -e "\n${BLUE}Next Steps:${NC}"
    echo -e "1. View real-time metrics: $GRAFANA_URL"
    echo -e "2. Monitor NATS: $NATS_URL"
    echo -e "3. Check application logs: docker-compose logs ads-tracker"
    echo -e "4. Scale services: docker-compose up --scale ads-tracker=3"
}

# Run tests
main "$@"
