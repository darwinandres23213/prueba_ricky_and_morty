# API Gateway Rick and Morty

Este proyecto implementa un API Gateway para la API de Rick and Morty con autenticación JWT y límite de uso de tokens. El sistema está construido usando una arquitectura de microservicios con Docker.

## Arquitectura

El sistema está compuesto por tres microservicios:

1. **Auth Service** (Puerto 8081)
   - Maneja la autenticación de usuarios
   - Implementa registro y login
   - Gestiona tokens JWT con límite de uso
   - Usa SQLite para almacenamiento de usuarios

2. **Gateway Service** (Puerto 8080)
   - Proxy inverso para las peticiones a Rick and Morty
   - Valida tokens JWT
   - Implementa rate limiting por token
   - Maneja CORS

3. **Rick and Morty Service** (Puerto 8082)
   - Proxy a la API pública de Rick and Morty
   - Cachea respuestas
   - Maneja errores de la API externa

## ¿Por qué no usamos archivos .env?

Aunque el proyecto usa la librería `godotenv`, no utilizamos archivos `.env` locales porque:

1. **Dockerización completa**: Todas las variables de entorno están definidas en el `docker-compose.yml`
2. **Portabilidad**: Los contenedores son autónomos y no dependen de archivos locales
3. **Seguridad**: Las variables sensibles (como JWT_SECRET) están definidas en el entorno de Docker
4. **Consistencia**: Asegura que todos los entornos (desarrollo, producción) usen las mismas configuraciones

Las variables de entorno se definen en el `docker-compose.yml`:
```yaml
services:
  auth:
    environment:
      - JWT_SECRET=supersecret
      - AUTH_SERVICE_PORT=8081
      - COOKIE_NAME=auth_token
  gateway:
    environment:
      - GATEWAY_SERVICE_PORT=8080
      - AUTH_SERVICE_PORT=8081
      - RICKMORTY_SERVICE_PORT=8082
  rickmorty:
    environment:
      - RICKMORTY_SERVICE_PORT=8082
```

## Flujo de la Aplicación

1. **Registro de Usuario**
   ```bash
   POST http://localhost:8081/api/v1/register
   Body: {
       "username": "usuario1",
       "password": "contraseña123"
   }
   ```

2. **Login y Obtención de Token**
   ```bash
   POST http://localhost:8081/api/v1/login
   Body: {
       "username": "usuario1",
       "password": "contraseña123"
   }
   ```
   - El token se guarda en una cookie HTTP-only
   - Cada token tiene 5 usos disponibles
   - El token expira después del quinto uso

3. **Uso de la API**
   - Todas las peticiones al Gateway requieren el token en cookie
   - El Gateway valida el token con el Auth Service
   - Si el token es válido, la petición se reenvía al Rick and Morty Service
   - El Rick and Morty Service obtiene los datos de la API pública

## Endpoints Disponibles

### 1. Servicio de Autenticación (8081)

#### Registro
```
POST http://localhost:8081/api/v1/register
Body: {
    "username": string,
    "password": string
}
```

#### Login
```
POST http://localhost:8081/api/v1/login
Body: {
    "username": string,
    "password": string
}
```

#### Validar Token
```
GET http://localhost:8080/api/v1/validate
(Requiere cookie con token JWT)
```

### 2. Servicio Gateway (8080)
Todos los endpoints requieren autenticación (cookie con token JWT)

#### Personajes
```
GET http://localhost:8080/api/v1/character
GET http://localhost:8080/api/v1/characters
GET http://localhost:8080/api/v1/character/{id}
GET http://localhost:8080/api/v1/character?page=2
```

#### Ubicaciones
```
GET http://localhost:8080/api/v1/location
GET http://localhost:8080/api/v1/locations
GET http://localhost:8080/api/v1/location/{id}
GET http://localhost:8080/api/v1/location?page=2
```

#### Episodios
```
GET http://localhost:8080/api/v1/episode
GET http://localhost:8080/api/v1/episodes
GET http://localhost:8080/api/v1/episode/{id}
GET http://localhost:8080/api/v1/episode?page=2
```

### 3. Servicio Rick and Morty (8082)
Estos endpoints NO requieren autenticación

