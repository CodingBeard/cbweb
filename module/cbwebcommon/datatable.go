package cbwebcommon

import (
	"encoding/json"
	"html/template"
)

var DataTableEditButtonHtml = `<a data-id="%d" class="material-table-edit-row" href="#"><i class="material-icons">edit</i></a>`

type DataTable struct {
	TableId           template.JS
	Columns           []DataTableColumn
	Data              [][]interface{}
	AjaxRoute         string
	GroupByColumn     bool
	GroupColumnOffset int
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
	var columns []string
	for _, column := range d.Columns {
		columns = append(columns, column.Title)
	}
	jsonBytes, e := json.Marshal(columns)
	if e == nil {
		return template.JS(jsonBytes)
	}

	return ""
}

func (d *DataTable) GetColumns() []DataTableColumn {
	return d.Columns
}

func (d *DataTable) GetColumnsLen() int {
	return len(d.Columns)
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

type DataTableColumn struct {
	Title        string
	Filterable   bool
	Editable     bool
	EditableName string
}

func (c *DataTableColumn) GetTitle() string {
	return c.Title
}

func (c *DataTableColumn) GetFilterable() bool {
	return c.Filterable
}

func (c *DataTableColumn) GetEditable() bool {
	return c.Editable
}

func (c *DataTableColumn) GetEditableName() string {
	return c.EditableName
}
