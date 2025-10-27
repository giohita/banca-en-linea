# Desarrollo del Backend

## Limitaciones en Windows

### TigerBeetle
TigerBeetle no es compatible con desarrollo local en Windows debido a restricciones de build constraints. Los archivos Go de TigerBeetle son excluidos automáticamente en sistemas Windows.

**Soluciones:**
1. **Docker (Recomendado)**: Usar `docker-compose up --build backend` para desarrollo
2. **WSL2**: Desarrollar dentro de Windows Subsystem for Linux
3. **Linux/macOS**: Desarrollo nativo en sistemas Unix

### Tests Unitarios
Los tests que dependen de TigerBeetle no pueden ejecutarse localmente en Windows. Usar Docker o CI/CD para ejecutar tests completos.

## Desarrollo con Docker

```bash
# Iniciar el backend con todas las dependencias
docker-compose up --build backend

# El servidor estará disponible en:
# http://localhost:8080
```

## Endpoints Disponibles

- `GET /health` - Health check
- `POST /users` - Crear usuario
- `GET /users/:id` - Obtener usuario
- `GET /users` - Listar usuarios
- `POST /users/:id/deposit` - Depositar dinero
- `POST /users/:id/withdraw` - Retirar dinero
- `POST /users/:id/transfer` - Transferir dinero

## Estructura del Proyecto

```
packages/backend/
├── main.go                 # Punto de entrada
├── models/                 # Modelos de datos
├── internal/
│   ├── db/                # Repositorios y servicios
│   └── tigerbeetle/       # Servicios de TigerBeetle
├── database/              # Configuración y migraciones
├── migrations/            # Migraciones SQL
└── tests/                 # Tests unitarios
```