#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Banner
cat << 'EOF'
 █████╗ ██████╗ ███████╗    ███╗   ███╗███████╗████████╗██████╗ ██╗ ██████╗
██╔══██╗██╔══██╗██╔════╝    ████╗ ████║██╔════╝╚══██╔══╝██╔══██╗██║██╔════╝
███████║██║  ██║███████╗    ██╔████╔██║█████╗     ██║   ██████╔╝██║██║     
██╔══██║██║  ██║╚════██║    ██║╚██╔╝██║██╔══╝     ██║   ██╔══██╗██║██║     
██║  ██║██████╔╝███████║    ██║ ╚═╝ ██║███████╗   ██║   ██║  ██║██║╚██████╗
╚═╝  ╚═╝╚═════╝ ╚══════╝    ╚═╝     ╚═╝╚══════╝   ╚═╝   ╚═╝  ╚═╝╚═╝ ╚═════╝
                                                                              
████████╗██████╗  █████╗  ██████╗██╗  ██╗███████╗██████╗                    
╚══██╔══╝██╔══██╗██╔══██╗██╔════╝██║ ██╔╝██╔════╝██╔══██╗                   
   ██║   ██████╔╝███████║██║     █████╔╝ █████╗  ██████╔╝                   
   ██║   ██╔══██╗██╔══██║██║     ██╔═██╗ ██╔══╝  ██╔══██╗                   
   ██║   ██║  ██║██║  ██║╚██████╗██║  ██╗███████╗██║  ██║                   
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝                   
EOF

echo -e "${BLUE}🚀 High-Performance Video Ad Click Tracking System${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Function to check if Docker is running
check_docker() {
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed or not in PATH${NC}"
        exit 1
    fi
    
    if ! docker info &> /dev/null; then
        echo -e "${RED}❌ Docker daemon is not running${NC}"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed or not in PATH${NC}"
        exit 1
    fi
}

# Function to show current status
show_status() {
    echo -e "\n${PURPLE}📊 Current System Status:${NC}"
    
    # Check if containers are running
    nats_running=$(docker ps --filter "name=ads-tracker" --format "table {{.Names}}" | grep -c "ads-tracker" || echo "0")
    postgres_running=$(docker ps --filter "name=ads-postgres" --format "table {{.Names}}" | grep -c "ads-postgres" || echo "0")
    redis_running=$(docker ps --filter "name=ads-redis" --format "table {{.Names}}" | grep -c "ads-redis" || echo "0")
    
    if [ "$nats_running" -gt 0 ] && [ "$postgres_running" -gt 0 ] && [ "$redis_running" -gt 0 ]; then
        echo -e "   🟢 Application: ${GREEN}RUNNING${NC}"
        echo -e "   🌐 API: http://localhost:8080"
        echo -e "   ❤️ Health: http://localhost:8080/health"
        echo -e "   📈 Metrics: http://localhost:8080/metrics"
        echo -e "   📊 Prometheus: http://localhost:9090"
        echo -e "   📉 Grafana: http://localhost:3000 (admin/admin123)"
        
        # Test API health
        if curl -s http://localhost:8080/health &> /dev/null; then
            echo -e "   ✅ API Health: ${GREEN}HEALTHY${NC}"
        else
            echo -e "   ⚠️  API Health: ${YELLOW}STARTING${NC}"
        fi
    else
        echo -e "   🔴 Application: ${RED}STOPPED${NC}"
    fi
}

# Function to start NATS setup
start_nats() {
    echo -e "\n${BLUE}🚀 Starting Ads Metric Tracker with NATS...${NC}"
    echo -e "${YELLOW}   • Lightweight messaging with JetStream${NC}"
    echo -e "${YELLOW}   • Recommended for development and production${NC}"
    
    docker-compose -f docker-compose.nats.yaml up -d
    
    if [ $? -eq 0 ]; then
        echo -e "\n${GREEN}✅ All services started successfully!${NC}"
        echo -e "\n${PURPLE}📊 Access Points:${NC}"
        echo -e "   🌐 API Endpoint: ${BLUE}http://localhost:8080${NC}"
        echo -e "   ❤️  Health Check: ${BLUE}http://localhost:8080/health${NC}"
        echo -e "   📈 Prometheus: ${BLUE}http://localhost:9090${NC}"
        echo -e "   📉 Grafana: ${BLUE}http://localhost:3000${NC} (admin/admin123)"
        echo -e "   🔗 NATS Monitor: ${BLUE}http://localhost:8222${NC}"
        
        echo -e "\n${YELLOW}⏳ Waiting for services to be ready...${NC}"
        sleep 5
        
        # Test health
        for i in {1..30}; do
            if curl -s http://localhost:8080/health &> /dev/null; then
                echo -e "${GREEN}✅ Application is ready!${NC}"
                break
            fi
            echo -n "."
            sleep 1
        done
        
        echo -e "\n${PURPLE}🧪 Quick API Test:${NC}"
        echo -e "   curl http://localhost:8080/health"
        echo -e "   curl http://localhost:8080/ads"
    else
        echo -e "${RED}❌ Failed to start services${NC}"
    fi
}

