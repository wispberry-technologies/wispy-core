package apiv1

import (
	"wispy-core/core/apiv1/forms"
	"wispy-core/core/site"

	"github.com/go-chi/chi/v5"
)

func MountApiV1(router chi.Router, siteManager site.SiteManager) chi.Router {
	router.Route("/v1", func(r chi.Router) {

		formApi := forms.NewFormApi(siteManager)

		formApi.MountApi(r)

	})

	return router
}
