# Go Microservices Workspace

This workspace contains three Go microservices:
- **Products**: REST API for product catalog
- **Orders**: REST API for order management
- **Payment**: Listens to order events and updates payment status

## Features
- JWT authentication for all APIs
- PostgreSQL for Orders and Payments
- Google Pub/Sub for event-driven messaging
- Service discovery with Consul
- Docker Compose for local development

## Getting Started
1. Install Docker and Docker Compose
2. Run `docker-compose up --build`
3. Access services via their respective ports

## Structure
- `/products` - Product service
- `/orders` - Order service
- `/payment` - Payment service

## TODO
- Implement REST endpoints
- Add authentication and database integration
- Set up Pub/Sub and service discovery
