package page_cmd

import (
	page_handler "auth-login/module/page/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, group string, db *pgxpool.Pool) {
	page_handler.NewPageHandler(router, group, db)
}
