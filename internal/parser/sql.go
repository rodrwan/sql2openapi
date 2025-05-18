package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Column representa una columna SQL.
type Column struct {
	Name       string
	Type       string
	Primary    bool
	ForeignKey *ForeignKey // nil si no es FK
	Required   bool        // nil si no es required
}

type ForeignKey struct {
	RefTable string
	RefCol   string
}

// Table representa una tabla SQL.
type Table struct {
	Name    string
	Columns []Column
}

// ParseSQLFile lee y parsea un archivo .sql, devolviendo una lista de tablas.
func ParseSQLFile(path string) ([]Table, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("no se pudo leer el archivo: %w", err)
	}

	return ParseSQL(string(content))
}

// ParseSQL analiza contenido SQL y devuelve una lista de tablas.
func ParseSQL(sql string) ([]Table, error) {
	createTableRegex := regexp.MustCompile(`(?is)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)\s*\((.*?)\);`)
	matches := createTableRegex.FindAllStringSubmatch(sql, -1)
	fkRegex := regexp.MustCompile(`(?i)FOREIGN\s+KEY\s*\((\w+)\)\s+REFERENCES\s+(\w+)\s*\((\w+)\)`)

	var tables []Table

	for _, match := range matches {
		tableName := match[1]
		columnsRaw := match[2]
		lines := strings.Split(columnsRaw, ",")

		var columns []Column

		for _, line := range lines {
			line = strings.TrimSpace(line)

			// Detectar FOREIGN KEY
			if fkMatch := fkRegex.FindStringSubmatch(line); fkMatch != nil {
				colName := fkMatch[1]
				refTable := fkMatch[2]
				refCol := fkMatch[3]

				// Buscar columna y asignar ForeignKey
				for i, col := range columns {
					if col.Name == colName {
						columns[i].ForeignKey = &ForeignKey{
							RefTable: refTable,
							RefCol:   refCol,
						}
					}
				}
				continue
			}

			// Evitar línea FOREIGN KEY doble procesamiento
			if strings.HasPrefix(strings.ToUpper(line), "FOREIGN KEY") {
				continue
			}

			// Columnas normales (igual que antes)
			colParts := strings.Fields(line)
			if len(colParts) < 2 {
				continue
			}

			colName := colParts[0]
			colType := colParts[1]
			isPrimary := strings.Contains(strings.ToUpper(line), "PRIMARY KEY")
			isRequired := strings.Contains(strings.ToUpper(line), "NOT NULL")

			columns = append(columns, Column{
				Name:     colName,
				Type:     colType,
				Primary:  isPrimary,
				Required: isRequired,
			})
		}

		tables = append(tables, Table{
			Name:    tableName,
			Columns: columns,
		})
	}

	return tables, nil
}
