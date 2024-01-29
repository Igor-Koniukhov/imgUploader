package dbrepo

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/net/context"
	"imageAploaderS3/models"
	"time"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, sql.Result, error)
	GetUserByEmail(email string) (*models.User, error)
}
type UserRepo struct {
	DB   *sql.DB
	User models.User
}

func NewUserRepository(db *sql.DB) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

func (usr UserRepo) CreateUser(user *models.User) (*models.User, sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sqlStmt := fmt.Sprintf("INSERT INTO %s (name, email, birth_date) VALUES($1, $2, $3) RETURNING id", models.TableUsers)

	res, err := usr.DB.ExecContext(ctx, sqlStmt,
		user.Name,
		user.Email,
		user.BirthDate)
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	return user, res, nil
}

func (usr UserRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	sqlStmt := fmt.Sprintf("SELECT id, name, email FROM %s WHERE email = $1", models.TableUsers)
	var user models.User
	err := usr.DB.QueryRowContext(ctx, sqlStmt, email).Scan(&user.ID, &user.Name, &user.Email)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println(err)
		return nil, nil
	} else if err != nil {
		fmt.Println("GET user error: ", err)
		return nil, err
	}
	return &user, nil
}
