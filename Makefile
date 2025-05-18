APP_NAME := sql2openapi
CMD_DIR := ./cmd/$(APP_NAME)

.PHONY: all build install go-install clean example

## Compilar el binario localmente en ./bin
build:
	@echo "🔧 Compilando $(APP_NAME)..."
	@mkdir -p ./bin
	go build -o ./bin/$(APP_NAME) $(CMD_DIR)
	@echo "✅ Binario generado en ./bin/$(APP_NAME)"

## Instalar en /usr/local/bin usando cp (requiere sudo)
install: build
	@echo "📦 Instalando en /usr/local/bin/$(APP_NAME)..."
	@sudo cp ./bin/$(APP_NAME) /usr/local/bin/$(APP_NAME)
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@echo "✅ Instalado correctamente."

## Instalar usando `go install` directamente al $GOBIN
go-install:
	@echo "🚀 Instalando con go install..."
	go install $(CMD_DIR)
	@echo "✅ Instalado correctamente en $$GOBIN o $$GOPATH/bin."

## Eliminar binarios generados
clean:
	@echo "🧹 Limpiando..."
	@rm -rf ./bin

example:
	./bin/sql2openapi -i ./example/schema.sql -o ./example/openapi.yaml -c ./example/sql2openapi.yaml