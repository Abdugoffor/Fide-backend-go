package tournament_cmd

import (
	tournament_handler "auth-login/module/tournament/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

func Cmd(router *httprouter.Router, group string, db *pgxpool.Pool) {
	tournament_handler.NewTournamentHandler(router, group, db)
}
