package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rodrwan/sql2openapi/internal/config"
	"github.com/rodrwan/sql2openapi/internal/generator"
	"github.com/rodrwan/sql2openapi/internal/parser"
)

func main() {
	// Flags
	inputPath := flag.String("i", "", "Ruta del archivo .sql de entrada")
	outputPath := flag.String("o", "openapi.yaml", "Ruta del archivo de salida OpenAPI")
	format := flag.String("f", "yaml", "Formato de salida: yaml (por defecto) o json")
	configPath := flag.String("c", "", "Ruta al archivo de configuración YAML")

	flag.Parse()

	if *inputPath == "" {
		fmt.Println("❌ Debes proporcionar un archivo SQL de entrada con -i")
		flag.Usage()
		os.Exit(1)
	}

	tables, err := parser.ParseSQLFile(*inputPath)
	if err != nil {
		log.Fatal(err)
	}

	var cfg *config.Config
	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			log.Fatalf("Error leyendo config: %v", err)
		}
	}

	// Generar documento OpenAPI
	var output []byte
	switch *format {
	case "yaml":
		output, err = generator.GenerateOpenAPI(tables, cfg)
	default:
		log.Fatalf("Formato no soportado: %s", *format)
	}

	if err != nil {
		log.Fatalf("Error al generar la especificación OpenAPI: %v", err)
	}

	// Crear carpeta si no existe
	if err := os.MkdirAll(filepath.Dir(*outputPath), os.ModePerm); err != nil {
		log.Fatalf("Error al crear carpeta de salida: %v", err)
	}

	// Escribir archivo
	if err := os.WriteFile(*outputPath, output, 0644); err != nil {
		log.Fatalf("Error al escribir archivo OpenAPI: %v", err)
	}

	fmt.Printf("✅ SQL to OpenAPI generated: %s\n", *outputPath)

	// // ejecutar el binario de go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	// exec.Command("go", "install", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest")

	// // ejecutar el binario de oapi-codegen -config=config.yaml output.yaml
	// exec.Command("oapi-codegen", "-config=config.yaml", *outputPath)

	// fmt.Printf("✅ Code generated from %s\n", *outputPath)
}
