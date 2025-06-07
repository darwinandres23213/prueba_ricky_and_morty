# API Gateway Rick and Morty

Este proyecto implementa un API Gateway para la API de Rick and Morty con autenticación JWT y límite de uso de tokens. El sistema está construido usando una arquitectura de microservicios con Docker.

## 🚀 Instalación y Ejecución

### Requisitos Previos
- Docker
- Docker Compose
- Postman o similar (recomendado para probar la API)

### Pasos de Instalación

1. **Clonar el Repositorio**
   ```bash
   git clone https://github.com/yourusername/api_ricky_and_morty.git
   cd api_ricky_and_morty
   ```

2. **Construir y Levantar Contenedores**
   ```bash
   # Detener contenedores existentes (si los hay)
   docker-compose down

   # Construir y levantar contenedores
   docker-compose up --build
   ```

3. **Verificar Servicios**
   Los siguientes servicios estarán disponibles en:
   - Auth Service: http://localhost:8081
   - Gateway: http://localhost:8080
   - Rick and Morty Service: http://localhost:8082 (no accesible directamente)

## 🛠️ Probando la API

**IMPORTANTE**: Se recomienda usar Postman o una herramienta similar para probar la API, ya que:
- Maneja automáticamente las cookies HTTP-only
- Permite guardar y reutilizar tokens
- Facilita el seguimiento de las peticiones
- Proporciona una interfaz amigable para probar los endpoints

### Configuración en Postman

1. **Crear una nueva colección** para la API con la URL base:
   - Auth Service: `http://localhost:8081`
   - Gateway: `http://localhost:8080`

2. **Configurar el manejo de cookies**:
   - Settings > General > Automatically follow redirects
   - Settings > General > Send cookies

### Flujo de Prueba en Postman

1. **Registro de Usuario**
   ```
   POST http://localhost:8081/api/v1/register
   Headers:
   Content-Type: application/json

   Body (JSON):
   {
       "username": "usuario1",
       "password": "contraseña123"
   }
   ```

2. **Login**
   ```
   POST http://localhost:8081/api/v1/login
   Headers:
   Content-Type: application/json

   Body (JSON):
   {
       "username": "usuario1",
       "password": "contraseña123"
   }
   ```
   - Postman guardará automáticamente la cookie con el token en `http://localhost:8081`

3. **Probar Endpoints Protegidos**
   Todas las peticiones al Gateway usarán automáticamente la cookie guardada:
   ```
   GET http://localhost:8080/api/v1/characters
   GET http://localhost:8080/api/v1/character/1
   GET http://localhost:8080/api/v1/locations
   GET http://localhost:8080/api/v1/episodes
   ```

## 📡 Endpoints Disponibles

### 1. Servicio de Autenticación (http://localhost:8081)

#### Registro
```
POST http://localhost:8081/api/v1/register
Headers:
Content-Type: application/json

Body:
{
    "username": string,
    "password": string
}
```

#### Login
```
POST http://localhost:8081/api/v1/login
Headers:
Content-Type: application/json

Body:
{
    "username": string,
    "password": string
}
```

#### Validar Token
```
GET http://localhost:8081/api/v1/validate
Headers:
Cookie: auth_token=<token>
```

### 2. Servicio Gateway (http://localhost:8080) - Único punto de acceso público

Todos los endpoints requieren autenticación (cookie con token JWT)

#### Personajes
```
GET http://localhost:8080/api/v1/characters
GET http://localhost:8080/api/v1/character
GET http://localhost:8080/api/v1/character/{id}
GET http://localhost:8080/api/v1/character?page=2

Headers:
Cookie: auth_token=<token>
```

#### Ubicaciones
```
GET http://localhost:8080/api/v1/locations
GET http://localhost:8080/api/v1/location
GET http://localhost:8080/api/v1/location/{id}
GET http://localhost:8080/api/v1/location?page=2

Headers:
Cookie: auth_token=<token>
```

#### Episodios
```
GET http://localhost:8080/api/v1/episodes
GET http://localhost:8080/api/v1/episode
GET http://localhost:8080/api/v1/episode/{id}
GET http://localhost:8080/api/v1/episode?page=2

Headers:
Cookie: auth_token=<token>
```

### Ejemplos de Uso con curl

1. **Registro**
   ```bash
   curl -X POST http://localhost:8081/api/v1/register \
     -H "Content-Type: application/json" \
     -d '{"username": "usuario1", "password": "contraseña123"}'
   ```

