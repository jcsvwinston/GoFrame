# Quark ORM

Quark es un ORM (Object-Relational Mapping) moderno, ligero y fuertemente tipado para Go, diseñado en base a Generics. Ofrece una API fluida y segura, optimizada para rendimiento y facilidad de uso sin los típicos cuellos de botella de la reflexión continua.

## Características Principales

*   **100% Type-Safe**: Desarrollado aprovechando Go Generics. Evita errores en tiempo de ejecución devolviendo directamente tus structs.
*   **Seguridad por Diseño (SQLGuard)**: Verificación estricta de nombres de columnas, operadores y palabras clave en tiempo real para evitar SQL Injection.
*   **Caché de Reflexión**: Metadatos analizados una sola vez por modelo (al inicio) para un mapeo `O(1)` ultrarrápido y consultas sin latencia de _reflection_.
*   **Builder Inmutable**: Construye consultas dinámicas concurrentemente de forma segura; los métodos no mutan la instancia original sino que clonan el estado.
*   **Soporte Multidialecto**: Soporte nativo y optimizado para **PostgreSQL, MySQL/MariaDB, SQLite, MSSQL y Oracle**.
*   **Transacciones Robustas**: API dual (Callbacks automáticos y manuales con soporte nativo de `Savepoints`).
*   **Auto-Migraciones Evolutivas (`Sync`)**: Sincronización inteligente de esquemas (`client.Sync()`) con soporte para adición de columnas, renombrado dinámico (`quark:"rename:old_name"`) y borrado destructivo controlado.
*   **Validación Integrada**: Soporte nativo para `validator/v10` y validación programática (interfaz `Validatable`) interceptando llamadas CRUD.
*   **Persistencia de Asociaciones Recursiva**: Guarda grafos de objetos profundos (HasMany, BelongsTo, ManyToMany) en una sola operación de `Create` o `Update`.
*   **Eager Loading**: Carga eficiente de relaciones con `Preload()` evitando el problema N+1.
*   **Multi-Tenant**: `TenantRouter` nativo con soporte para aislamiento por Base de Datos, Esquemas, y Seguridad a Nivel de Fila (RLS).
*   **Ejecución Interceptable**: Arquitectura de `Middleware` nativa, `Hooks` de ciclo de vida (`BeforeCreate`, `AfterUpdate`) para interceptores robustos.
*   **Ejecución Nativa**: Soporte robusto y unificado para Funciones, Procedimientos Almacenados (con parámetros `OUT`) y Eventos (PubSub).

---

## 📖 Documentación Adicional

Para más detalles sobre el diseño y capacidades avanzadas de Quark, consulta:

*   [**Arquitectura**](docs/ARCHITECTURE.md): Principios de diseño, ciclo de vida de peticiones y flujo interno.
*   [**Multi-Tenant**](docs/MULTI-TENANT-CONSIDERATIONS.md): Guía detallada sobre estrategias de aislamiento y configuración.
*   [**Roadmap**](docs/ROADMAP.md): Estado actual del proyecto y planes de desarrollo futuro.

---

## 🚀 Inicialización (Client)

Para usar Quark, instancia un `Client` pasándole una conexión nativa `*sql.DB`:

```go
package main

import (
    "database/sql"
    _ "modernc.org/sqlite"
    "github.com/jcsvwinston/GoFrame/pkg/quark"
)

func main() {
    db, _ := sql.Open("sqlite", "file:data.db?cache=shared")
    
    client, err := quark.New(db, 
        quark.WithDialect(quark.SQLite()), // Opcional: Autodetectado por defecto
    )
    if err != nil {
        panic(err)
    }
    defer client.Close()
}
```

---

## 📦 Definición de Modelos

Quark detecta automáticamente el nombre de la tabla (pluralizando y en *snake_case*) y tu clave primaria (por defecto el campo con tag `db:"id"` o especificando `pk:"true"`).

