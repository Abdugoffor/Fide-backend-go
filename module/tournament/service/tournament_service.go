package tournament_service

import (
	"context"
	"errors"
	"slices"

	tournament_dto "auth-login/module/tournament/dto"
	tournament_repository "auth-login/module/tournament/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInvalidDates = errors.New("ends_at must be after starts_at")

type TournamentService struct {
	repository tournament_repository.TournamentRepository
}

func NewTournamentService(db *pgxpool.Pool) *TournamentService {
	return &TournamentService{
		repository: tournament_repository.NewTournamentRepository(db),
	}
}

func (service *TournamentService) Create(ctx context.Context, req *tournament_dto.Create) (*tournament_dto.Response, error) {
	if !req.EndsAt.After(req.StartsAt) {
		return nil, ErrInvalidDates
	}

	return service.repository.Create(ctx, req)
}

func (service *TournamentService) List(ctx context.Context, filter *tournament_dto.Filter) (*tournament_dto.ListResponse, error) {
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 20
	}

	rows, err := service.repository.List(ctx, filter)
	{
		if err != nil {
			return nil, err
		}
	}

	hasMore := len(rows) > filter.Limit

	if hasMore {
		rows = rows[:filter.Limit]
	}

	if filter.Direction == tournament_dto.DirectionPrev {
		slices.Reverse(rows)
	}

	res := &tournament_dto.ListResponse{Data: rows}

	if len(rows) == 0 {
		return res, nil
	}

	firstID := rows[0].ID

	lastID := rows[len(rows)-1].ID

	if filter.Direction == tournament_dto.DirectionPrev {
		res.HasPrev = hasMore

		res.HasNext, err = service.repository.HasMore(ctx, filter, lastID, tournament_dto.DirectionNext)

	} else {

		res.HasNext = hasMore

		if filter.Cursor > 0 {
			res.HasPrev, err = service.repository.HasMore(ctx, filter, firstID, tournament_dto.DirectionPrev)
		}
	}

	if err != nil {
		return nil, err
	}

	if res.HasNext {
		res.NextCursor = &lastID
	}

	if res.HasPrev {
		res.PrevCursor = &firstID
	}

	return res, nil
}

func (service *TournamentService) Show(ctx context.Context, id int64) (*tournament_dto.Response, error) {
	return service.repository.Show(ctx, id)
}

func (service *TournamentService) Update(ctx context.Context, id int64, req *tournament_dto.Update) (*tournament_dto.Response, error) {
	if !req.EndsAt.After(req.StartsAt) {
		return nil, ErrInvalidDates
	}

	return service.repository.Update(ctx, id, req)
}

func (service *TournamentService) Delete(ctx context.Context, id int64) error {
	return service.repository.Delete(ctx, id)
}
