package repositories

import (
	"database/sql"
	"time"

	"github.com/tracewayapp/traceway/backend/app/db"
	"github.com/tracewayapp/traceway/backend/app/models"

	"github.com/tracewayapp/lit/v2"
)

type userRepository struct{}

func (r *userRepository) FindByEmail(tx *sql.Tx, email string) (*models.User, error) {
	return lit.SelectSingleNamed[models.User](
		tx,
		"SELECT id, email, name, password, created_at FROM users WHERE email = :email",
		lit.P{"email": email},
	)
}

func (r *userRepository) FindById(tx *sql.Tx, id int) (*models.User, error) {
	return lit.SelectSingleNamed[models.User](
		tx,
		"SELECT id, email, name, password, created_at FROM users WHERE id = :id",
		lit.P{"id": id},
	)
}

func (r *userRepository) Create(tx *sql.Tx, email string, name string, hashedPassword string) (*models.User, error) {
	user := &models.User{
		Email:     email,
		Name:      name,
		Password:  hashedPassword,
		CreatedAt: time.Now().UTC(),
	}

	id, err := lit.Insert(tx, user)
	if err != nil {
		return nil, err
	}
	user.Id = id

	return user, nil
}

func (r *userRepository) EmailExists(tx *sql.Tx, email string) (bool, error) {
	user, err := r.FindByEmail(tx, email)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

func (r *userRepository) SetPasswordResetToken(tx *sql.Tx, userId int, token string, expiresAt time.Time) error {
	now := time.Now()
	return lit.UpdateNamed[models.User](
		tx,
		&models.User{
			PasswordResetToken:       &token,
			PasswordResetExpiresAt:   &expiresAt,
			PasswordResetRequestedAt: &now,
		},
		"id = :id",
		lit.P{"id": userId},
	)
}

func (r *userRepository) ClearPasswordResetToken(tx *sql.Tx, userId int) error {
	q, a, err := lit.ParseNamedQuery(db.Driver, "UPDATE users SET password_reset_token = NULL, password_reset_expires_at = NULL, password_reset_requested_at = NULL WHERE id = :id", lit.P{"id": userId})
	if err != nil {
		return err
	}
	_, err = tx.Exec(q, a...)
	return err
}

func (r *userRepository) FindByPasswordResetToken(tx *sql.Tx, token string) (*models.User, error) {
	return lit.SelectSingleNamed[models.User](
		tx,
		"SELECT id, email, name, password, created_at, password_reset_token, password_reset_expires_at, password_reset_requested_at FROM users WHERE password_reset_token = :token",
		lit.P{"token": token},
	)
}

func (r *userRepository) UpdatePassword(tx *sql.Tx, userId int, hashedPassword string) error {
	return lit.UpdateNamed[models.User](
		tx,
		&models.User{Password: hashedPassword},
		"id = :id",
		lit.P{"id": userId},
	)
}

var UserRepository = userRepository{}
