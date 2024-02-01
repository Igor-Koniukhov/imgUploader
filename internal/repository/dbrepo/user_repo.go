package dbrepo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/net/context"
	"imageAploaderS3/models"
	"time"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, sql.Result, error)
	GetUserByEmail(email string) (*models.User, error)
	GetUserFromCacheByEmail(email string) (*models.User, error)
}
type UserRepo struct {
	DB        *sql.DB
	PrimaryRC *redis.Client
	ReaderRC  *redis.Client
	User      models.User
}

func NewUserRepository(db *sql.DB, pdb *redis.Client, rdb *redis.Client) *UserRepo {
	return &UserRepo{
		DB:        db,
		PrimaryRC: pdb,
		ReaderRC:  rdb,
	}
}

func (usr UserRepo) GetUserFromCacheByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var userFromCache *models.User
	val, err := usr.ReaderRC.Get(ctx, "user:"+email).Result()
	if err != nil {
		fmt.Println("Error getting value from Redis:", err)
		return nil, err
	}
	err = json.Unmarshal([]byte(val), &userFromCache)
	if err != nil {
		return nil, err
	}

	fmt.Println("Got value from Redis:", val, userFromCache)
	return userFromCache, nil
}

func (usr UserRepo) CreateUser(user *models.User) (*models.User, sql.Result, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	sqlStmt := fmt.Sprintf("INSERT INTO %s (name, email, birth_date) VALUES($1, $2, $3) RETURNING id", models.TableUsers)
	userData, err := json.Marshal(user)
	if err != nil {
		fmt.Println("Error marshaling user data:", err)
		return nil, nil, err
	}
	err = usr.PrimaryRC.Set(ctx, "user:"+user.Email, userData, 0).Err()
	if err != nil {
		fmt.Println("Error setting value in Redis:", err)
		return nil, nil, err
	}

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
