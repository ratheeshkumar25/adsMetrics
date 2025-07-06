# Quick Start Guide - NATS Version

## Production-Ready Ads Metric Tracker with NATS

This guide will running with the NATS-based implementation of the Ads Metric Tracker.

### Why NATS instead of Kafka?

✅ **Lightweight**: Smaller footprint, faster startup  
✅ **Simpler**: No Zookeeper dependency  
✅ **Cloud-Native**: Better for microservices  
✅ **High Performance**: Lower latency messaging  
✅ **Built-in Monitoring**: HTTP monitoring interface  

## 🚀 One-Command Start

```bash
docker-compose -f docker-compose.nats.yaml up -d
```

That's it! Wait 30-60 seconds for all services to start.

## 🔍 Verify Everything Works

### 1. Check Service Health
```bash
# Check all services
docker-compose -f docker-compose.nats.yaml ps

# Test API health
curl http://localhost:8080/health

# Test NATS monitoring
curl http://localhost:8222/healthz
```

### 2. Run Comprehensive Tests
```bash
chmod +x test_api.sh
./test_api.sh
```

## 📊 Access Dashboards

| Service | URL | Credentials |
|---------|-----|-------------|
| **API** | http://localhost:8080 | - |
| **Grafana** | http://localhost:3000 | admin/admin123 |
| **Prometheus** | http://localhost:9090 | - |
| **NATS Monitor** | http://localhost:8222 | - |

## 🧪 Test API Endpoints

### Get All Ads
```bash
curl http://localhost:8080/api/v1/ads
```

### Record a Click
```bash
curl -X POST http://localhost:8080/api/v1/ads/click \
  -H "Content-Type: application/json" \
  -d '{
    "ad_id": "ad-001",
    "ip": "192.168.1.100",
    "video_play_time": 30
  }'
```

### View Analytics
```bash
curl "http://localhost:8080/api/v1/ads/analytics?ad_id=ad-001"
```

## ⚡ Load Testing

### Quick Load Test
```bash
# Send 100 clicks quickly
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/ads/click \
    -H "Content-Type: application/json" \
    -d "{\"ad_id\":\"ad-001\",\"ip\":\"192.168.1.$i\",\"video_play_time\":25}" \
    -o /dev/null -s &
done
wait
```

### View Results
```bash
# Check analytics after load test
curl "http://localhost:8080/api/v1/ads/analytics?ad_id=ad-001"

# View metrics
curl http://localhost:8080/metrics | grep -E "(ad_clicks|http_requests)"
```

## 📈 Monitoring & Metrics

### Key Metrics to Watch
- **HTTP Request Rate**: `rate(http_requests_total[5m])`
- **Click Processing Rate**: `rate(ad_clicks_total[5m])`
- **Error Rate**: `rate(http_requests_total{status=~"5.."}[5m])`
- **NATS Message Rate**: Check at http://localhost:8222

### Grafana Dashboard
1. Go to http://localhost:3000
2. Login: admin/admin123
3. Dashboard is auto-loaded: "Ads Metric Tracker - NATS Dashboard"

## 🔧 Advanced Usage

### Scale the Application
```bash
# Run 3 instances of the ads tracker
docker-compose -f docker-compose.nats.yaml up --scale ads-tracker=3 -d
```

### View Logs
```bash
# Application logs
docker-compose -f docker-compose.nats.yaml logs -f ads-tracker

# NATS logs
docker-compose -f docker-compose.nats.yaml logs -f nats

# All logs
docker-compose -f docker-compose.nats.yaml logs -f
```

### Monitor NATS
```bash
# Connection stats
curl http://localhost:8222/connz

# Subscription stats
curl http://localhost:8222/subsz

# General stats
curl http://localhost:8222/varz
```

## 🛠️ Development Mode

### Local Development
```bash
# Start only dependencies
docker-compose -f docker-compose.nats.yaml up postgres redis nats -d

# Run application locally
go run cmd/main.go
```

### Environment Variables for Local Development
```bash
export HTTP_HOST=0.0.0.0
export HTTP_PORT=8080
export POSTGRES_HOST=localhost
export POSTGRES_PORT=5432
export POSTGRES_USER=adsuser
export POSTGRES_PASSWORD=adspassword
export POSTGRES_DB=adsmetrics
export REDIS_HOST=localhost
export REDIS_PORT=6379
export NATS_URL=nats://localhost:4222
```

## 🚨 Troubleshooting

### NATS Not Connecting
```bash
# Check NATS status
docker-compose -f docker-compose.nats.yaml logs nats

# Test NATS connectivity
curl http://localhost:8222/healthz
```

### Application Won't Start
```bash
# Check all service dependencies
docker-compose -f docker-compose.nats.yaml ps

# View application logs
docker-compose -f docker-compose.nats.yaml logs ads-tracker
```

### Performance Issues
```bash
# Check resource usage
docker stats

# View detailed metrics
curl http://localhost:8080/metrics
```

## 🎯 Production Deployment

### Resource Requirements
- **CPU**: 2 cores minimum
- **Memory**: 4GB minimum  
- **Storage**: 10GB minimum
- **Network**: 1Gbps recommended

### Environment Checklist
- [ ] PostgreSQL configured with proper connection pooling
- [ ] Redis configured with appropriate memory limits
- [ ] NATS JetStream enabled for persistence
- [ ] Prometheus scraping configured
- [ ] Grafana dashboards imported
- [ ] Log rotation configured
- [ ] Health checks configured in load balancer

## 📚 What's Different from Kafka Version?

| Feature | Kafka Version | NATS Version |
|---------|---------------|--------------|
| **Startup Time** | ~60 seconds | ~20 seconds |
| **Memory Usage** | ~1GB | ~200MB |
| **Dependencies** | Kafka + Zookeeper | NATS only |
| **Configuration** | Complex | Simple |
| **Monitoring** | JMX + external tools | Built-in HTTP |
| **Clustering** | Complex setup | Simple clustering |

## 🎉 Success!

If you can see metrics in Grafana and the API responds correctly to the test script, you have successfully deployed a production-ready ads metric tracking system with:

✅ High-performance Go backend  
✅ NATS message processing  
✅ PostgreSQL data persistence  
✅ Redis caching  
✅ Prometheus metrics  
✅ Grafana dashboards  
✅ Circuit breakers & fault tolerance  
✅ Comprehensive testing  

---

**Next Steps**: Check out the main README.md for detailed API documentation and advanced configuration options.
