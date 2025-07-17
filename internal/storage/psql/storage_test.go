package psql

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
	"github.com/hesoyamTM/apphelper-sso/pkg/logger"
)

var (
	// psqlConfig = migrations.Config{
	// 	Host:     "localhost",
	// 	Port:     5432,
	// 	User:     "root",
	// 	Password: "1234",
	// 	DB:       "auth",
	// }
	cfg = PsqlConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "root",
		Password: "1234",
		DB:       "auth",
	}
)

func TestCreateUser(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// if err := migrations.RunMigrations(ctx, psqlConfig, "migrations"); err != nil {
	// 	t.Errorf("unexpected error: %v", err)
	// }
	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// clear
	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProvideUserById(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	user, err := db.ProvideUserById(ctx, id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.UserInfo.Id != id {
		t.Errorf("unexpected user id: %v", user.UserInfo.Id)
	}

	if user.UserAuth.Id != id {
		t.Errorf("unexpected user auth id: %v", user.UserAuth.Id)
	}

	if user.Name != "John" {
		t.Errorf("unexpected user name: %v", user.Name)
	}

	if user.Surname != "Doe" {
		t.Errorf("unexpected user surname: %v", user.Surname)
	}

	if user.Email != "john.doe@example.com" {
		t.Errorf("unexpected user email: %v", user.Email)
	}

	// clear
	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProvideUsers(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	users, err := db.ProvideUsersById(ctx, []uuid.UUID{id})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if users[0].UserInfo.Id != id {
		t.Errorf("unexpected users id: %v", users[0].UserInfo.Id)
	}
	if users[0].UserAuth.Email != "john.doe@example.com" {
		t.Errorf("unexpected users email: %v", users[0].UserAuth.Email)
	}
	if users[0].Name != "John" {
		t.Errorf("unexpected users name: %v", users[0].Name)
	}
	if users[0].Surname != "Doe" {
		t.Errorf("unexpected users surname: %v", users[0].Surname)
	}

	// clear
	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestProvideUserByEmail(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	user, err := db.ProvideUserByEmail(ctx, "john.doe@example.com")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.UserInfo.Id != id {
		t.Errorf("unexpected users id: %v", user.UserInfo.Id)
	}
	if user.UserAuth.Email != "john.doe@example.com" {
		t.Errorf("unexpected users email: %v", user.UserAuth.Email)
	}
	if user.Name != "John" {
		t.Errorf("unexpected users name: %v", user.Name)
	}
	if user.Surname != "Doe" {
		t.Errorf("unexpected users surname: %v", user.Surname)
	}

	// clear
	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestUpdateUser(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	newName := "Johnny"
	newSurname := "Doe2"

	if err := db.UpdateUser(ctx, models.UserInfo{
		Id:      id,
		Name:    newName,
		Surname: newSurname,
	}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	user, err := db.ProvideUserById(ctx, id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if user.UserInfo.Id != id {
		t.Errorf("unexpected users id: %v", user.UserInfo.Id)
	}
	if user.UserAuth.Email != "john.doe@example.com" {
		t.Errorf("unexpected users email: %v", user.UserAuth.Email)
	}
	if user.Name != newName {
		t.Errorf("unexpected users name: %v", user.Name)
	}
	if user.Surname != newSurname {
		t.Errorf("unexpected users surname: %v", user.Surname)
	}

	// clear
	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestChangePassword(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte("password"))
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := db.ChangePassword(ctx, "john.doe@example.com", []byte("new-password")); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	user, err := db.ProvideUserById(ctx, id)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(user.PassHash) != "new-password" {
		t.Errorf("unexpected user password: %v", user.PassHash)
	}

	// clear
	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteUser(t *testing.T) {
	ctx, err := logger.New(context.Background(), "dev")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	db, err := New(ctx, cfg)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	id, err := db.CrateUser(ctx, "John", "Doe", "john.doe@example.com", []byte{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if err := db.DeleteUser(ctx, id); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
