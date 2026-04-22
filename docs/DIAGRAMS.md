# Electroteca — Diagramas Técnicos

> Documentación unificada (backend Go + frontend React). Todos los diagramas usan sintaxis Mermaid.

---

## 1a. Diagrama de Clases — Backend

```mermaid
classDiagram
    %% ─── BACKEND: DOMAIN MODELS ───────────────────────────────────────────

    class User {
        +UUID ID
        +string Username
        +string Email
        +string PasswordHash
        +Role Role
        +bool Active
        +*time.Time LastLogin
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +gorm.DeletedAt DeletedAt
        --
        +CanManage() bool
        +IsAdmin() bool
        +BeforeCreate()
    }

    class Product {
        +UUID ID
        +string Name
        +string RepairDescription
        +string RepairReference
        +*DateOnly EntryDate
        +*DateOnly ExitDate
        +string Observations
        +float64 Price
        +string SKU
        +string ImageURL
        +string ImageKey
        +*UUID CategoryID
        +UUID CreatedByID
        +bool Paid
        +ProductStatus Status
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +gorm.DeletedAt DeletedAt
        --
        +BeforeCreate()
    }

    class Category {
        +UUID ID
        +string Name
        +string Description
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +gorm.DeletedAt DeletedAt
        --
        +BeforeCreate()
    }

    class Contact {
        +UUID ID
        +UUID ProductID
        +string Name
        +string Subdato
        +string Email
        +string Phone
        +time.Time CreatedAt
        +time.Time UpdatedAt
        +gorm.DeletedAt DeletedAt
        --
        +BeforeCreate()
    }

    class ProductImage {
        +UUID ID
        +UUID ProductID
        +string ImageURL
        +string ImageKey
        +int Position
        +time.Time CreatedAt
        +gorm.DeletedAt DeletedAt
        --
        +BeforeCreate()
    }

    class RefreshToken {
        +UUID ID
        +UUID UserID
        +string TokenHash
        +time.Time ExpiresAt
        +bool Revoked
        +time.Time CreatedAt
        +gorm.DeletedAt DeletedAt
        --
        +BeforeCreate()
        +IsValid() bool
    }

    class TokenPair {
        +string AccessToken
        +string RefreshToken
        +time.Time ExpiresAt
    }

    class Claims {
        +UUID UserID
        +string Username
        +Role Role
        +RegisteredClaims embedded
    }

    %% ─── BACKEND: SERVICES ─────────────────────────────────────────────────

    class AuthService {
        -userRepository userRepo
        -tokenRepository tokenRepo
        -[]byte jwtSecret
        -time.Duration accessTTL
        -time.Duration refreshTTL
        --
        +Login(identifier, password) TokenPair User error
        +Register(username, email, password) TokenPair User error
        +Refresh(rawToken) TokenPair User error
        +Logout(rawToken) error
        +LogoutAll(userID) error
        +ValidateAccessToken(tokenString) Claims error
        +CheckPassword(hash, plain) bool
        +HashPassword(password) string error
        -issueTokenPair(user) TokenPair error
        -generateOpaqueToken() string error
        -hashToken(raw) string
        -validatePassword(password) error
    }

    class MinIOService {
        -minio.Client client
        -string bucket
        -int64 maxBytes
        -url.URL publicURL
        --
        +Ping() bool
        +UploadProductImage(file, header) string error
        +GetPresignedURL(objectKey, expiry) string error
        +DeleteObject(objectKey)
        -ensureBucket() error
        -sniffMIME(data) string
    }

    %% ─── BACKEND: HANDLERS ─────────────────────────────────────────────────

    class AuthHandler {
        -AuthService authSvc
        -userRepository userRepo
        --
        +Login(c)
        +Register(c)
        +Refresh(c)
        +Logout(c)
        +LogoutAll(c)
        +Me(c)
        +UpdateMe(c)
    }

    class ProductHandler {
        -productRepository productRepo
        -categoryRepository categoryRepo
        -MinIOService minioSvc
        --
        +List(c)
        +Get(c)
        +Create(c)
        +Update(c)
        +Delete(c)
        +UploadImage(c)
        +AddImage(c)
        +DeleteImage(c)
        -enrichImageURLs(products)
        -parseDateField(raw) DateOnly error
    }

    class CategoryHandler {
        -categoryRepository categoryRepo
        --
        +List(c)
        +Get(c)
        +Create(c)
        +Update(c)
        +Delete(c)
    }

    class UserHandler {
        -userRepository userRepo
        -AuthService authSvc
        --
        +List(c)
        +Get(c)
        +Create(c)
        +Update(c)
        +Delete(c)
    }

    class ContactHandler {
        -contactRepository contactRepo
        -productRepository productRepo
        --
        +Get(c)
        +Upsert(c)
        +Delete(c)
    }

    class StatsHandler {
        -db gorm.DB
        --
        +GetStats(c)
    }

    class HealthHandler {
        -db gorm.DB
        -minioCheck func bool
        --
        +Health(c)
        +Live(c)
    }

    %% ─── ENUMS ──────────────────────────────────────────────────────────────

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

    %% ─── RELACIONES ─────────────────────────────────────────────────────────

    User "1" --> "0..*" Product : creates
    User "1" --> "0..*" RefreshToken : has
    Product "1" --> "0..1" Contact : has
    Product "1" --> "0..*" ProductImage : has gallery
    Product "0..*" --> "0..1" Category : belongs to
    User --> Role : has
    Product --> ProductStatus : has

    AuthService --> TokenPair : produces
    AuthService --> Claims : validates
    AuthHandler --> AuthService : uses
    ProductHandler --> MinIOService : uses
    AuthHandler ..> User : returns
    ProductHandler ..> Product : returns
    CategoryHandler ..> Category : returns
    UserHandler ..> User : returns
    ContactHandler ..> Contact : returns
```

