package field_handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"auth-login/helper"
	auth_dto "auth-login/module/auth/dto"
	auth_handler "auth-login/module/auth/handler"
	auth_service "auth-login/module/auth/service"
	field_dto "auth-login/module/field/dto"
	field_service "auth-login/module/field/service"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/julienschmidt/httprouter"
)

type fieldHandler struct {
	service *field_service.FieldService
}

func NewFieldHandler(router *httprouter.Router, group string, db *pgxpool.Pool) {
	handler := &fieldHandler{
		service: field_service.NewFieldService(db),
	}

	auth := auth_service.NewAuthService(db, helper.ENV("JWT_SECRET"))

	roles := []string{auth_dto.RoleSuperadmin, auth_dto.RoleAdmin}

	routes := group + "/fields"
	{
		router.POST(routes, auth_handler.RequireAuth(auth, handler.Create, roles...))
		router.GET(routes, auth_handler.RequireAuth(auth, handler.List))
		router.GET(routes+"/:id", auth_handler.RequireAuth(auth, handler.Show))
		router.PUT(routes+"/:id", auth_handler.RequireAuth(auth, handler.Update, roles...))
		router.DELETE(routes+"/:id", auth_handler.RequireAuth(auth, handler.Delete, roles...))
	}
}

func (handler *fieldHandler) Create(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var req field_dto.Create
	{
		if !parseBody(w, r, &req) {
			return
		}
	}

	res, err := handler.service.Create(r.Context(), &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusCreated, res)
}

func (handler *fieldHandler) List(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	query := r.URL.Query()

	filter := field_dto.Filter{
		Name: strings.TrimSpace(query.Get("name")),
		Type: strings.TrimSpace(query.Get("type")),
	}

	if query.Has("page_id") {
		value, err := strconv.ParseInt(query.Get("page_id"), 10, 64)
		{
			if err != nil || value < 1 {
				helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid page_id"})
				return
			}
		}

		filter.PageID = value
	}

	if filter.Type != "" {
		switch filter.Type {
		case field_dto.TypeText, field_dto.TypeTextarea, field_dto.TypeEmail, field_dto.TypeTel,
			field_dto.TypeNumber, field_dto.TypeDate, field_dto.TypeSelect, field_dto.TypeRadio,
			field_dto.TypeCheckbox, field_dto.TypeFile:
		default:
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid type"})
			return
		}
	}

	if query.Has("is_active") {
		raw := query.Get("is_active")

		if raw != "true" && raw != "false" {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": "is_active must be true or false"})
			return
		}

		isActive := raw == "true"

		filter.IsActive = &isActive
	}

	res, err := handler.service.List(r.Context(), &filter)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *fieldHandler) Show(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}
	}

	res, err := handler.service.Show(r.Context(), id)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *fieldHandler) Update(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}
	}

	var body field_dto.Create
	{
		if !parseBody(w, r, &body) {
			return
		}
	}

	req := field_dto.Update(body)

	res, err := handler.service.Update(r.Context(), id, &req)
	{
		if err != nil {
			writeServiceError(w, err)
			return
		}
	}

	helper.JSON(w, http.StatusOK, res)
}

func (handler *fieldHandler) Delete(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	id, ok := helper.ParseID(w, params)
	{
		if !ok {
			return
		}

		if err := handler.service.Delete(r.Context(), id); err != nil {
			writeServiceError(w, err)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func parseBody(w http.ResponseWriter, r *http.Request, req *field_dto.Create) bool {
	{
		if err := helper.DecodeJSON(w, r, req); err != nil {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return false
		}
	}

	req.Name = strings.TrimSpace(req.Name)

	req.Type = strings.TrimSpace(req.Type)
	for i := range req.Options {
		req.Options[i] = strings.TrimSpace(req.Options[i])
	}

	if errs := helper.Validate(req); errs != nil {
		helper.JSON(w, http.StatusBadRequest, map[string]any{"errors": errs})
		return false
	}

	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	{
		if errors.Is(err, pgx.ErrNoRows) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": "field not found"})
			return
		}

		if errors.Is(err, field_service.ErrPageNotFound) {
			helper.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}

		if errors.Is(err, field_service.ErrInvalidField) || errors.Is(err, field_service.ErrInvalidOptions) || errors.Is(err, field_service.ErrInvalidValidation) {
			helper.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			helper.JSON(w, http.StatusConflict, map[string]string{"error": "field sort or name already exists on this page"})
			return
		}

		if pgErr.Code == "23503" {
			helper.JSON(w, http.StatusConflict, map[string]string{"error": "page changed; retry the request"})
			return
		}
	}

	log.Printf("Field service error: %v", err)

	helper.JSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}
