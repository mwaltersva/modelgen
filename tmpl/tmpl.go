package tmpl

import (
	"fmt"
	"html/template"
	"strings"
)

var FuncMap = template.FuncMap{
	"insert_fields": GetInsertFields,
	"insert_values": GetInsertValues,
	"insert_args":   GetInsertArgs,
	"scan_fields":   GetScanFields,
	"update_args":   GetUpdateArgs,
	"update_values": GetUpdateValues,
	"upsert_values": GetUpsertValues,
}

func GetInsertFields(fields []TmplField) string {
	var parts []string
	for _, fl := range fields {
		if fl.ColumnName == "id" {
			continue
		}
		parts = append(parts, "`"+fl.ColumnName+"`")
	}
	return strings.Join(parts, ", ")
}

func GetInsertValues(fields []TmplField) string {
	var parts []string
	for _, fl := range fields {
		switch fl.ColumnName {
		case "id":
			continue
		case "created_at":
			parts = append(parts, "NOW()")
			continue
		default:
			parts = append(parts, "?")
		}
	}
	return strings.Join(parts, ", ")
}

func GetInsertArgs(m StructTmplData) string {
	var parts []string
	for _, fl := range m.Model.Fields {
		switch fl.Name {
		case "ID", "CreatedAt":
			continue
		}
		parts = append(parts, fmt.Sprintf("%s.%s", m.Receiver, fl.Name))
	}
	return strings.Join(parts, ", ")
}

func GetScanFields(m StructTmplData) template.HTML {
	var parts []string
	for _, fl := range m.Model.Fields {
		parts = append(parts, fmt.Sprintf("&%s.%s", m.Receiver, fl.Name))
	}
	return template.HTML(strings.Join(parts, ", "))
}

func GetUpdateArgs(m StructTmplData) template.HTML {
	var parts []string
	for _, fl := range m.Model.Fields {
		switch fl.Name {
		case "ID", "CreatedAt", "UpdatedAt":
			continue
		}
		parts = append(parts, fmt.Sprintf("%s.%s", m.Receiver, fl.Name))
	}
	return template.HTML(strings.Join(parts, ", "))
}

func GetUpdateValues(m StructTmplData) string {
	var parts []string
	for _, fl := range m.Model.Fields {
		switch fl.Name {
		case "ID", "CreatedAt":
			continue
		case "UpdatedAt":
			parts = append(parts, fmt.Sprintf("`%s`=UTC_TIMESTAMP()", fl.ColumnName))
		default:
			parts = append(parts, fmt.Sprintf("`%s`=?", fl.ColumnName))
		}
	}
	return strings.Join(parts, ", ")
}

func GetUpsertValues(m StructTmplData) string {
	var parts []string
	for _, fl := range m.Model.Fields {
		switch fl.Name {
		case "CreatedAt":
			continue
		case "ID":
			parts = append(parts, fmt.Sprintf("`%s`=LAST_INSERT_ID(`%s`)", fl.ColumnName, fl.ColumnName))
		case "UpdatedAt":
			parts = append(parts, fmt.Sprintf("`%s`=UTC_TIMESTAMP()", fl.ColumnName))
		default:
			parts = append(parts, fmt.Sprintf("`%s`=`%s`", fl.ColumnName, fl.ColumnName))
		}
	}
	return strings.Join(parts, ", ")
}
