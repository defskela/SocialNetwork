package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/defskela/SocialNetwork/internal/entity"
	"github.com/defskela/SocialNetwork/internal/repository"
	"github.com/defskela/SocialNetwork/pkg/client/postgresql"
)

type followerRepository struct {
	client postgresql.Client
}

func NewFollowerRepository(client postgresql.Client) repository.FollowerRepository {
	return &followerRepository{
		client: client,
	}
}

func (r *followerRepository) Follow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	q := `
		INSERT INTO social.followers (follower_id, followee_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`
	_, err := r.client.Exec(ctx, q, followerID, followeeID)
	return err
}

func (r *followerRepository) Unfollow(ctx context.Context, followerID, followeeID uuid.UUID) error {
	q := `
		DELETE FROM social.followers
		WHERE follower_id = $1 AND followee_id = $2
	`
	ct, err := r.client.Exec(ctx, q, followerID, followeeID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("relationship not found")
	}
	return nil
}

func (r *followerRepository) GetFollowers(ctx context.Context, userID uuid.UUID) ([]entity.User, error) {
	q := `
		SELECT u.id, u.username, u.email, u.bio, u.birthday, u.created_at
		FROM social.users u
		JOIN social.followers f ON u.id = f.follower_id
		WHERE f.followee_id = $1
	`
	return r.fetchUsers(ctx, q, userID)
}

func (r *followerRepository) GetFollowing(ctx context.Context, userID uuid.UUID) ([]entity.User, error) {
	q := `
		SELECT u.id, u.username, u.email, u.bio, u.birthday, u.created_at
		FROM social.users u
		JOIN social.followers f ON u.id = f.followee_id
		WHERE f.follower_id = $1
	`
	return r.fetchUsers(ctx, q, userID)
}

func (r *followerRepository) fetchUsers(ctx context.Context, query string, arg interface{}) ([]entity.User, error) {
	rows, err := r.client.Query(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Bio, &u.Birthday, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
