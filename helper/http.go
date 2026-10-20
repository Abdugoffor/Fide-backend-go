package helper

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

func ParseID(w http.ResponseWriter, params httprouter.Params) (int64, bool) {
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	{
		if err != nil || id < 1 {
			JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
			return 0, false
		}
	}

	return id, true
}
