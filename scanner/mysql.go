package scanner

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nicklanng/modelgen/model"
)

type mySQLExplain struct {
	Field   *string
	Type    *string
	Null    *string
	Key     *string
	Default *string
	Extra   *string
}

var mysqlDataTypes = map[string]model.SQLType{
	"char":    model.TypeString,
	"varchar": model.TypeString,

	"tinytext":   model.TypeString,
	"text":       model.TypeString,
	"mediumtext": model.TypeString,
	"longtext":   model.TypeString,

	"json": model.TypeJSON,

	"enum": model.TypeString,
	"set":  model.TypeString,

	"bit":        model.TypeByteSlice,
	"binary":     model.TypeByteSlice,
	"varbinary":  model.TypeByteSlice,
	"tinyblob":   model.TypeByteSlice,
	"mediumblob": model.TypeByteSlice,
	"blob":       model.TypeByteSlice,
	"longblob":   model.TypeByteSlice,

	"tinyint_as_bool": model.TypeBool,
	"tinyint":         model.TypeInt,
	"smallint":        model.TypeInt,
	"mediumint":       model.TypeInt,
	"int":             model.TypeInt,
	"bigint":          model.TypeInt,

	"float":   model.TypeFloat,
	"double":  model.TypeFloat,
	"decimal": model.TypeFloat,

	"date":      model.TypeDate,
	"datetime":  model.TypeDate,
	"timestamp": model.TypeDate,
	"time":      model.TypeString,
	"year":      model.TypeInt,
}

type MySQL struct {
	username   string
	password   string
	host       string
	port       string
	database   string
	connection *sql.DB
}

func NewMySQL(username, password, host, port, database string) *MySQL {
	return &MySQL{
		username: username,
		password: password,
		host:     host,
		port:     port,
		database: database,
	}
}

func (m *MySQL) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", m.username, m.password, m.host, m.port, m.database)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	// check for a valid connection
	if err := conn.Ping(); err != nil {
		return err
	}

	m.connection = conn

	return nil
}

func (m *MySQL) QueryStructure() ([]model.EntityDescriptor, error) {
	tables, err := m.queryTables()
	if err != nil {
		return nil, err
	}

	if len(tables) == 0 {
		return nil, errors.New("no tables to read")
	}

	var structs []model.EntityDescriptor
	for tableName := range tables {
		explanations, err := m.explainTable(tableName)
		if err != nil {
			return nil, err
		}

		s, err := m.parseExplanation(tableName, explanations)
		if err != nil {
			return nil, err
		}

		structs = append(structs, s)
	}

	return structs, nil
}

func (m *MySQL) queryTables() (map[string]string, error) {
	tables := make(map[string]string)

	stmt := `SELECT table_name, column_comment
			 FROM information_schema.columns AS c
			 WHERE c.column_key = "PRI"
			 AND c.table_schema = ?
		     AND column_name = "id"`

	rows, err := m.connection.Query(stmt, m.database)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var name string
		var comment string
		if err := rows.Scan(&name, &comment); err != nil {
			return nil, err
		}
		tables[name] = comment
	}

	return tables, nil
}

func (m *MySQL) explainTable(table string) ([]mySQLExplain, error) {
	var tableExplanations []mySQLExplain

	rows, err := m.connection.Query("EXPLAIN " + table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var explanation mySQLExplain
		if err := rows.Scan(&explanation.Field, &explanation.Type, &explanation.Null, &explanation.Key, &explanation.Default, &explanation.Extra); err != nil {
			return nil, err
		}
		tableExplanations = append(tableExplanations, explanation)
	}

	return tableExplanations, nil
}

func (m *MySQL) parseExplanation(table string, explanations []mySQLExplain) (model.EntityDescriptor, error) {
	t := model.EntityDescriptor{
		Name:      toPascalCase(table),
		TableName: table,
		Imports:   make(map[string]struct{}),
	}

	for _, expl := range explanations {
		typ, err := mapType(*expl.Type, *expl.Null, mysqlDataTypes)
		if err != nil {
			return model.EntityDescriptor{}, err
		}

		f := model.Field{
			Name:       toPascalCase(*expl.Field),
			Type:       typ,
			ColumnName: strings.ToLower(*expl.Field),
			Nullable:   *expl.Null == "YES",
		}

		t.Fields = append(t.Fields, f)
		if imp, ok := needsImport(f.Type); ok {
			t.Imports[imp] = struct{}{}
		}
	}

	return t, nil
}
