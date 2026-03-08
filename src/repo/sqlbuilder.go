package repo

import (
	"fmt"
	"strconv"
	"strings"
)

const sqlCurrentTimestamp = "CURRENT_TIMESTAMP"

func buildTransactionScript(statements ...string) string {
	var builder strings.Builder

	builder.WriteString("BEGIN;\n")
	for _, statement := range statements {
		trimmed := strings.TrimSpace(statement)
		if trimmed == "" {
			continue
		}

		builder.WriteString(trimmed)
		if !strings.HasSuffix(trimmed, "\n") {
			builder.WriteString("\n")
		}
	}
	builder.WriteString("COMMIT;\n")

	return builder.String()
}

func buildInsertStatement(table string, columns []string, values []string) string {
	if len(columns) != len(values) {
		panic(fmt.Sprintf("insert statement for %s has %d columns and %d values", table, len(columns), len(values)))
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s);",
		table,
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
	)
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func sqlInt(value int) string {
	return strconv.Itoa(value)
}

func sqlInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func sqlBool(value bool) string {
	if value {
		return "1"
	}

	return "0"
}

func sqlStringList(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, sqlString(value))
	}

	return strings.Join(quoted, ", ")
}