---

## 1b. Diagrama de Clases — Frontend

```mermaid
classDiagram
    %% ─── STORES ─────────────────────────────────────────────────────────────

    class AuthStore {
        +User user
        +string accessToken
        +string refreshToken
        +Date expiresAt
        +bool isAuthenticated
        --
        +setAuth(user, accessToken, refreshToken, expiresAt)
        +updateUser(user)
        +logout()
    }

    %% ─── API CLIENTS ─────────────────────────────────────────────────────────

    class AuthApi {
        +login(identifier, password) TokenPair
        +register(data) TokenPair
        +refresh(refreshToken) TokenPair
        +logout(refreshToken?)
        +me() User
        +updateMe(data) User
    }

    class ProductsApi {
        +list(filters?) PaginatedResponse
        +get(id) Product
        +create(data) Product
        +update(id, data) Product
        +delete(id)
        +addImage(id, file) ProductImage
        +deleteImage(productId, imageId)
    }

    class CategoriesApi {
        +list() ListResponse
        +get(id) Category
        +create(data) Category
        +update(id, data) Category
        +delete(id)
    }

    class UsersApi {
        +list() ListResponse
        +get(id) User
        +create(data) User
        +update(id, data) User
        +delete(id)
    }

    class ContactsApi {
        +get(productId) Contact
        +upsert(productId, data) Contact
        +delete(productId)
    }

    class StatsApi {
        +get() InventoryStats
    }

    %% ─── HOOKS ───────────────────────────────────────────────────────────────

    class UsePermissions {
        +bool isAdmin
        +bool isManager
        +bool isViewer
        +bool canManage
        +canDeleteProduct(product) bool
    }

    class UseDebounce {
        +debounce(value, delay) debouncedValue
    }

    %% ─── PAGES ───────────────────────────────────────────────────────────────

    class LoginPage {
        -string identifier
        -string password
        --
        +handleSubmit()
    }

    class RegisterPage {
        -string username
        -string email
        -string password
        --
        +handleSubmit()
    }

    class DashboardPage

    class ProductsPage {
        -ProductFilters filters
        -string viewMode
        --
        +handleSearch()
        +handleFilterChange()
        +handleCreate()
        +handleEdit()
        +handleDelete()
    }

    class ProductDetailPage {
        -string productId
        --
        +handleEdit()
        +handleDelete()
        +handleImageUpload()
    }

    class CategoriesPage {
        --
        +handleCreate()
        +handleEdit()
        +handleDelete()
    }

    class UsersPage {
        --
        +handleCreate()
        +handleEdit()
        +handleDelete()
    }

    %% ─── RELACIONES ─────────────────────────────────────────────────────────

    LoginPage --> AuthStore : updates
    RegisterPage --> AuthStore : updates
    LoginPage --> AuthApi : calls
    RegisterPage --> AuthApi : calls
    ProductsPage --> ProductsApi : calls
    ProductsPage --> CategoriesApi : calls
    ProductDetailPage --> ProductsApi : calls
    ProductDetailPage --> ContactsApi : calls
    CategoriesPage --> CategoriesApi : calls
    UsersPage --> UsersApi : calls
    DashboardPage --> StatsApi : calls

    ProductsPage --> UsePermissions : uses
    ProductDetailPage --> UsePermissions : uses
    CategoriesPage --> UsePermissions : uses
    UsersPage --> UsePermissions : uses

    AuthStore --> AuthApi : refreshes token via
```

