# E-Commerce Dashboard Example

Full-stack e-commerce dashboard with Nucleus backend API and React frontend.

## Architecture

This example follows the MVC pattern with clean separation:

```
backend/
├── main.go           # Application entry point (48 lines)
├── models/           # Domain models (Product, Order, Customer, etc.)
├── handlers/         # HTTP handlers (API endpoints)
└── seed/             # Database seeding

frontend/
├── React + TypeScript + Vite
├── TailwindCSS for styling
└── React Router for navigation
```

## Backend Structure

The backend demonstrates the simplified Nucleus API:

- **Models**: Define schema with struct tags
- **Handlers**: HTTP handlers using `*nucleus.Context`
- **Fluent API**: Chain configuration with `nucleus.New()`

## Run

From repository root:

```bash
# Terminal 1: Backend
cd examples/ecommerce_dashboard/backend
go run .

# Terminal 2: Frontend (development)
cd examples/ecommerce_dashboard/frontend
npm install
npm run dev
```

## Production Build

```bash
cd examples/ecommerce_dashboard/frontend
npm run build

cd ../backend
go build -o server
./server
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /api/stats | Dashboard statistics |
| GET | /api/products | List products |
| POST | /api/products | Create product |
| GET | /api/products/:id | Get product |
| GET | /api/orders | List orders |
| POST | /api/orders | Create order |
| GET | /api/customers | List customers |
| GET | /api/customers/:id | Get customer |
| GET | /api/categories | List categories |

## Access

- **App**: http://localhost:8080
- **API**: http://localhost:8080/api

## Purpose

Use this example as a reference for:

- Building APIs with the simplified `pkg/nucleus` fluent API
- Organizing code with handlers/models separation
- Serving SPAs with `app.SPA()`
- Full-stack Nucleus applications
