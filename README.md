# Radgifa

## Introduction
Radgifa is a simple project aiming to solve a simple problem, creating and reviewing questionnaires, that's everything this does!
I just wanted something simple and self contained to ask questions to anyone and get answers and share look the results.

This isn't anything but a hobby project that I started as a joke but decided to finnish because a I really never finnish any of my projects. 

This is what I personally consider to be the bare minimum a project should have for basic maintenance and security:
## Technologies
### Backend
- Go 1.24
- Echo v4
- Ent ORM
- PostgreSQL (pgx driver)
- BadgerDB key-value storage
- JWT with `github.com/golang-jwt/jwt/v5`
- Validation with `go-playground/validator`
- Observability with `zap` logging

### Frontend
- Vue 3 (Composition API)
- Vite
- Vue Router
- Axios
- `qrcode` for QR generation

### Infrastructure and tools
- Docker and Docker Compose
- Swagger via `swag` and `echo-swagger`
- Testcontainers for integration tests
