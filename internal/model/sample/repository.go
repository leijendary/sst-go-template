package sample

import (
	"context"
	"database/sql"
	"fmt"
	"sst-go-template/internal/db"
	"sst-go-template/internal/response"
	"strings"
	"time"
)

type Sample struct {
	ID             int64
	Name           string
	Description    string
	Amount         float64
	Version        int8
	Translations   []*SampleTranslation
	CreatedAt      time.Time
	CreatedBy      string
	LastModifiedAt time.Time
	LastModifiedBy string
}

type SampleTranslation struct {
	Name        string
	Description string
	Language    string
	Ordinal     int8
}

type Repository interface {
	save(ctx context.Context, tx *sql.Tx, s *Sample) error
	get(ctx context.Context, id int64) (*Sample, error)
}

type repository struct {
	conn *sql.DB
}

func NewRepository(conn *sql.DB) *repository {
	return &repository{conn: conn}
}

func (r *repository) save(ctx context.Context, tx *sql.Tx, s *Sample) error {
	query := `INSERT INTO sample (name, description, amount, created_by, last_modified_by)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id, name, description, amount, created_at`
	row := tx.QueryRowContext(ctx, query, s.Name, s.Description, s.Amount, s.CreatedBy, s.LastModifiedBy)
	if err := row.Scan(&s.ID, &s.Name, &s.Description, &s.Amount, &s.CreatedAt); err != nil {
		return db.ParseError(err)
	}

	return saveTranslations(ctx, tx, s.ID, s.Translations)
}

func (r *repository) get(ctx context.Context, id int64) (*Sample, error) {
	query := `SELECT id, name, description, amount, version, created_at FROM sample WHERE id = $1 and deleted_at is null`
	row := r.conn.QueryRowContext(ctx, query, id)
	var s Sample
	if err := row.Scan(&s.ID, &s.Name, &s.Description, &s.Amount, &s.Version, &s.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, response.ResourceNotFound(id, "/data/sample")
		}

		return nil, db.ParseError(err)
	}

	st, err := getTranslations(ctx, r.conn, id)
	if err != nil {
		return nil, err
	}

	s.Translations = st

	return &s, nil
}

func saveTranslations(ctx context.Context, tx *sql.Tx, id int64, st []*SampleTranslation) error {
	if len(st) == 0 {
		return nil
	}

	query := `INSERT INTO sample_translation (id, name, description, language, ordinal)
	VALUES %s
	RETURNING name, description, language, ordinal`
	params := []string{}
	args := []any{}
	for i, v := range st {
		param := fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
		params = append(params, param)
		args = append(args, id, v.Name, v.Description, v.Language, v.Ordinal)
	}

	param := strings.Join(params, ", ")
	query = fmt.Sprintf(query, param)
	rows, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return db.ParseError(err)
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		if err := rows.Scan(&st[i].Name, &st[i].Description, &st[i].Language, &st[i].Ordinal); err != nil {
			return db.ParseError(err)
		}
	}

	return nil
}

func getTranslations(ctx context.Context, conn *sql.DB, id int64) ([]*SampleTranslation, error) {
	query := `SELECT name, description, language, ordinal FROM sample_translation where id = $1`
	rows, err := conn.QueryContext(ctx, query, id)
	if err != nil {
		return nil, db.ParseError(err)
	}
	defer rows.Close()

	st := []*SampleTranslation{}
	var t SampleTranslation
	for i := 0; rows.Next(); i++ {
		if err := rows.Scan(&t.Name, &t.Description, &t.Language, &t.Ordinal); err != nil {
			return nil, db.ParseError(err)
		}
		st = append(st, &t)
	}

	return st, nil
}