#### Personajes
```
GET http://localhost:8080/api/v1/character
GET http://localhost:8080/api/v1/characters
GET http://localhost:8080/api/v1/character/{id}
GET http://localhost:8080/api/v1/character?page=2
```

#### Ubicaciones
```
GET http://localhost:8080/api/v1/location
GET http://localhost:8080/api/v1/locations
GET http://localhost:8080/api/v1/location/{id}
GET http://localhost:8080/api/v1/location?page=2
```

#### Episodios
```
GET http://localhost:8080/api/v1/episode
GET http://localhost:8080/api/v1/episodes
GET http://localhost:8080/api/v1/episode/{id}
GET http://localhost:8080/api/v1/episode?page=2
```

## Formato de Respuestas

Todas las respuestas siguen este formato JSON:
```json
{
    "status": "success" | "error",
    "message": "mensaje descriptivo",
    "data": { ... } // datos opcionales
}
```

### Ejemplos de Respuestas

#### Login Exitoso
```json
{
    "status": "success",
    "message": "Login exitoso",
    "data": {
        "username": "usuario1",
        "message": "Token guardado en cookie"
    }
}
```

#### Token Válido
```json
{
    "status": "success",
    "message": "Token válido",
    "data": {
        "usos_restantes": 4,
        "username": "usuario1",
        "message": "Token expirará después de este uso"
    }
}
```

#### Token Expirado
```json
{
    "status": "error",
    "message": "Token expirado por uso máximo alcanzado"
}
```

## Gestión de Tokens

1. **Creación**
   - Se crea al hacer login exitoso
   - Se guarda en una cookie HTTP-only
   - Tiene 5 usos disponibles

2. **Validación**
   - Se valida en cada petición al Gateway
   - El contador de usos se decrementa
   - En el último uso, se avisa que expirará

3. **Expiración**
   - Después del quinto uso, el token se elimina
   - La cookie se elimina automáticamente
   - Se requiere nuevo login para obtener otro token

## Levantar el Proyecto

1. **Requisitos**
   - Docker
   - Docker Compose

2. **Construir y Levantar Contenedores**
   ```bash
   # Detener contenedores existentes
   docker-compose down

   # Construir y levantar contenedores
   docker-compose up --build
   ```

3. **Verificar Servicios**
   - Auth Service: http://localhost:8081
   - Gateway: http://localhost:8080
   - Rick and Morty Service: http://localhost:8082

## Ejemplo de Uso

1. **Registrar Usuario**
   ```bash
   curl -X POST http://localhost:8081/api/v1/register \
     -H "Content-Type: application/json" \
     -d '{"username": "usuario1", "password": "contraseña123"}'
   ```

2. **Login**
   ```bash
   curl -X POST http://localhost:8081/api/v1/login \
     -H "Content-Type: application/json" \
     -d '{"username": "usuario1", "password": "contraseña123"}'
   ```

3. **Obtener Personajes**
   ```bash
   curl http://localhost:8080/api/v1/characters
   ```

## Características de Seguridad

1. **Autenticación**
   - Tokens JWT
   - Cookies HTTP-only
   - Límite de uso por token

2. **CORS**
   - Configurado para permitir peticiones de cualquier origen
   - Métodos permitidos: GET, OPTIONS
   - Headers permitidos: Content-Type

3. **Rate Limiting**
   - 5 usos por token
   - Expiración automática
   - Eliminación de cookie al expirar

## Estructura del Proyecto

```
.
├── cmd/
│   ├── auth/          # Servicio de autenticación
│   ├── gateway/       # Servicio gateway
│   └── rickmorty/     # Servicio Rick and Morty
├── internal/
│   ├── auth/          # Lógica de autenticación
│   ├── gateway/       # Lógica del gateway
│   └── rickmorty/     # Lógica de Rick and Morty
├── Dockerfile.auth
├── Dockerfile.gateway
├── Dockerfile.rickmorty
├── docker-compose.yml
├── go.mod
└── go.sum
```

## Tecnologías Utilizadas

- Go 1.21
- Docker & Docker Compose
- JWT para autenticación
- SQLite para base de datos
- Gorilla Mux para routing
- CORS para manejo de CORS 