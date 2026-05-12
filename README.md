# Czanix Boilerplate — API Go

> Go idiomático. Sem framework pesado, sem ORM mágico, sem abstração que esconde o que importa. Struct, interface, e goroutine fazem o trabalho.

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go&logoColor=white)](https://go.dev)
[![Gin](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-316192?style=flat&logo=postgresql&logoColor=white)](https://postgresql.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tech Reference](https://img.shields.io/badge/Czanix-Tech%20Reference-gold)](https://czanix.com/pt/stack)

---

## Filosofia

Go brilha pela simplicidade. Este boilerplate respeita isso:

1. **Sem ORM** — SQL direto com `pgx`. Você sabe exatamente o que vai pro banco
2. **Sem DI container** — dependency injection manual via constructor. Go não precisa de mágica
3. **Interfaces implícitas** — o compilador garante o contrato sem annotations
4. **Errors are values** — Go já tem Result Pattern nativo: `(T, error)`

**O que não tem aqui:** GORM, framework opinado demais, reflection desnecessária, channel onde um mutex resolvia.

---

## Estrutura

```
├── cmd/
│   └── api/
│       └── main.go                  # Entrypoint — wire manual
│
├── internal/
│   ├── domain/
│   │   ├── order.go                 # Entidade pura
│   │   └── order_repository.go      # Interface (contrato)
│   │
│   ├── application/
│   │   ├── create_order.go          # Use case
│   │   └── cancel_order.go
│   │
│   ├── infrastructure/
│   │   ├── postgres/
│   │   │   ├── connection.go        # Pool pgx
│   │   │   └── order_repo.go        # Implementação do contrato
│   │   └── cache/
│   │       └── redis.go
│   │
│   └── presentation/
│       ├── handlers/
│       │   └── order_handler.go     # Gin handlers
│       ├── middleware/
│       │   ├── auth.go
│       │   ├── rate_limit.go
│       │   └── security_headers.go  # OWASP
│       └── router.go
│
├── migrations/                      # SQL versionado
├── docker-compose.yml
└── Makefile
```

### Por que `internal/`?

Go impede importação de pacotes `internal/` por outros módulos. É encapsulamento enforced pelo compilador, não por convenção.

---

## Início rápido

```bash
# 1. Clone
git clone https://github.com/czanix/boilerplate-api-go.git meu-projeto
cd meu-projeto

# 2. Dependências
go mod download

# 3. Ambiente
cp .env.example .env

# 4. Banco + Cache
docker compose up -d

# 5. Migrations
make migrate-up

# 6. Desenvolvimento
make dev
# ou: go run cmd/api/main.go
```

---

## Error handling — Go já tem Result Pattern

```go
// Em Go, (T, error) É o Result Pattern
func (uc *CreateOrderUseCase) Execute(input CreateOrderInput) (*OrderOutput, error) {
    if len(input.Items) == 0 {
        return nil, ErrOrderEmpty // erro de negócio, não panic
    }

    order := domain.NewOrder(input.CustomerID, input.Items)

    if err := uc.repo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("persisting order: %w", err) // wrap com contexto
    }

    return toOutput(order), nil
}

// Handler — tratamento explícito
func (h *OrderHandler) Create(c *gin.Context) {
    var input CreateOrderInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(400, gin.H{"error": "invalid input"})
        return
    }

    output, err := h.createOrder.Execute(input)
    if err != nil {
        // Erro de negócio vs erro de infra
        if errors.Is(err, domain.ErrOrderEmpty) {
            c.JSON(422, gin.H{"error": err.Error()})
            return
        }
        c.JSON(500, gin.H{"error": "internal error"})
        return
    }

    c.JSON(201, output)
}
```

**Go não tem exceção.** O chamador é sempre forçado a lidar com o erro. Isso é uma feature, não uma limitação.

---

## Connection Pool — pgx (não database/sql)

```go
// connection.go
func NewPool(cfg Config) (*pgxpool.Pool, error) {
    poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("parsing database URL: %w", err)
    }

    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5
    poolConfig.MaxConnLifetime = 30 * time.Minute
    poolConfig.MaxConnIdleTime = 5 * time.Minute

    pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
    if err != nil {
        return nil, fmt.Errorf("creating connection pool: %w", err)
    }

    return pool, nil
}
```

**Por que pgx e não database/sql?** Performance. pgx é nativo PostgreSQL, suporta COPY, notifications, tipos customizados. `database/sql` é genérico e adiciona overhead de abstração.

---

## Schema SQL

```sql
CREATE TABLE orders (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    public_id   UUID NOT NULL DEFAULT gen_random_uuid(),
    customer_id BIGINT NOT NULL REFERENCES customers(id),
    status      TEXT NOT NULL DEFAULT 'pending',
    deleted_at  TIMESTAMPTZ NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_orders_public_id UNIQUE (public_id)
);
```

**BIGINT PK + UUID público.** Padrão Czanix. [Mais detalhes →](https://czanix.com/pt/stack/dados)

---

## Segurança

```go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Next()
    }
}
```

---

## Makefile

```makefile
dev:              go run cmd/api/main.go
build:            go build -o bin/api cmd/api/main.go
test:             go test ./... -v -count=1
test-coverage:    go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
lint:             golangci-lint run
migrate-up:       migrate -path migrations -database $(DATABASE_URL) up
migrate-down:     migrate -path migrations -database $(DATABASE_URL) down 1
docker-build:     docker build -t czanix-api .
```

---

## Testes

```bash
make test                # Todos os testes
make test-coverage       # Com coverage report
go test ./internal/domain/... -v   # Só domínio
```

**Meta:** 80%+ em `domain/` e `application/`. Go testa rápido — use isso a seu favor.

---

## Referência técnica

- [Guia de Backend & Arquitetura](https://czanix.com/pt/stack/backend)
- [Guia de Database](https://czanix.com/pt/stack/dados)
- [Catálogo de Trade-offs](https://czanix.com/pt/stack/tradeoffs)
- [DevOps & CI/CD](https://czanix.com/pt/stack/devops)

---

## Licença

MIT — use, adapte, melhore. Se ajudou, [deixa uma estrela](https://github.com/czanix/boilerplate-api-go) ⭐

---

<div align="center">
<sub>Desenvolvido e mantido por <a href="https://czanix.com">Cesar Zanis</a> — Czanix</sub>
</div>
