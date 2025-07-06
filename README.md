# Ads Metric Tracker

A high-performance GoLang backend service for managing and tracking video advertisement clicks with real-time analytics, fault tolerance, and scalability.

## Features

### Core Requirements ✅
- **GET /ads** - Returns a list of ads with basic metadata
- **POST /ads/click** - Accepts click details with asynchronous processing
- **GET /ads/analytics** - Returns real-time performance metrics
- **Data Integrity** - No data loss with circuit breakers and retry mechanisms
- **Scalability** - Handles concurrent requests and traffic spikes
- **Real-time Analytics** - Efficient aggregated click data retrieval

### Production Ready Features ✅
- **Docker** - Complete containerization with multi-stage builds
- **Configuration** - Environment-based configuration management
- **Logging** - Structured logging with rotation
- **Monitoring** - Prometheus metrics and Grafana dashboards
- **Health Checks** - Application and dependency health monitoring

### Advanced Features ✅
- **PostgreSQL** - Primary database with connection pooling
- **Redis** - Caching layer for improved performance
- **NATS** - Lightweight asynchronous message processing
- **Circuit Breakers** - Fault tolerance and graceful degradation
- **Batch Processing** - Efficient bulk operations
- **Rate Limiting** - Protection against traffic spikes

## Architecture

```
┌─────────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Load Balancer │────│  API Gateway │────│   Application   │
└─────────────────┘    └──────────────┘    └─────────────────┘
                                                    │
                              ┌─────────────────────┼─────────────────────┐
                              │                     │                     │
                        ┌──────────┐          ┌──────────┐          ┌──────────┐
                        │   Redis  │          │   NATS   │          │PostgreSQL│
                        │  (Cache) │          │(Messages)│          │(Primary) │
                        └──────────┘          └──────────┘          └──────────┘
                              │                     │                     │
                        ┌──────────┐          ┌──────────┐          ┌──────────┐
                        │Prometheus│          │ Grafana  │          │  Logs    │
                        │(Metrics) │          │(Monitor) │          │(Storage) │
                        └──────────┘          └──────────┘          └──────────┘
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Make (optional)

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd adsMetricTracker
   ```

2. **Start all services**
   ```bash
   docker-compose -f docker-compose.prod.yaml up -d
   ```

3. **Verify services are running**
   ```bash
   docker-compose -f docker-compose.prod.yaml ps
   ```

4. **Check application health**
   ```bash
   curl http://localhost:8080/health
   ```

### Services and Ports
- **Ads Tracker API**: http://localhost:8080
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin123)
- **PostgreSQL**: localhost:5432
- **Redis**: localhost:6379
- **NATS**: localhost:4222 (Client), localhost:8222 (HTTP Monitoring)

## API Documentation

### Interactive API Documentation (Swagger)

The API includes comprehensive interactive documentation powered by Swagger/OpenAPI:

- **Swagger UI**: http://localhost:8080/swagger/index.html
- **OpenAPI JSON**: http://localhost:8080/swagger/doc.json
- **OpenAPI YAML**: http://localhost:8080/swagger/swagger.yaml

#### Swagger Features:
- 🔍 **Interactive Testing**: Test all endpoints directly from the browser
- 📚 **Complete Documentation**: Detailed request/response schemas
- 🎯 **Real-time Validation**: Input validation with examples
- 📋 **Copy-paste Ready**: Generate curl commands automatically
- 🔄 **Try It Out**: Execute API calls with sample data

#### Quick Access:
```bash
# Open Swagger UI in browser
open http://localhost:8080/swagger/index.html

# Or get API info
curl http://localhost:8080/info
```

### Endpoints

#### GET /api/v1/ads
Returns a list of ads with basic metadata.

**Response:**
```json
{
  "ads": [
    {
      "id": "ad-001",
      "image_url": "https://example.com/ad1.jpg",
      "target_url": "https://example.com/product1",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 1
}
```

#### POST /api/v1/ads/click
Records a click event (asynchronous processing).

**Request:**
```json
{
  "ad_id": "ad-001",
  "ip": "192.168.1.1",
  "video_play_time": 30,
  "timestamp": "2024-01-01T12:00:00Z"
}
```clr

**Response (202 Accepted):**
```json
{
  "message": "Click recorded",
  "click_id": "click-uuid",
  "ad_id": "ad-001",
  "timestamp": "2024-01-01T12:00:00Z",
  "processing": "asynchronous"
}
```

#### GET /api/v1/ads/analytics
Returns real-time analytics.

**Query Parameters:**
- `ad_id` (optional): Specific ad ID
- `timeframe` (optional): Time range (1m, 5m, 15m, 1h, 24h)