---

## 2. Casos de Uso

### 2.1 — Viewer

```mermaid
graph LR
    V([Viewer])

    subgraph Auth
        A1[Iniciar sesión]
        A2[Registrarse]
        A3[Cerrar sesión]
        A4[Ver / Editar perfil propio]
    end

    subgraph Productos
        P1[Ver listado]
        P2[Buscar y filtrar]
        P3[Ver detalle]
    end

    subgraph Contactos
        C1[Ver contacto del producto]
    end

    subgraph Categorias
        G1[Ver categorías]
    end

    subgraph Dashboard
        D1[Ver estadísticas]
        D2[Ver productos recientes]
    end

    V --> A1
    V --> A2
    V --> A3
    V --> A4
    V --> P1
    V --> P2
    V --> P3
    V --> C1
    V --> G1
    V --> D1
    V --> D2
```

### 2.2 — Manager

```mermaid
graph LR
    M([Manager])

    subgraph Auth
        A1[Iniciar sesión]
        A3[Cerrar sesión]
        A4[Ver / Editar perfil propio]
    end

    subgraph Productos
        P1[Ver listado]
        P2[Buscar y filtrar]
        P3[Ver detalle]
        P4[Crear producto]
        P5[Editar producto]
        P6[Eliminar producto propio]
        P7[Gestionar imágenes]
    end

    subgraph Contactos
        C1[Ver contacto]
        C2[Crear / Editar contacto]
        C3[Eliminar contacto]
    end

    subgraph Categorias
        G1[Ver categorías]
        G2[Crear categoría]
        G3[Editar categoría]
    end

    subgraph Dashboard
        D1[Ver estadísticas]
        D2[Ver productos recientes]
    end

    M --> A1
    M --> A3
    M --> A4
    M --> P1
    M --> P2
    M --> P3
    M --> P4
    M --> P5
    M --> P6
    M --> P7
    M --> C1
    M --> C2
    M --> C3
    M --> G1
    M --> G2
    M --> G3
    M --> D1
    M --> D2
```

### 2.3 — Admin

