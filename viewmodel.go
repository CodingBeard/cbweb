package cbweb

import "html/template"

type ViewIncludeType string

var (
	ViewIncludeType_JsHead           ViewIncludeType = "js-head"
	ViewIncludeType_JsHeadInline     ViewIncludeType = "js-head-inline"
	ViewIncludeType_CssHead          ViewIncludeType = "css-head"
	ViewIncludeType_CssHeadInline    ViewIncludeType = "css-head-inline"
	ViewIncludeType_JsBody           ViewIncludeType = "js-body"
	ViewIncludeType_JsBodyInline     ViewIncludeType = "js-body-inline"
	ViewIncludeType_CssBody          ViewIncludeType = "css-body"
	ViewIncludeType_CssBodyInline    ViewIncludeType = "css-body-inline"
	ViewIncludeType_JsPostBody       ViewIncludeType = "js-postBody"
	ViewIncludeType_JsPostBodyInline ViewIncludeType = "js-postBody-inline"
)

type ViewInclude struct {
	Type      ViewIncludeType
	Src       template.URL
	Html      template.HTML
	Attribute template.HTMLAttr
	Js        template.JS
	Css       template.CSS
}

type NavItem struct {
	Src         template.URL
	Title       string
	Active      bool
	SubNavItems []NavItem
	Divider     bool
	Permitted   bool
}

type NavItemCollection []NavItem

func (n *NavItemCollection) FilterPermitted() []NavItem {
	return n.filterPermitted(*n)
}

func (n *NavItemCollection) filterPermitted(navItems NavItemCollection) []NavItem {
	var newItems []NavItem

	for _, item := range navItems {
		if len(item.SubNavItems) > 0 {
			item.SubNavItems = n.filterPermitted(item.SubNavItems)
			if len(item.SubNavItems) > 0 {
				newItems = append(newItems, item)
				continue
			}
		} else {
			if item.Permitted {
				newItems = append(newItems, item)
				continue
			}
		}
	}

	return newItems
}

type MasterViewModel interface {
	GetMasterViewModel() DefaultMasterViewModel
}

type ExecutableViewModel interface {
	GetTemplates() []string
	GetMainTemplate() string
}

type DefaultMasterViewModel struct {
	ViewIncludes []ViewInclude
	Title        string
	PageTitle    string
	BodyClasses  string
	NavItems     []NavItem
	Path         template.URL
	Flash        *Flash
}

func (m DefaultMasterViewModel) GetViewIncludes() []ViewInclude {
	return m.ViewIncludes
}

func (m DefaultMasterViewModel) GetTitle() string {
	return m.Title
}

func (m DefaultMasterViewModel) GetPageTitle() string {
	return m.PageTitle
}

func (m DefaultMasterViewModel) GetBodyClasses() string {
	return m.BodyClasses
}

func (m DefaultMasterViewModel) GetNavItems() []NavItem {
	return m.NavItems
}

func (m DefaultMasterViewModel) GetPath() template.URL {
	return m.Path
}

func (m DefaultMasterViewModel) GetFlash() *Flash {
	if m.Flash == nil {
		m.Flash = &Flash{}
	}
	return m.Flash
}

func (h ViewIncludeType) IsJsHead() bool {
	return h == ViewIncludeType_JsHead
}

func (h ViewIncludeType) IsJsHeadInline() bool {
	return h == ViewIncludeType_JsHeadInline
}

func (h ViewIncludeType) IsCssHead() bool {
	return h == ViewIncludeType_CssHead
}

func (h ViewIncludeType) IsCssHeadInline() bool {
	return h == ViewIncludeType_CssHeadInline
}

func (h ViewIncludeType) IsJsBody() bool {
	return h == ViewIncludeType_JsBody
}

func (h ViewIncludeType) IsJsBodyInline() bool {
	return h == ViewIncludeType_JsBodyInline
}

func (h ViewIncludeType) IsCssBody() bool {
	return h == ViewIncludeType_CssBody
}

func (h ViewIncludeType) IsCssBodyInline() bool {
	return h == ViewIncludeType_CssBodyInline
}

func (h ViewIncludeType) IsJsPostBody() bool {
	return h == ViewIncludeType_JsPostBody
}

func (h ViewIncludeType) IsJsPostBodyInline() bool {
	return h == ViewIncludeType_JsPostBodyInline
}

// This is here purely for typehinting in go template files
type TypehintingViewModel struct{}

func (t TypehintingViewModel) GetMasterViewModel() DefaultMasterViewModel {
	return DefaultMasterViewModel{}
}
