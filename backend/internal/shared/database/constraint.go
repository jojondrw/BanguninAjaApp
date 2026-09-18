package database

import "fmt"

const (
	DeleteRestrict = "RESTRICT"
	DeleteCascade  = "CASCADE"
	DeleteSetNull  = "SET NULL"
)

func ForeignKey(table, column, referencedTable, onDelete string) string {
	name := fmt.Sprintf("fk_%s_%s", table, column)
	return guarded(name, fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (id) ON UPDATE CASCADE ON DELETE %s",
		table, name, column, referencedTable, onDelete,
	))
}

func Check(table, name, expression string) string {
	full := fmt.Sprintf("chk_%s_%s", table, name)
	return guarded(full, fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s CHECK (%s)",
		table, full, expression,
	))
}

func Unique(table, name string, columns string) string {
	full := fmt.Sprintf("uq_%s_%s", table, name)
	return guarded(full, fmt.Sprintf(
		"ALTER TABLE %s ADD CONSTRAINT %s UNIQUE (%s)",
		table, full, columns,
	))
}

func guarded(name, statement string) string {
	return fmt.Sprintf(`DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = '%s') THEN
    %s;
  END IF;
END $$`, name, statement)
}
