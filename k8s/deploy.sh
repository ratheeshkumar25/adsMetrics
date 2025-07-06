#!/bin/bash

echo "🚀 Deploying Ads Metric Tracker to Kubernetes (Simplified)"
echo "========================================================="

# Deploy in order (dependencies first)
echo "1. Deploying PostgreSQL..."
kubectl apply -f postgres.yaml

echo "2. Deploying Redis..."
kubectl apply -f redis.yaml

echo "3. Deploying NATS..."
kubectl apply -f nats.yaml

echo "4. Waiting for dependencies to be ready..."
sleep 10

echo "5. Deploying Application..."
kubectl apply -f deployment.yaml

echo "6. Checking deployment status..."
kubectl get pods -l app=ads-tracker
kubectl get services

echo ""
echo "✅ Deployment completed!"
echo ""
echo "📊 Access the application:"
echo "   kubectl port-forward service/ads-tracker-service 8080:80"
echo "   Then access: http://localhost:8080"
echo ""
echo "🔍 Check logs:"
echo "   kubectl logs -f deployment/ads-tracker"
echo ""
echo "📈 Monitor pods:"
echo "   kubectl get pods -w"
