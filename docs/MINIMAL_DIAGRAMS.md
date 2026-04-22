# Electroteca — Diagramas Simplificados

> Versión compacta para lectura rápida. Diagramas detallados en `DIAGRAMS.md`.

---

## 1. Entidad-Relación

```mermaid
erDiagram
    users {
        UUID id PK
        VARCHAR username
        VARCHAR email
        VARCHAR role
        BOOLEAN active
    }

    categories {
        UUID id PK
        VARCHAR name
    }

    products {
        UUID id PK
        VARCHAR name
        VARCHAR sku
        UUID category_id FK
        UUID created_by_id FK
        VARCHAR status
        BOOLEAN paid
        NUMERIC price
    }

    contacts {
        UUID id PK
        UUID product_id FK
        VARCHAR name
        VARCHAR email
        VARCHAR phone
    }

    product_images {
        UUID id PK
        UUID product_id FK
        VARCHAR image_key
        INT position
    }

    refresh_tokens {
        UUID id PK
        UUID user_id FK
        VARCHAR token_hash
        TIMESTAMPTZ expires_at
        BOOLEAN revoked
    }

    users ||--o{ products : "crea"
    users ||--o{ refresh_tokens : "posee"
    categories ||--o{ products : "clasifica"
    products ||--o| contacts : "tiene"
    products ||--o{ product_images : "galeria"
```

---

## 2. Diagrama de Clases

```mermaid
classDiagram
    class User {
        +UUID id
        +string username
        +string email
        +Role role
        +bool active
    }

    class Product {
        +UUID id
        +string name
        +string sku
        +float64 price
        +ProductStatus status
        +bool paid
        +UUID categoryId
    }

    class Category {
        +UUID id
        +string name
        +string description
    }

    class Contact {
        +UUID id
        +UUID productId
        +string name
        +string email
        +string phone
    }

    class ProductImage {
        +UUID id
        +UUID productId
        +string imageKey
        +int position
    }

    class RefreshToken {
        +UUID id
        +UUID userId
        +string tokenHash
        +bool revoked
    }

    class Role {
        <<enumeration>>
        admin
        manager
        viewer
    }

    class ProductStatus {
        <<enumeration>>
        reparado
        en_progreso
        no_reparado
    }

    User --> Role
    Product --> ProductStatus
    User "1" --> "0..*" Product : crea
    User "1" --> "0..*" RefreshToken : posee
    Product "0..*" --> "0..1" Category : pertenece
    Product "1" --> "0..1" Contact : tiene
    Product "1" --> "0..*" ProductImage : galeria
```

---

## 3. Arquitectura — Capas del sistema

```mermaid
graph TD
    subgraph Frontend
        UI[React Pages]
        Store[Zustand AuthStore]
        Client[Axios API Client]
    end

    subgraph Backend
        Router[Gin Router]
        Middleware[JWT Auth Middleware]
        Handlers[Handlers]
        Services[Services]
        Repos[Repositories]
    end

    subgraph Infraestructura
        DB[(PostgreSQL)]
        Storage[(MinIO)]
    end

    UI --> Store
    UI --> Client
    Client --> Router
    Router --> Middleware
    Middleware --> Handlers
    Handlers --> Services
    Handlers --> Repos
    Services --> Repos
    Repos --> DB
    Services --> Storage
```

---

## 4. Flujo de autenticación

```mermaid
sequenceDiagram
    actor U as Usuario
    participant FE as Frontend
    participant BE as Backend
    participant DB as PostgreSQL

    U->>FE: Login (usuario + password)
    FE->>BE: POST /auth/login
    BE->>DB: Buscar usuario, verificar password
    BE-->>FE: access_token + refresh_token
    FE->>FE: Guardar tokens en Zustand/localStorage

    Note over FE,BE: Token expirado

    FE->>BE: POST /auth/refresh
    BE->>DB: Validar y rotar refresh token
    BE-->>FE: Nuevos tokens
```

---

## 5. Roles y permisos

```mermaid
graph LR
    V[Viewer] -->|lee| P[Productos / Categorías / Dashboard]
    M[Manager] -->|todo Viewer +| CRUD[Crear / Editar productos e imágenes]
    A[Admin] -->|todo Manager +| Users[Gestionar usuarios / Eliminar categorías]
```