**Response:**
```json
{
  "ad_id": "ad-001",
  "total_clicks": 1500,
  "ctr": 0.15,
  "time_frames": {
    "last_1_minute": 5,
    "last_5_minutes": 25,
    "last_15_minutes": 75,
    "last_1_hour": 300,
    "last_24_hours": 1500
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Health and Monitoring

#### GET /health
Application health check.

#### GET /metrics
Prometheus metrics endpoint.

## Development

### Local Development Setup

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Set up environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start dependencies**
   ```bash
   docker-compose -f docker-compose.nats.yaml up postgres redis nats -d
   ```

4. **Run the application**
   ```bash
   go run cmd/main.go
   ```

### Testing


### Load Testing

```bash
# Install k6 (if not already installed)
# Test click endpoint
k6 run scripts/load-test.js
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_HOST` | `0.0.0.0` | Server bind address |
| `HTTP_PORT` | `8080` | Server port |
| `LOG_FILE` | `logs/app.log` | Log file path |
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_USER` | `adsuser` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `adspassword` | PostgreSQL password |
| `POSTGRES_DB` | `adsmetrics` | PostgreSQL database |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password |
| `REDIS_DB` | `0` | Redis database |
| `KAFKA_BROKER` | `localhost:9092` | Kafka broker |

## Performance & Scalability

### Concurrency Features
- **Asynchronous Processing**: Click events processed via NATS
- **Batch Processing**: Efficient bulk database operations
- **Connection Pooling**: Optimized database connections
- **Caching**: Redis for frequently accessed data
- **Circuit Breakers**: Fault tolerance for external dependencies

### Scalability Strategies
- **Horizontal Scaling**: Stateless application design
- **Database Optimization**: Indexed queries and connection pooling
- **Cache Strategy**: Redis for hot data
- **Message Queuing**: NATS for decoupled processing
- **Load Balancing**: Ready for multiple instances

## Monitoring & Observability

### Metrics (Prometheus)
- **HTTP Requests**: Request count, duration, status codes
- **Database Operations**: Query performance and errors
- **Redis Operations**: Cache hit/miss rates
- **Kafka Operations**: Message throughput and errors
- **Application Metrics**: Click processing rates

### Dashboards (Grafana)
- **Application Overview**: Key performance indicators
- **System Health**: Resource utilization
- **Business Metrics**: Click rates and analytics
- **Error Tracking**: Error rates and types

### Logs
- **Structured Logging**: JSON format with contextual information
- **Log Rotation**: Automated log file management
- **Log Levels**: Debug, Info, Warn, Error

## Deployment

### Production Deployment

1. **Build Docker image**
   ```bash
   docker build -t ads-metric-tracker:latest .
   ```

2. **Deploy with Docker Compose**
   ```bash
   docker-compose -f docker-compose.prod.yaml up -d
   ```

3. **Kubernetes (Optional)**
   ```bash
   kubectl apply -f k8s/
   ```

### Security Considerations
- **Non-root containers**: Application runs as non-privileged user
- **Network isolation**: Services isolated in Docker networks
- **Environment secrets**: Sensitive data via environment variables
- **Health checks**: Kubernetes-ready health endpoints

## Troubleshooting

### Common Issues

1. **Database connection failed**
   - Verify PostgreSQL is running
   - Check connection parameters
   - Ensure database exists

2. **NATS connection failed**
   - Verify NATS server is running
   - Check server configuration
   - Application continues without NATS (direct processing)

3. **Redis connection failed**
   - Verify Redis is running
   - Application continues without Redis (database-only mode)

### Logs
```bash
# Application logs
docker-compose logs ads-tracker

# Database logs
docker-compose logs postgres

# NATS logs
docker-compose logs nats
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details.

## Demonstration

### Sample Data
The application automatically seeds sample ads on startup:
- **ad-001**: Sample product advertisement
- **ad-002**: Sample service advertisement  
- **ad-003**: Sample brand advertisement

### Testing Click Recording
```bash
# Record a click
curl -X POST http://localhost:8080/api/v1/ads/click \
  -H "Content-Type: application/json" \
  -d '{
    "ad_id": "ad-001",
    "video_play_time": 30
  }'

# Get analytics
curl "http://localhost:8080/api/v1/ads/analytics?ad_id=ad-001"
```

### Monitoring Dashboard
1. Visit Grafana: http://localhost:3000
2. Login: admin/admin123
3. Import dashboard from `grafana/dashboards/`
4. View real-time metrics and analytics

---

**Built with ❤️ using Go, PostgreSQL, Redis, NATS, and modern DevOps practices.**
