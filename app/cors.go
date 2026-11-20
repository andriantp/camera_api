package app

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// HeaderClientType selects how the client sends tokens ("api" or "web").
const HeaderClientType = "X-Client-Type"

// productionCORS returns the CORS config used in production. It must list
// every method the API routes use and every request header clients send;
// otherwise browsers reject the preflight and the request never reaches
// the API.
func productionCORS(allowOrigins []string) middleware.CORSConfig {
	return middleware.CORSConfig{
		AllowOriginFunc: func(origin string) (bool, error) {
			if origin == "" {
				return true, nil
			}

			for _, allowed := range allowOrigins {
				if origin == allowed {
					return true, nil
				}
			}

			return false, nil
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			HeaderClientType,
		},
		AllowCredentials: true,
	}
}