# Function to start Kafka setup
start_kafka() {
    echo -e "\n${BLUE}🚀 Starting Ads Metric Tracker with Kafka...${NC}"
    echo -e "${YELLOW}   • Enterprise messaging with Kafka + Zookeeper${NC}"
    echo -e "${YELLOW}   • Production-ready setup${NC}"
    
    docker-compose -f docker-compose.prod.yaml up -d
    
    if [ $? -eq 0 ]; then
        echo -e "\n${GREEN}✅ All services started successfully!${NC}"
        echo -e "\n${PURPLE}📊 Access Points:${NC}"
        echo -e "   🌐 API Endpoint: ${BLUE}http://localhost:8080${NC}"
        echo -e "   ❤️  Health Check: ${BLUE}http://localhost:8080/health${NC}"
        echo -e "   📈 Prometheus: ${BLUE}http://localhost:9090${NC}"
        echo -e "   📉 Grafana: ${BLUE}http://localhost:3000${NC} (admin/admin123)"
        
        echo -e "\n${YELLOW}⏳ Waiting for services to be ready (Kafka takes longer)...${NC}"
        sleep 10
        
        # Test health
        for i in {1..60}; do
            if curl -s http://localhost:8080/health &> /dev/null; then
                echo -e "${GREEN}✅ Application is ready!${NC}"
                break
            fi
            echo -n "."
            sleep 1
        done
    else
        echo -e "${RED}❌ Failed to start services${NC}"
    fi
}

# Function to stop services
stop_services() {
    echo -e "\n${BLUE}🛑 Stopping all services...${NC}"
    docker-compose -f docker-compose.nats.yaml down 2>/dev/null || true
    docker-compose -f docker-compose.prod.yaml down 2>/dev/null || true
    echo -e "${GREEN}✅ All services stopped${NC}"
}

# Function to run tests
run_tests() {
    echo -e "\n${BLUE}🧪 Running comprehensive tests...${NC}"
    if [ -f "test_requirements.sh" ]; then
        chmod +x test_requirements.sh
        ./test_requirements.sh
    else
        echo -e "${RED}❌ Test script not found${NC}"
    fi
}

# Function to show logs
show_logs() {
    echo -e "\n${BLUE}📋 Showing application logs...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to exit${NC}"
    docker logs -f ads-tracker 2>/dev/null || echo "Application not running"
}

# Main menu
main_menu() {
    while true; do
        show_status
        
        echo -e "\n${PURPLE}🎯 Choose an option:${NC}"
        echo -e "   ${GREEN}1)${NC} Start with NATS (Recommended)"
        echo -e "   ${GREEN}2)${NC} Start with Kafka (Production)"
        echo -e "   ${GREEN}3)${NC} Stop all services"
        echo -e "   ${GREEN}4)${NC} Show logs"
        echo -e "   ${GREEN}5)${NC} Run tests"
        echo -e "   ${GREEN}6)${NC} Rebuild and restart"
        echo -e "   ${GREEN}7)${NC} Clean and restart"
        echo -e "   ${GREEN}q)${NC} Quit"
        
        echo -e "\n${YELLOW}Enter your choice [1-7/q]:${NC} "
        read -r choice
        
        case $choice in
            1)
                start_nats
                ;;
            2)
                start_kafka
                ;;
            3)
                stop_services
                ;;
            4)
                show_logs
                ;;
            5)
                run_tests
                ;;
            6)
                echo -e "\n${BLUE}🔄 Rebuilding and restarting...${NC}"
                docker-compose -f docker-compose.nats.yaml up -d --build
                ;;
            7)
                echo -e "\n${BLUE}🧹 Cleaning and restarting...${NC}"
                stop_services
                docker system prune -f
                start_nats
                ;;
            q|Q)
                echo -e "\n${GREEN}👋 Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}❌ Invalid option. Please try again.${NC}"
                ;;
        esac
        
        echo -e "\n${YELLOW}Press Enter to continue...${NC}"
        read -r
    done
}

# Check requirements
check_docker

# Handle command line arguments
if [ $# -eq 0 ]; then
    main_menu
else
    case $1 in
        "nats"|"dev")
            start_nats
            ;;
        "kafka"|"prod")
            start_kafka
            ;;
        "stop"|"down")
            stop_services
            ;;
        "test")
            run_tests
            ;;
        "logs")
            show_logs
            ;;
        "status")
            show_status
            ;;
        *)
            echo -e "${RED}❌ Unknown command: $1${NC}"
            echo -e "Usage: $0 [nats|kafka|stop|test|logs|status]"
            exit 1
            ;;
    esac
fi
