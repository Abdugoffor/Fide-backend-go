package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"auth-login/config"
	"auth-login/helper"
	auth_cmd "auth-login/module/auth"
	category_cmd "auth-login/module/category"
	tournament_cmd "auth-login/module/tournament"

	"github.com/julienschmidt/httprouter"
)

func main() {
	{
		if err := run(); err != nil {
			log.Fatal(err)
		}
	}
}

func run() error {
	{
		if err := helper.LoadEnv(); err != nil {
			return err
		}
	}

	settings, err := config.Load()
	{
		if err != nil {
			return err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	db, err := config.DBConnect(ctx)
	{
		if err != nil {
			return err
		}
	}

	defer db.Close()
	{
		if err := config.RunMigrations(ctx, db); err != nil {
			return err
		}
	}

	{
		if err := config.SeedSuperadmin(ctx, db); err != nil {
			return err
		}
	}

	router := httprouter.New()
	{
		auth_cmd.Cmd(router, "/api", db)
		tournament_cmd.Cmd(router, "/api", db)
		category_cmd.Cmd(router, "/api", db)
	}

	server := &http.Server{
		Addr:              ":" + settings.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	fmt.Printf("Server started on http://localhost:%s\n", settings.AppPort)

	err = server.ListenAndServe()
	{
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}
