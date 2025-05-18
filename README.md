# sql2openapi

Transforma archivos `.sql` con definiciones de tablas en una especificación **OpenAPI 3.0**, generando automáticamente endpoints CRUD y esquemas de datos a partir de la estructura de tus tablas.

Ideal para construir documentación de APIs basada en bases de datos existentes.

---

## ✨ Características

- ✅ Soporte para múltiples tablas con claves primarias
- 🔄 Generación automática de rutas CRUD (`GET`, `POST`, `PUT`, `DELETE`)
- 🔐 Soporte para autenticación vía Bearer Token
- 🔁 Rutas anidadas basadas en claves foráneas (`GET /users/{id}/posts`)
- ⚙️ Configuración extensible vía archivo `config.yaml`
- 📦 Instalación fácil con `go install` o `make`

---

## 🚀 Instalación

### Opción 1: Go (recomendado)

```bash
go install ./cmd/sql2openapi
```

### Opción 2: Makefile

```bash
make go-install
```

> Asegúrate de tener `$GOBIN` o `$GOPATH/bin` en tu `$PATH`.

---

## 🛠 Uso básico

```bash
sql2openapi -i schema.sql -o openapi.yaml
```

- `-i`: ruta al archivo SQL con sentencias `CREATE TABLE`
- `-o`: ruta de salida del archivo `.yaml` generado (por defecto: `openapi.yaml`)

---

## ⚙️ Configuración avanzada (opcional)

Puedes usar un archivo `config.yaml` para definir seguridad y endpoints personalizados:

### config.yaml

```yaml
security:
  scheme: bearer
  format: JWT
  protected:
    - path: /users
      methods: [post, put, delete]

customEndpoints:
  - path: /login
    method: post
    summary: User login
    requestSchema:
      type: object
      properties:
        email:
          type: string
        password:
          type: string
      required: [email, password]
    responseSchema:
      type: object
      properties:
        token:
          type: string
```

### Ejecutar con configuración:

```bash
sql2openapi -i schema.sql -o openapi.yaml -c config.yaml
```

---

## 📂 Estructura del proyecto

```bash
sql2openapi/
├── cmd/                # CLI principal
├── internal/
│   ├── parser/         # Parsing del SQL
│   ├── generator/      # Generación del OpenAPI
│   └── config/         # Lógica para leer config.yaml
├── bin/                # Binarios compilados
├── example/
│   ├── schema.sql          # Archivo de ejemplo de entrada
│   └── sql2openapi.yaml
├── config.yaml         # Archivo opcional de configuración
└── README.md
```

---

## 🔭 Roadmap

- [ ] Soporte para tipos compuestos y columnas JSON
- [ ] Tags por tabla en OpenAPI
- [ ] Exportación en formato JSON
- [ ] Validación de schemas antes de escribir
- [ ] Swagger UI preview integrado (`--serve`)

---

## 🧑‍💻 Autor

Desarrollado por Rodrigo Fuenzalida
¡Con cariño, en Go 🦫!

