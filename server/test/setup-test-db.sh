#!/bin/bash

# Setup script for BDD test database

set -e

echo "🚀 Setting up BDD test database..."

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if container already exists
if docker ps -a | grep -q postgres-test; then
    echo "📦 postgres-test container already exists"
    
    # Check if it's running
    if docker ps | grep -q postgres-test; then
        echo "✅ postgres-test is already running"
    else
        echo "▶️  Starting postgres-test container..."
        docker start postgres-test
        echo "✅ postgres-test started"
    fi
else
    echo "📦 Creating postgres-test container..."
    docker run -d \
        --name postgres-test \
        -e POSTGRES_USER=vending \
        -e POSTGRES_PASSWORD=vending \
        -e POSTGRES_DB=vending_test \
        -p 5432:5432 \
        postgres:15
    
    echo "⏳ Waiting for PostgreSQL to be ready..."
    sleep 3
    
    echo "✅ postgres-test container created and started"
fi

# Test connection
echo "🔍 Testing database connection..."
if docker exec postgres-test pg_isready -U vending > /dev/null 2>&1; then
    echo "✅ Database is ready!"
else
    echo "⏳ Waiting for database to be ready..."
    sleep 2
    if docker exec postgres-test pg_isready -U vending > /dev/null 2>&1; then
        echo "✅ Database is ready!"
    else
        echo "❌ Database is not responding. Please check Docker logs:"
        echo "   docker logs postgres-test"
        exit 1
    fi
fi

echo ""
echo "🎉 Test database setup complete!"
echo ""
echo "Connection details:"
echo "  Host:     localhost"
echo "  Port:     5432"
echo "  Database: vending_test"
echo "  User:     vending"
echo "  Password: vending"
echo ""
echo "Connection string:"
echo "  postgres://vending:vending@localhost:5432/vending_test?sslmode=disable"
echo ""
echo "Next steps:"
echo "  1. Run tests: make test-bdd"
echo "  2. Or: cd server && go test -v ./test/..."
echo ""
echo "To stop the database:"
echo "  docker stop postgres-test"
echo ""
echo "To remove the database:"
echo "  docker rm -f postgres-test"