```go
type User struct {
    ID        int64     `db:"id" pk:"true"` // Opcional: pk:"true" para IDs personalizados
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    Active    bool      `db:"active"`
    CreatedAt time.Time `db:"created_at"`
}
```

### Clave Primaria Compuesta (Composite PK)

Añade `pk:"true"` a **múltiples campos** para declarar una PK compuesta. Quark detecta esto automáticamente y genera las cláusulas `WHERE col1 = $1 AND col2 = $2` correctas en `Update`, `Delete` y `HardDelete`.

```go
// Tabla de unión típica (ej. order_items)
type OrderItem struct {
    OrderID   int64  `db:"order_id"   pk:"true"`
    ProductID int64  `db:"product_id" pk:"true"`
    Qty       int    `db:"qty"`
}

// También funciona con claves string
type RolePermission struct {
    RoleID       string `db:"role_id"       pk:"true"`
    PermissionID string `db:"permission_id" pk:"true"`
    Granted      bool   `db:"granted"`
}
```

```go
// Create — incluye siempre todas las columnas PK (no son auto-increment)
item := OrderItem{OrderID: 1, ProductID: 10, Qty: 5}
err := quark.For[OrderItem](ctx, client).Create(&item)

// Update — WHERE order_id = $1 AND product_id = $2
item.Qty = 99
rows, err := quark.For[OrderItem](ctx, client).Update(&item)

// HardDelete — DELETE FROM order_items WHERE order_id = $1 AND product_id = $2
rows, err := quark.For[OrderItem](ctx, client).HardDelete(&item)

// List con filtro parcial (solo una columna de la PK)
items, err := quark.For[OrderItem](ctx, client).Where("order_id", "=", 1).List()
```

> **Nota:** `Find(id)` utiliza la PK simple (`ModelMeta.PK`). Para modelos con composite PK usa `Where(...).First()` o `Where(...).List()`.

---

## 🛠 Operaciones CRUD Básicas

El punto de entrada para trabajar con un modelo es `quark.For[Model](ctx, client)`.

### Create (Insertar)
Inserta un registro. Dependiendo del dialecto y base de datos, poblará automáticamente el ID devuelto en el struct.
```go
user := User{Name: "Alice", Email: "alice@example.com", Active: true}
err := quark.For[User](ctx, client).Create(&user)
```

### Read (Consultar)
```go
// Encontrar por ID
user, err := quark.For[User](ctx, client).Find(1)

// Obtener el primer resultado
first, err := quark.For[User](ctx, client).Where("active", "=", true).First()

// Listar resultados
users, err := quark.For[User](ctx, client).Limit(10).List()
```

### Update (Actualizar — Parcial)

> **Importante:** `Update()` realiza una actualización **parcial**: sólo actualiza los campos con valor no-zero en el struct. Los campos con su valor cero (0, "", false, nil) **no se modifican** en base de datos. Para actualizar a valores zero de forma explícita, usa `UpdateMap`.

```go
user, _ := quark.For[User](ctx, client).Find(1)
user.Name = "Bob"  // Solo este campo será actualizado
err := quark.For[User](ctx, client).Update(&user)
// SQL: UPDATE users SET name = 'Bob' WHERE id = 1
```

Para actualizar campos específicos o forzar valores zero, usa `UpdateMap`:
```go
// UPDATE users SET active = false, name = 'Evelyn' WHERE id = 1
affected, err := quark.For[User](ctx, client).
    Where("id", "=", user.ID).
    UpdateMap(map[string]any{"active": false, "name": "Evelyn"})
```

### Upsert (INSERT … ON CONFLICT)

Inserta o actualiza un registro cuando se produce un conflicto en las columnas indicadas. Compatible con los 6 dialectos (Postgres, MySQL, MariaDB, SQLite, MSSQL, Oracle).

```go
user := User{Email: "alice@example.com", Name: "Alice", Age: 30}
// Inserta; si ya existe un registro con ese email, actualiza name y age
err := quark.For[User](ctx, client).Upsert(&user, []string{"email"}, []string{"name", "age"})
```

