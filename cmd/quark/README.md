# Quark CLI

La interfaz de línea de comandos de Quark ORM diseñada para automatizar tareas repetitivas y gestionar el ciclo de vida de tus aplicaciones.

## Instalación

```bash
go build -o quark ./cmd/quark/main.go
# Opcionalmente, muévelo a tu bin path
mv quark /usr/local/bin/
```

## Comandos Principales

### 1. Inicialización (`init`)
Prepara un nuevo proyecto con la configuración necesaria.
```bash
quark init --dir ./mi-proyecto --dialect postgres
```
Esto creará:
- `.quark.yml`: Archivo de configuración central.
- `models/`: Directorio para los modelos generados.
- `migrations/`: Directorio para las migraciones versionadas.
- `seeders/`: Directorio para los seeders de datos.

### 2. Generación de Modelos (`model generate`)
Genera structs de Go a partir de tablas existentes o definiciones rápidas.

**Desde Base de Datos:**
```bash
quark model gen --from-table users,orders
```

**Desde Definición (rápido):**
```bash
quark model gen User --fields "id:int64,name:string,email:string,active:bool"
```

### 3. Migraciones (`migrate`)
Sistema de migraciones basado en código Go para máximo control.

- **Crear:** `quark migrate create add_profile_to_users --message "Add bio field"`
- **Aplicar:** `quark migrate up`
- **Revertir:** `quark migrate down --steps 1`
- **Estado:** `quark migrate status`
- **Versión:** `quark migrate version`

### 4. Multi-Tenancy (`tenant`)
Facilita la gestión de entornos con múltiples clientes.

- **Provisionar:** `quark tenant provision "tenant_alpha"` (Crea BD/Esquema y corre migraciones).
- **Migrar:** `quark tenant migrate "tenant_alpha"`

### 5. Inspección (`inspect`)
Visualiza el estado de tu base de datos sin salir de la terminal.
```bash
quark inspect table users
```

## Configuración (`.quark.yml`)

El CLI utiliza un archivo YAML para conocer los detalles de conexión y rutas:

```yaml
database:
  default:
    driver: postgres
    dsn: "postgres://user:pass@localhost:5432/dbname?sslmode=disable"

paths:
  models: "./models"
  migrations: "./migrations"
  seeders: "./seeders"
```

También soporta variables de entorno prefijadas con `QUARK_` (ej: `QUARK_DATABASE_DEFAULT_DSN`).
