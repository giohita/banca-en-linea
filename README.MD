# 🏦 Sistema de Banca en Línea

Un sistema completo de banca en línea desarrollado con **React** (frontend), **Go** (backend), **PostgreSQL** (base de datos) y **TigerBeetle** (motor de contabilidad de alto rendimiento).

## 📋 Tabla de Contenidos

- [🚀 Inicio Rápido](#-inicio-rápido)
- [🛠️ Tecnologías](#️-tecnologías)
- [📁 Estructura del Proyecto](#-estructura-del-proyecto)
- [⚙️ Configuración del Entorno](#️-configuración-del-entorno)
- [🐳 Instalación con Docker](#-instalación-con-docker)
- [🔐 Credenciales de Prueba](#-credenciales-de-prueba)
- [🌐 URLs del Sistema](#-urls-del-sistema)
- [🧪 Funcionalidades](#-funcionalidades)
- [🔧 Desarrollo Local](#-desarrollo-local)
- [📊 Base de Datos](#-base-de-datos)
- [🚨 Solución de Problemas](#-solución-de-problemas)
- [📝 Logs y Debugging](#-logs-y-debugging)

## 🚀 Inicio Rápido

### Prerrequisitos

- **Docker** y **Docker Compose** instalados
- **Git** para clonar el repositorio
- **Puerto 8082** disponible (frontend)
- **Puerto 8080** disponible (backend)
- **Puerto 5432** disponible (PostgreSQL)

### Instalación en 3 pasos

```bash
# 1. Clonar el repositorio
git clone <url-del-repositorio>
cd banca-en-linea

# 2. Configurar variables de entorno
cp .env.example .env
cp packages/backend/.env.example packages/backend/.env
cp packages/frontend/.env.example packages/frontend/.env

# 3. Levantar el sistema completo
docker-compose up --build
```

¡Listo! El sistema estará disponible en: **http://localhost:8082**

## 🛠️ Tecnologías

### Frontend
- **React 18** con Hooks
- **Vite** como bundler
- **Tailwind CSS** para estilos
- **React Router** para navegación
- **Context API** para manejo de estado

### Backend
- **Go 1.21+** con Gin framework
- **PostgreSQL 15** como base de datos principal
- **TigerBeetle** para contabilidad de alto rendimiento
- **JWT** para autenticación
- **CORS** configurado para desarrollo

### DevOps
- **Docker** y **Docker Compose**
- **Multi-stage builds** para optimización
- **Nginx** como servidor web para producción

## 📁 Estructura del Proyecto

```
banca-en-linea/
├── 📄 .env.example              # Variables de entorno globales
├── 📄 docker-compose.yml       # Configuración de Docker
├── 📄 README.MD                # Este archivo
├── 📁 packages/
│   ├── 📁 backend/             # API en Go
│   │   ├── 📄 .env.example     # Variables del backend
│   │   ├── 📄 main.go          # Punto de entrada
│   │   ├── 📁 internal/        # Lógica de negocio
│   │   ├── 📁 models/          # Modelos de datos
│   │   ├── 📁 database/        # Configuración de DB
│   │   └── 📁 migrations/      # Migraciones SQL
│   ├── 📁 frontend/            # Aplicación React
│   │   ├── 📄 .env.example     # Variables del frontend
│   │   ├── 📁 src/             # Código fuente
│   │   ├── 📁 public/          # Archivos estáticos
│   │   └── 📄 package.json     # Dependencias
│   └── 📁 tigerbeetle/         # Motor de contabilidad
└── 📁 scripts/                 # Scripts de utilidad
```

## ⚙️ Configuración del Entorno

### 1. Variables de Entorno Globales (`.env`)

```bash
# Base de datos PostgreSQL
POSTGRES_DB=banca_db
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres123
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Backend
BACKEND_PORT=8080
JWT_SECRET=tu-clave-secreta-muy-segura-aqui

# Frontend
FRONTEND_PORT=8082
VITE_API_URL=http://localhost:8080

# Docker
COMPOSE_PROJECT_NAME=banca
```

### 2. Variables del Backend (`packages/backend/.env`)

```bash
# Base de datos
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres123
DB_NAME=banca_db

# Servidor
PORT=8080
JWT_SECRET=tu-clave-secreta-muy-segura-aqui
ENVIRONMENT=development

# CORS (para desarrollo)
CORS_ORIGINS=http://localhost:8082,http://localhost:3000
```

### 3. Variables del Frontend (`packages/frontend/.env`)

```bash
# API
VITE_API_URL=http://localhost:8080

# Desarrollo
VITE_PORT=8082
VITE_APP_NAME=Banca en Línea
VITE_APP_VERSION=1.0.0
```

## 🐳 Instalación con Docker

### Opción 1: Instalación Completa (Recomendada)

```bash
# Construir y levantar todos los servicios
docker-compose up --build

# En segundo plano
docker-compose up --build -d
```

### Opción 2: Servicios Individuales

```bash
# Solo la base de datos
docker-compose up postgres

# Backend + Base de datos
docker-compose up postgres backend

# Sistema completo
docker-compose up postgres backend frontend
```

### Comandos Útiles de Docker

```bash
# Ver logs en tiempo real
docker-compose logs -f

# Ver logs de un servicio específico
docker-compose logs -f backend
docker-compose logs -f frontend
docker-compose logs -f postgres

# Reconstruir un servicio específico
docker-compose build backend
docker-compose build frontend

# Reiniciar servicios
docker-compose restart backend
docker-compose restart frontend

# Detener todos los servicios
docker-compose down

# Detener y eliminar volúmenes
docker-compose down -v

# Ver estado de los contenedores
docker-compose ps
```

## 🔐 Credenciales de Prueba

### Usuario Administrador
- **Email:** `admin@banco.com`
- **Contraseña:** `admin123`
- **Rol:** Administrador del sistema

### Usuario Cliente
- **Email:** `cliente@banco.com`
- **Contraseña:** `cliente123`
- **Rol:** Cliente bancario

### Base de Datos PostgreSQL
- **Host:** `localhost:5432`
- **Usuario:** `postgres`
- **Contraseña:** `postgres123`
- **Base de datos:** `banca_db`

## 🌐 URLs del Sistema

| Servicio | URL | Descripción |
|----------|-----|-------------|
| **Frontend** | http://localhost:8082 | Aplicación web principal |
| **Backend API** | http://localhost:8080 | API REST |
| **Health Check** | http://localhost:8080/health | Estado del backend |
| **API Docs** | http://localhost:8080/api/docs | Documentación de la API |

### Endpoints Principales

```bash
# Autenticación
POST /auth/login          # Iniciar sesión
POST /auth/register       # Registrar usuario
POST /auth/logout         # Cerrar sesión
GET  /auth/me            # Obtener usuario actual

# Cuentas
GET  /accounts           # Listar cuentas
POST /accounts           # Crear cuenta
GET  /accounts/:id       # Obtener cuenta específica

# Transacciones
GET  /transactions       # Listar transacciones
POST /transactions       # Crear transacción
GET  /transactions/:id   # Obtener transacción específica

# Usuarios (Admin)
GET  /users              # Listar usuarios
POST /users              # Crear usuario
PUT  /users/:id          # Actualizar usuario
DELETE /users/:id        # Eliminar usuario
```

## 🧪 Funcionalidades

### ✅ Implementadas

- **Autenticación JWT** con login/logout
- **Registro de usuarios** con validación
- **Dashboard principal** con resumen de cuentas
- **Gestión de cuentas bancarias**
- **Historial de transacciones**
- **Interfaz responsive** con Tailwind CSS
- **Validación de formularios**
- **Manejo de errores** centralizado

### 🚧 En Desarrollo

- **Transferencias entre cuentas**
- **Integración completa con TigerBeetle**
- **Reportes financieros**
- **Notificaciones en tiempo real**
- **Panel de administración**

## 🔧 Desarrollo Local

### Sin Docker (Desarrollo Nativo)

#### Backend (Go)

```bash
cd packages/backend

# Instalar dependencias
go mod download

# Configurar variables de entorno
cp .env.example .env

# Ejecutar migraciones (requiere PostgreSQL corriendo)
go run main.go migrate

# Iniciar servidor de desarrollo
go run main.go
```

#### Frontend (React)

```bash
cd packages/frontend

# Instalar dependencias
npm install
# o
pnpm install

# Configurar variables de entorno
cp .env.example .env

# Iniciar servidor de desarrollo
npm run dev
# o
pnpm dev
```

### Herramientas de Desarrollo

```bash
# Linting del frontend
cd packages/frontend
npm run lint

# Build de producción
npm run build

# Tests del backend
cd packages/backend
go test ./...

# Formatear código Go
go fmt ./...
```

## 📊 Base de Datos

### Migraciones

Las migraciones se ejecutan automáticamente al iniciar el backend. Los archivos están en `packages/backend/migrations/`.

### Esquema Principal

```sql
-- Usuarios
users (id, email, password_hash, first_name, last_name, role, created_at, updated_at)

-- Cuentas bancarias
accounts (id, user_id, account_number, account_type, balance, currency, created_at, updated_at)

-- Transacciones
transactions (id, from_account_id, to_account_id, amount, transaction_type, description, status, created_at)
```

### Conexión Manual a la Base de Datos

```bash
# Usando Docker
docker-compose exec postgres psql -U postgres -d banca_db

# Usando cliente local
psql -h localhost -p 5432 -U postgres -d banca_db
```

## 🚨 Solución de Problemas

### Problema: Puerto ya en uso

```bash
# Verificar qué proceso usa el puerto
netstat -ano | findstr :8082
netstat -ano | findstr :8080

# Matar proceso específico (Windows)
taskkill /PID <PID> /F

# Cambiar puerto en docker-compose.yml
ports:
  - "8083:8082"  # Cambiar puerto externo
```

### Problema: Error de conexión a la base de datos

```bash
# Verificar que PostgreSQL esté corriendo
docker-compose ps

# Reiniciar solo la base de datos
docker-compose restart postgres

# Ver logs de PostgreSQL
docker-compose logs postgres
```

### Problema: Frontend no se conecta al backend

1. Verificar que `VITE_API_URL` esté configurado correctamente
2. Verificar que el backend esté corriendo en el puerto correcto
3. Revisar configuración de CORS en el backend

### Problema: Error de autenticación

1. Verificar que `JWT_SECRET` sea el mismo en backend y frontend
2. Limpiar localStorage del navegador
3. Verificar que las credenciales sean correctas

## 📝 Logs y Debugging

### Ver Logs en Tiempo Real

```bash
# Todos los servicios
docker-compose logs -f

# Solo backend
docker-compose logs -f backend

# Solo frontend
docker-compose logs -f frontend

# Solo base de datos
docker-compose logs -f postgres
```

### Logs de Debugging

El sistema incluye logs detallados para debugging:

#### Frontend (Consola del navegador)
- `🔐 AuthContext:` Logs de autenticación
- `🌐 API:` Logs de peticiones HTTP
- `❌ Errores:` Logs de errores

#### Backend (Terminal)
- Logs de peticiones HTTP con método, URL y tiempo de respuesta
- Logs de conexión a base de datos
- Logs de errores con stack trace

### Debugging en Desarrollo

```bash
# Habilitar logs detallados en el backend
export LOG_LEVEL=debug

# Habilitar debugging en el frontend
export VITE_DEBUG=true
```

## 🤝 Contribución

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/nueva-funcionalidad`)
3. Commit tus cambios (`git commit -am 'Agregar nueva funcionalidad'`)
4. Push a la rama (`git push origin feature/nueva-funcionalidad`)
5. Abre un Pull Request

## 📄 Licencia

Este proyecto está bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

---

## 🆘 Soporte

Si tienes problemas o preguntas:

1. Revisa la sección de [Solución de Problemas](#-solución-de-problemas)
2. Verifica los logs con `docker-compose logs -f`
3. Abre un issue en el repositorio

**¡Feliz desarrollo! 🚀**