### CreateBatch (Inserción masiva)

Inserta múltiples registros en una sola sentencia SQL.

```go
users := []*User{
    {Name: "Alice", Email: "alice@example.com"},
    {Name: "Bob",   Email: "bob@example.com"},
}
err := quark.For[User](ctx, client).CreateBatch(users)
```

### Delete (Eliminar)
```go
// Eliminar un struct (usando su PK)
err := quark.For[User](ctx, client).Delete(&user)

// Eliminar mediante condición
affected, err := quark.For[User](ctx, client).Where("active", "=", false).DeleteBy()
```

---

## 🔍 Consultas Avanzadas (Query Builder)

El Query Builder de Quark es inmutable, permitiéndote reutilizar la base de tu consulta en distintos hilos o ramas sin provocar "Data Races".

### Where, WhereIn, WhereBetween y OR

```go
// AND implícito
q := quark.For[User](ctx, client).
    Where("active", "=", true).
    Where("age", ">", 18)

// IN y BETWEEN
q = q.WhereIn("role", []any{"admin", "editor"}).
    WhereBetween("created_at", start, end)

// Condiciones OR agrupadas
users, err := q.Or(func(oq *quark.Query[User]) *quark.Query[User] {
    return oq.Where("name", "=", "John").Where("email", "LIKE", "%@acme.com")
}).List()
// SQL: WHERE active = 1 AND age > 18 ... OR (name = 'John' AND email LIKE '%@acme.com')
```

### Ordering, Limiting y Count

```go
// Orden y Paginación simple
users, err := quark.For[User](ctx, client).
    OrderBy("created_at", "DESC").
    Limit(10).
    Offset(20).
    List()

// Contar registros
count, err := quark.For[User](ctx, client).Where("active", "=", true).Count()
```

### WhereNot

Niega una condición añadiendo `NOT` al predicado:

```go
// WHERE NOT active = true  →  todos los usuarios inactivos
users, _ := quark.For[User](ctx, client).WhereNot("active", "=", true).List()
```

### Distinct

```go
// SELECT DISTINCT name FROM users
users, _ := quark.For[User](ctx, client).Select("name").Distinct().List()
```

### GroupBy y Having

```go
// SELECT department, COUNT(*) FROM employees GROUP BY department HAVING age > 30
rows, _ := quark.For[Employee](ctx, client).
    Select("department").
    GroupBy("department").
    Having("age", ">", 30).
    List()
```

### Agregaciones (Sum, Avg, Min, Max)

```go
totalSales, _ := quark.For[Order](ctx, client).Sum("amount")
avgAge,     _ := quark.For[User](ctx, client).Where("active", "=", true).Avg("age")
oldest,     _ := quark.For[User](ctx, client).Max("age")
youngest,   _ := quark.For[User](ctx, client).Min("age")
```

### Scopes (Consultas Reutilizables)

Encapsula filtros frecuentes como funciones y aplícalos con `Apply()`:

```go
activeOnly := quark.Scope[User](func(q *quark.Query[User]) *quark.Query[User] {
    return q.Where("active", "=", true)
})
adults := quark.Scope[User](func(q *quark.Query[User]) *quark.Query[User] {
    return q.Where("age", ">=", 18)
})

users, _ := quark.For[User](ctx, client).Apply(activeOnly, adults).List()
```

### Subqueries en WHERE

```go
// WHERE id IN (SELECT user_id FROM orders WHERE total > 100)
users, _ := quark.For[User](ctx, client).
    WhereSubquery("id", "IN", `SELECT user_id FROM orders WHERE total > 100`).
    List()
```

---

## 🔗 JOINs

Puedes añadir `Join`, `LeftJoin` y `RightJoin` a tus consultas. 
*(Nota: Para recuperar todos los datos de tablas unidas en un struct de forma plana, asegúrate de que los campos cruzados coincidan o usa vistas).*