```mermaid
graph LR
    A([Admin])

    subgraph Auth
        A1[Iniciar sesión]
        A3[Cerrar sesión]
        A4[Ver / Editar perfil propio]
    end

    subgraph Productos
        P4[Crear producto]
        P5[Editar cualquier producto]
        P6[Eliminar cualquier producto]
        P7[Gestionar imágenes]
    end

    subgraph Contactos
        C2[Crear / Editar contacto]
        C3[Eliminar contacto]
    end

    subgraph Categorias
        G2[Crear categoría]
        G3[Editar categoría]
        G4[Eliminar categoría]
    end

    subgraph Usuarios
        U1[Ver lista de usuarios]
        U2[Crear usuario]
        U3[Editar usuario]
        U4[Eliminar usuario]
    end

    subgraph Dashboard
        D1[Ver estadísticas]
        D2[Ver productos recientes]
    end

    A --> A1
    A --> A3
    A --> A4
    A --> P4
    A --> P5
    A --> P6
    A --> P7
    A --> C2
    A --> C3
    A --> G2
    A --> G3
    A --> G4
    A --> U1
    A --> U2
    A --> U3
    A --> U4
    A --> D1
    A --> D2
```

---

## 3. Diagrama de Secuencia

### 3.1 — Login con refresh token automático

```mermaid
sequenceDiagram
    actor U as Usuario
    participant FE as Frontend (React)
    participant AS as AuthStore (Zustand)
    participant API as API Client (axios)
    participant BE as Backend (Gin)
    participant DB as PostgreSQL

    U->>FE: Ingresa identifier + password
    FE->>API: POST /auth/login {identifier, password}
    API->>BE: POST /api/v1/auth/login
    BE->>DB: SELECT users WHERE username=? OR email=?
    DB-->>BE: User row
    BE->>BE: bcrypt.CompareHashAndPassword()
    BE->>DB: INSERT refresh_tokens (hash, expires_at)
    DB-->>BE: OK
    BE-->>API: {access_token, refresh_token, expires_at, user}
    API-->>FE: TokenPair + User
    FE->>AS: setAuth(user, accessToken, refreshToken, expiresAt)
    AS->>AS: Persistir en localStorage
    FE-->>U: Redirigir a /dashboard

    Note over FE,BE: ── Token expirado → refresh automático ──

    FE->>API: GET /products (access_token expirado)
    API->>BE: GET /api/v1/products
    BE-->>API: 401 Unauthorized
    API->>API: Encolar requests pendientes
    API->>BE: POST /api/v1/auth/refresh {refresh_token}
    BE->>DB: SELECT refresh_tokens WHERE hash=? AND NOT revoked
    DB-->>BE: RefreshToken row
    BE->>BE: IsValid() → true
    BE->>DB: UPDATE refresh_tokens SET revoked=true (token rotation)
    BE->>DB: INSERT nuevo refresh_token
    DB-->>BE: OK
    BE-->>API: {access_token, refresh_token, expires_at}
    API->>AS: setAuth(nuevos tokens)
    API->>BE: GET /api/v1/products (con nuevo access_token)
    BE-->>API: {data: [...products]}
    API-->>FE: Products data
```

### 3.2 — Crear producto con imagen y contacto

```mermaid
sequenceDiagram
    actor U as Manager/Admin
    participant FE as ProductForm (React)
    participant API as API Client
    participant BE as Backend (Gin)
    participant DB as PostgreSQL
    participant MN as MinIO

    U->>FE: Rellena formulario + imagen + contacto
    FE->>FE: react-hook-form validate()

    FE->>API: POST /api/v1/products {name, description, ...}
    API->>BE: POST /products (JWT auth middleware)
    BE->>BE: RequireManager() check
    BE->>DB: INSERT products (auto-genera SKU "REP-XXXXXX")
    DB-->>BE: Product con ID
    BE-->>API: Product (201)
    API-->>FE: Product created

    FE->>API: POST /api/v1/products/:id/images (multipart)
    API->>BE: POST /products/:id/images
    BE->>BE: ValidarMIME (jpg/png/webp) + MaxSize
    BE->>MN: PutObject("products/{uuid}.ext")
    MN-->>BE: ObjectKey
    BE->>DB: INSERT product_images (product_id, image_key, position)
    DB-->>BE: ProductImage
    BE-->>API: ProductImage (201)
    API-->>FE: Image uploaded

    FE->>API: PUT /api/v1/products/:id/contact {name, email, phone}
    API->>BE: PUT /products/:id/contact
    BE->>DB: UPSERT contacts (ON CONFLICT product_id DO UPDATE)
    DB-->>BE: Contact
    BE-->>API: Contact
    API-->>FE: Contact saved

    FE->>FE: invalidateQueries(['products'])
    FE-->>U: Toast "Producto creado" + cerrar modal
```

