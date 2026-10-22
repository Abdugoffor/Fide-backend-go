package application_cmd

import (
	application_handler "auth-login/module/application/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, group string, db *pgxpool.Pool) {
	application_handler.NewApplicationHandler(router, group, db)
}