```go
usersWithOrders, err := quark.For[User](ctx, client).
    Select("users.id", "users.name", "orders.amount").
    Join("orders", "orders.user_id = users.id").
    Where("orders.status", "=", "paid").
    List()
```

---

## 🔄 Transacciones y Savepoints

Quark expone una API limpia para el control transaccional de la BD.

### Estilo Callback (Recomendado)
Maneja el commit o rollback automáticamente (e intercepta los panics encapsulándolos en un rollback seguro).

```go
err := client.Tx(ctx, func(tx *quark.Tx) error {
    u := User{Name: "Charlie", Email: "charlie@tx.com"}
    if err := quark.ForTx[User](ctx, tx).Create(&u); err != nil {
        return err // Esto activa un ROLLBACK
    }
    
    // Savepoints soportados
    tx.Savepoint("mi_punto")
    
    return nil // Esto activa el COMMIT final
})
```

### Estilo Manual
```go
tx, err := client.BeginTx(ctx, nil)
defer tx.Rollback()

quark.ForTx[User](ctx, tx).Create(&u1)
tx.Commit()
```

---

## 📚 Datasets Grandes: Streaming e Iteradores

Para prevenir cuellos de botella de memoria (OOM), no traigas millones de filas con `List()`.

### Paginación Inteligente
```go
page, err := quark.For[User](ctx, client).Paginate(1, 20) // Página 1, 20 items
// page.Items, page.Total, page.TotalPages
```

### Iterador Continuo (Streaming)
```go
err := quark.For[User](ctx, client).Iter(func(user User) error {
    // Procesar uno por uno, la memoria se mantiene constante
    log.Println("Procesando", user.Name)
    return nil
})
```

### Cursor Manual
```go
cursor, err := quark.For[User](ctx, client).Cursor()
defer cursor.Close()

for cursor.Next() {
    var user User
    cursor.Scan(&user)
}
```

---

## 🧬 Relaciones (Eager Loading)

Quark permite evitar el problema de las N+1 consultas resolviendo relaciones en un solo viaje adicional a base de datos utilizando el método `Preload()`.

```go
// Cargar el usuario y también llenar sus posts
user, err := quark.For[User](ctx, client).
    Preload("Posts"). // Asumiendo que User struct tiene un campo `Posts []Post`
    Find(1)
```

Quark utiliza los meta-tags e inferencia de tipos para cargar automáticamente registros que coincidan con la Foreign Key `user_id`.

---

## 🔌 Middleware, Hooks y Observers

### Hooks de Ciclo de Vida
Quark soporta interceptar a nivel de la entidad durante las operaciones CRUD. Tu struct solo necesita implementar la interfaz deseada (`BeforeCreate`, `AfterDelete`, etc).

```go
func (u *User) BeforeCreate(ctx context.Context) error {
    u.CreatedAt = time.Now()
    return nil
}
```

### Observers (Auditoría/Métricas)
Se disparan *después* de que una query se ha completado.
```go
client, _ := quark.New(db, quark.WithQueryObserver(myObserver))
// myObserver debe implementar ObserveQuery(event quark.QueryEvent)
```

### Middleware (Interceptores Completos)
El middleware envuelve `Query`, `QueryRow` y `Exec`, ideal para inyectar *caching*, *rate-limiting*, o trazas (Opentelemetry).

```go
type MyInterceptor struct {
    quark.BaseMiddleware // Embed para heredar métodos base vacíos
}

func (m *MyInterceptor) WrapExec(next quark.ExecFunc) quark.ExecFunc {
    return func(ctx context.Context, exec quark.Executor, sqlStr string, args []any) (sql.Result, error) {
        log.Printf("[Intercept] Ejecutando: %s", sqlStr)
        return next(ctx, exec, sqlStr, args)
    }
}

client, _ := quark.New(db, quark.WithMiddleware(&MyInterceptor{}))
```

---

## ⚙️ Rutinas (Procedimientos, Funciones) y Eventos

