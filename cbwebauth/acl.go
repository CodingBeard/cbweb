package cbwebauth

import (
	"github.com/valyala/fasthttp"
	"strings"
)

var (
	AclDeniedUriPlaceholder = ":acl_denied_uri:"
)

type Acl struct {
	Auth *Container
}

func (a *Acl) Middleware(permittedPermissions []string, redirect string) func(ctx *fasthttp.RequestCtx) (bool, error) {
	return func(ctx *fasthttp.RequestCtx) (bool, error) {
		if a.Auth == nil {
			return true, nil
		}

		if a.PermittedCtx(ctx, permittedPermissions) {
			return true, nil
		}

		if redirect != "" {
			ctx.Redirect(strings.Replace(redirect, AclDeniedUriPlaceholder, string(ctx.RequestURI()), -1), 302)
		}

		return false, nil
	}
}

func (a *Acl) PermittedCtx(ctx *fasthttp.RequestCtx, permittedPermissions []string) bool {
	if a.Auth == nil {
		return true
	}

	permissions := []string{LoggedOut}
	for _, provider := range a.Auth.providers {
		if provider.IsAuthenticated(ctx) {
			permissions = []string{}
		}
	}

	for _, provider := range a.Auth.providers {
		if a.Permitted(append(permissions, provider.GetPermissions(ctx)...), permittedPermissions) {
			return true
		}
	}

	return false
}

func (a *Acl) Permitted(userPermissions, permittedPermissions []string) bool {
	for _, allow := range permittedPermissions {
		for _, permission := range userPermissions {
			if strings.ToLower(allow) == strings.ToLower(permission) {
				return true
			}
		}
	}

	return false
}
