package cbwebcommon

import (
	"encoding/json"
	"html/template"
)

type DataTable struct {
	TableId template.JS
	Columns []string
	Data [][]interface{}
	AjaxRoute string
}

func (d *DataTable) GetTableId() template.JS {
	return d.TableId
}

func (d *DataTable) GetDataJson() template.JS {
	jsonBytes, e := json.Marshal(d.Data)
	if e == nil {
		return template.JS(jsonBytes)
	}

	return ""
}

func (d *DataTable) GetColumnsJson() template.JS {
	jsonBytes, e := json.Marshal(d.Columns)
	if e == nil {
		return template.JS(jsonBytes)
	}

	return ""
}

func (d *DataTable) GetColumns() []string {
	return d.Columns
}

func (d *DataTable) HasData() bool {
	return len(d.Data) != 0
}

func (d *DataTable) GetAjaxRoute() string {
	return d.AjaxRoute
}

func (d *DataTable) IsAjax() bool {
	return d.AjaxRoute != ""
}