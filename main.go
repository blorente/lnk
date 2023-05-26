// main.go
package main

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func updateViews(e *core.RecordViewEvent) error {
	log.Info().Msgf("%v", e.Record)
	return nil
}

func registerHandleLinkSlugs(app *pocketbase.PocketBase) {
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.AddRoute(echo.Route{
			Method: http.MethodGet,
			Path:   "/lnk/:slug",
			Handler: func(c echo.Context) error {
				slug := c.PathParam("slug")
				log.Debug().Str("slug", slug).Msg("Requested link")
				record, err := app.Dao().FindFirstRecordByData("links", "slug", slug)
				if err != nil {
					return apis.NewNotFoundError("The link does not exist.", err)
				}

				count := record.GetInt("count")
				target := record.GetString("target")
				log.Debug().Str("target", target).Int("count", count).Msg("Got Record")

				record.Set("count", count+1)
				// TODO Schedule saves and flush on intervals
				if err := app.Dao().SaveRecord(record); err != nil {
					return err
				}
				err = c.Redirect(http.StatusPermanentRedirect, target)
				log.Debug().Msg("Redirect successful")
				return err
			},
			Middlewares: []echo.MiddlewareFunc{
				apis.ActivityLogger(app),
			},
		})
		return nil
	},
	)
}

type Args struct {
	datadir string
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// args := Args{}
	// flag.StringVar(&args.datadir, "datadir", "/tmp/pb_data", "Directory to store the pocketbase data")
	// flag.Parse()
	// log.Info().Any("config", args).Msg("Initialized with config")

	app := pocketbase.NewWithConfig(&pocketbase.Config{
		// DefaultDataDir: args.datadir,
	})
	log.Info().Msg("Setting up")

	log.Info().Msg("Installing hooks")
	registerHandleLinkSlugs(app)
	// or you can also use the shorter e.Router.GET("/articles/:slug", handler, middlewares...)
	log.Info().Msg("Starting server")
	if err := app.Start(); err != nil {
		log.Fatal().Err(err)
	}
}