2. **Login**
   ```bash
   curl -X POST http://localhost:8081/api/v1/login \
     -H "Content-Type: application/json" \
     -d '{"username": "usuario1", "password": "contraseña123"}' \
     -c cookies.txt
   ```

3. **Obtener Personajes**
   ```bash
   curl http://localhost:8080/api/v1/characters \
     -b cookies.txt
   ```

## 🔒 Gestión de Tokens

1. **Creación** (http://localhost:8081)
   - Se crea al hacer login exitoso en `/api/v1/login`
   - Se guarda en una cookie HTTP-only para `localhost:8081`
   - Tiene 5 usos disponibles

2. **Validación** (http://localhost:8080)
   - Se valida en cada petición al Gateway
   - El contador de usos se decrementa
   - En el último uso, se avisa que expirará

3. **Expiración**
   - Después del quinto uso, el token se elimina
   - La cookie se elimina automáticamente
   - Se requiere nuevo login en `http://localhost:8081/api/v1/login`

## 📝 Formato de Respuestas

Todas las respuestas siguen este formato JSON:
```json
{
    "status": "success" | "error",
    "message": "mensaje descriptivo",
    "data": { ... } // datos opcionales
}
```

### Ejemplos de Respuestas

#### Login Exitoso (http://localhost:8081/api/v1/login)
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

#### Token Válido (http://localhost:8080/api/v1/characters)
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

## ⚠️ IMPORTANTE: Acceso a la API

**El servicio Rick and Morty (puerto 8082) NO es accesible directamente**. Todas las peticiones DEBEN pasar por el Gateway (puerto 8080).

Si intentas acceder directamente a http://localhost:8082, recibirás:
```json
{
    "status": "error",
    "message": "Acceso directo no permitido. Use el Gateway en el puerto 8080"
}
```

### Flujo Correcto de Peticiones
1. Todas las peticiones deben ir a http://localhost:8080
2. El Gateway valida el token
3. Si el token es válido, el Gateway reenvía la petición al servicio Rick and Morty
4. El servicio Rick and Morty solo acepta peticiones que vengan del Gateway

### Flujo Incorrecto (No Permitido)
❌ http://localhost:8082/api/v1/characters
❌ http://localhost:8082/api/v1/locations
❌ http://localhost:8082/api/v1/episodes

### Flujo Correcto (Recomendado)
✅ http://localhost:8080/api/v1/characters
✅ http://localhost:8080/api/v1/locations
✅ http://localhost:8080/api/v1/episodes

## 🏗️ Arquitectura

El sistema está compuesto por tres microservicios:

1. **Auth Service** (http://localhost:8081)
   - Maneja la autenticación de usuarios
   - Implementa registro y login
   - Gestiona tokens JWT con límite de uso
   - Usa SQLite para almacenamiento de usuarios

2. **Gateway Service** (http://localhost:8080) - **Único punto de entrada público**
   - Proxy inverso para las peticiones a Rick and Morty
   - Valida tokens JWT
   - Implementa rate limiting por token
   - Maneja CORS
   - **Es el único punto de acceso permitido para los usuarios**

3. **Rick and Morty Service** (http://localhost:8082) - **No accesible directamente**
   - Proxy a la API pública de Rick and Morty
   - Cachea respuestas
   - Maneja errores de la API externa
   - **Solo acepta peticiones que vengan del Gateway**
   - Implementa validación de origen de peticiones
   - Rechaza cualquier petición que no venga del Gateway

### Diagrama de Flujo
```
Cliente -> Gateway (8080) -> Rick and Morty (8082)
   ↑          ↓
   └── Auth (8081) <──┘
```

### Seguridad
- El servicio Rick and Morty (8082) está completamente aislado
- Solo el Gateway puede comunicarse con el servicio Rick and Morty
- Todas las peticiones son validadas por el Gateway
- El servicio Rick and Morty verifica que las peticiones vengan del Gateway
- No hay forma de acceder directamente al servicio Rick and Morty

## 🔐 ¿Por qué no usamos archivos .env?

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

## 📁 Estructura del Proyecto

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

## 🛠️ Tecnologías Utilizadas

- Go 1.21
- Docker & Docker Compose
- JWT para autenticación
- SQLite para base de datos
- Gorilla Mux para routing
- CORS para manejo de CORS
- Postman (recomendado) para pruebas 