Quark ofrece soporte de primera clase para características avanzadas del motor de base de datos de manera agnóstica.

### Ejecución de Funciones (Table-Valued o Escalares)
Si necesitas llamar a una función que devuelve resultados, usa `Routine[T]`:
```go
// SELECT * FROM get_active_users($1) (Postgres) / CALL get_active_users(?) (MySQL)
users, err := quark.NewRoutine[User](ctx, client, "get_active_users", 100).List()

// Función escalar
tax, err := quark.NewRoutine[float64](ctx, client, "calculate_tax", 1500.5).Scalar()
```

### Ejecución de Procedimientos (CALL/EXEC y parámetros OUT)
Si un procedimiento sólo ejecuta lógica y devuelve valores por referencia (`OUT`), utiliza `Call()` nativo:
```go
var procesados int
err := quark.Call(ctx, client, "process_billing", "2026-05", sql.Out{Dest: &procesados})
```

### Eventos (Listen/Notify)
Para publicar y suscribirse a notificaciones de la base de datos (p.ej. Postgres `LISTEN/NOTIFY`), Quark introduce el `EventBus`:
```go
// Publicar
err := quark.Notify(ctx, client, "user_updates", "user_id_1_changed")
```

---

## 🗄 Schema y Migraciones

### Auto-Migrate

Crea la tabla para un modelo si no existe:

```go
// Crear tablas en la base de datos
err := client.Migrate(ctx, &User{}, &Order{})
```

### Auto-Sync (Evolución del esquema)

Sincroniza la tabla con el struct actual — añade columnas nuevas, renombra con `quark:"rename:old_col"`:

```go
err := client.Sync(ctx, &User{})
```

### Tags de columna: NOT NULL, DEFAULT y UNIQUE

```go
type Product struct {
    ID    int64   `db:"id"    pk:"true"`
    Name  string  `db:"name"  quark:"not_null"`
    Price float64 `db:"price" default:"0.00" quark:"not_null"`
    SKU   string  `db:"sku"   quark:"unique"`
}
```

| Tag                  | Efecto en DDL                        |
|----------------------|--------------------------------------|
| `quark:"not_null"`   | `NOT NULL`                           |
| `nullable:"false"`   | `NOT NULL` (alternativa)             |
| `default:"valor"`    | `DEFAULT valor`                      |
| `quark:"unique"`     | `UNIQUE`                             |

### CreateIndex

```go
// CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email)
err := client.CreateIndex(ctx, "users", "idx_users_email", []string{"email"}, true)
```

### AddForeignKey

```go
// ALTER TABLE orders ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
err := client.AddForeignKey(ctx, "orders", "fk_orders_user",
    []string{"user_id"}, "users", []string{"id"}, "CASCADE", "")
```

---

## 🛠 Quark CLI

Quark incluye una potente herramienta de línea de comandos para agilizar el desarrollo:

*   **`init`**: Inicializa un proyecto Quark con estructura recomendada.
*   **`model generate`**: Genera modelos Go automáticamente introspectando tu base de datos.
*   **`migrate`**: Gestión completa de migraciones versionadas (Up, Down, Status, Version).
*   **`tenant`**: Herramientas de provisión y migración para arquitecturas multi-tenant.
*   **`inspect`**: Visualización de esquemas y tablas directamente en la terminal.

Para instalarlo:
```bash
go build -o quark ./cmd/quark/main.go
./quark --help
```

---

## 🏢 Multi-Tenant (TenantRouter)

Quark incorpora un motor nativo y transparente de Multi-Tenant, soportando:
1. `DatabasePerTenant`: Piscinas aisladas de base de datos con caché LRU.
2. `SchemaPerTenant`: Compartición de recursos con aislamiento de namespace.
3. `RowLevelSecurity`: Inyección automática de `WHERE tenant_id = ?`.

Consulta [MULTI-TENANT-CONSIDERATIONS.md](./MULTI-TENANT-CONSIDERATIONS.md) para arquitecturas recomendadas.