### 3.3 — Logout y cierre de sesión global

```mermaid
sequenceDiagram
    actor U as Usuario
    participant FE as Frontend
    participant AS as AuthStore
    participant API as API Client
    participant BE as Backend
    participant DB as PostgreSQL

    U->>FE: Click "Cerrar sesión"
    FE->>API: POST /api/v1/auth/logout {refresh_token}
    API->>BE: POST /auth/logout
    BE->>BE: hashToken(rawToken)
    BE->>DB: UPDATE refresh_tokens SET revoked=true WHERE hash=?
    DB-->>BE: OK
    BE-->>API: {message: "logged out"}
    API-->>FE: OK
    FE->>AS: logout()
    AS->>AS: Limpiar localStorage
    FE-->>U: Redirigir a /login

    Note over U,DB: ── Logout global (todos los dispositivos) ──

    U->>FE: Click "Cerrar todas las sesiones"
    FE->>API: POST /api/v1/auth/logout-all
    API->>BE: POST /auth/logout-all (JWT auth)
    BE->>DB: UPDATE refresh_tokens SET revoked=true WHERE user_id=?
    DB-->>BE: OK (N tokens revocados)
    BE-->>API: {message: "all sessions logged out"}
    FE->>AS: logout()
    FE-->>U: Redirigir a /login
```

---

## 4. Diagrama Entidad-Relación (ER)

```mermaid
erDiagram
    users {
        UUID id PK
        VARCHAR username UK
        VARCHAR email UK
        VARCHAR password_hash
        VARCHAR role "admin|manager|viewer"
        BOOLEAN active
        TIMESTAMPTZ last_login "nullable"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at "nullable, soft delete"
    }

    categories {
        UUID id PK
        VARCHAR name UK
        TEXT description "nullable"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at "nullable, soft delete"
    }

    products {
        UUID id PK
        VARCHAR name
        TEXT repair_description
        VARCHAR repair_reference "nullable"
        DATE entry_date "nullable"
        DATE exit_date "nullable"
        TEXT observations "nullable"
        NUMERIC price "default 0"
        VARCHAR sku UK "REP-XXXXXX"
        VARCHAR image_key "nullable, MinIO key"
        UUID category_id FK "nullable"
        UUID created_by_id FK
        BOOLEAN paid "default false"
        VARCHAR status "reparado|en_progreso|no_reparado"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at "nullable, soft delete"
    }

    contacts {
        UUID id PK
        UUID product_id FK "unique - 1:1 con product"
        VARCHAR name
        VARCHAR subdato "legacy, nullable"
        VARCHAR email "nullable"
        VARCHAR phone "nullable"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ updated_at
        TIMESTAMPTZ deleted_at "nullable, soft delete"
    }

    product_images {
        UUID id PK
        UUID product_id FK
        VARCHAR image_key "nullable, MinIO key"
        INT position "default 0, ordering"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ deleted_at "nullable, soft delete"
    }

    refresh_tokens {
        UUID id PK
        UUID user_id FK
        VARCHAR token_hash UK "SHA-256"
        TIMESTAMPTZ expires_at
        BOOLEAN revoked "default false"
        TIMESTAMPTZ created_at
        TIMESTAMPTZ deleted_at "nullable, soft delete"
    }

    %% Relaciones
    users ||--o{ products : "crea (created_by_id)"
    users ||--o{ refresh_tokens : "posee"
    categories ||--o{ products : "clasifica"
    products ||--o| contacts : "tiene (1:1)"
    products ||--o{ product_images : "tiene galería"
```
