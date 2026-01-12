package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"
)

type postRepository struct {
	client postgresql.Client
}

func NewPostRepository(client postgresql.Client) repository.PostRepository {
	return &postRepository{
		client: client,
	}
}

func (r *postRepository) Create(ctx context.Context, post *entity.Post) error {
	q := `
		INSERT INTO social.posts (user_id, content)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	if err := r.client.QueryRow(ctx, q, post.UserID, post.Content).
		Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt); err != nil {
		return err
	}

	return nil
}

func (r *postRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Post, error) {
	q := `
		SELECT id, user_id, content, created_at, updated_at
		FROM social.posts
		WHERE id = $1
	`

	var post entity.Post
	err := r.client.QueryRow(ctx, q, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("post not found")
		}
		return nil, err
	}

	return &post, nil
}

func (r *postRepository) Update(ctx context.Context, post *entity.Post) error {
	q := `
		UPDATE social.posts
		SET content = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
		RETURNING updated_at
	`

	if err := r.client.QueryRow(ctx, q, post.Content, post.ID).Scan(&post.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("post not found")
		}
		return err
	}

	return nil
}

func (r *postRepository) Delete(ctx context.Context, id uuid.UUID) error {
	q := `
		DELETE FROM social.posts
		WHERE id = $1
	`

	ct, err := r.client.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}
