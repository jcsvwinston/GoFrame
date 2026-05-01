# Quark ORM

Quark es un ORM (Object-Relational Mapping) moderno, ligero y fuertemente tipado para Go, diseñado en base a Generics. Ofrece una API fluida y segura, optimizada para rendimiento y facilidad de uso sin los típicos cuellos de botella de la reflexión continua.

## Características Principales

*   **100% Type-Safe**: Desarrollado aprovechando Go Generics. Evita errores en tiempo de ejecución devolviendo directamente tus structs.
*   **Seguridad por Diseño (SQLGuard)**: Verificación estricta de nombres de columnas, operadores y palabras clave en tiempo real para evitar SQL Injection.
*   **Caché de Reflexión**: Metadatos analizados una sola vez por modelo (al inicio) para un mapeo `O(1)` ultrarrápido y consultas sin latencia de _reflection_.
*   **Builder Inmutable**: Construye consultas dinámicas concurrentemente de forma segura; los métodos no mutan la instancia original sino que clonan el estado.
*   **Soporte Multidialecto**: Soporte automático para PostgreSQL, MySQL/MariaDB y SQLite.
*   **Transacciones Robustas**: API dual (Callbacks automáticos y manuales con soporte nativo de `Savepoints`).
*   **Ejecución Interceptable**: Arquitectura de `Middleware` nativa para logs, métricas, caches y reintentos en todos los métodos (Lecturas y Escrituras).

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

### Update (Actualizar)
*Nota: Update() actualiza todos los campos de la entidad.*
```go
user, _ := quark.For[User](ctx, client).Find(1)
user.Name = "Bob"
err := quark.For[User](ctx, client).Update(&user)
```

Si sólo quieres actualizar campos específicos o múltiples filas, usa `UpdateMap`:
```go
// UPDATE users SET active = 0 WHERE name = 'Alice'
affected, err := quark.For[User](ctx, client).
    Where("name", "=", "Alice").
    UpdateMap(map[string]any{"active": false})
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

## 🧬 Relaciones (Lazy Loading)

Quark **V1** emplea carga de relaciones bajo demanda ("Lazy Explicit"). En lugar de auto-inyectar la estructura usando reflection compleja y poco predecible, realizas el fetch de relaciones tú mismo cuando lo necesites.

```go
// 1. Cargamos el usuario
user, _ := quark.For[User](ctx, client).Find(1)

// 2. Cargamos sus pedidos cuando hacemos falta
orders, err := quark.For[Order](ctx, client).Where("user_id", "=", user.ID).List()
```

---

## 🔌 Middleware y Observers

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
