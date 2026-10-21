package field_cmd

import (
	field_handler "auth-login/module/field/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, group string, db *pgxpool.Pool) {
	field_handler.NewFieldHandler(router, group, db)
}
