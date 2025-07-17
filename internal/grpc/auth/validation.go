package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/hesoyamTM/apphelper-sso/internal/models"
)

func validateRegister(ctx context.Context, name, surname, login, pass string) error {
	validate := validator.New()
	if err := validate.VarCtx(ctx, name, "required,lte=20"); err != nil {
		fmt.Println("name " + err.Error())
		return err
	}
	if err := validate.VarCtx(ctx, surname, "required,lte=20"); err != nil {
		fmt.Println("surname " + err.Error())
		return err
	}
	if err := validate.VarCtx(ctx, login, "required,lte=50"); err != nil {
		fmt.Println("login " + err.Error())
		return err
	}
	if err := validate.VarCtx(ctx, pass, "required,gte=8,lte=20"); err != nil {
		fmt.Println("password " + err.Error())
		return err
	}
	return nil
}

func validateLogin(ctx context.Context, login, pass string) error {
	validate := validator.New()
	if err := validate.VarCtx(ctx, login, "required,lte=50"); err != nil {
		return err
	}
	if err := validate.VarCtx(ctx, pass, "required,gte=8,lte=20"); err != nil {
		return err
	}
	return nil
}

func validateRefreshToken(ctx context.Context, token string) error {
	validate := validator.New()
	if err := validate.VarCtx(ctx, token, "required"); err != nil {
		return err
	}
	return nil
}

func validateGetUsers(ids []uuid.UUID) error {
	if ids == nil {
		return errors.New("id array is empty")
	}
	return nil
}

func validateLogout(ctx context.Context, token string) error {
	validate := validator.New()
	if err := validate.VarCtx(ctx, token, "required"); err != nil {
		return err
	}
	return nil
}

func validateUpdateUser(ctx context.Context, user models.UserInfo) error {
	validate := validator.New()
	if err := validate.VarCtx(ctx, user.Name, "required,lte=20"); err != nil {
		return err
	}
	if err := validate.VarCtx(ctx, user.Surname, "required,lte=20"); err != nil {
		return err
	}
	return nil
}

func validateChangePassword(ctx context.Context, newPassword, token string) error {
	validate := validator.New()
	if err := validate.VarCtx(ctx, newPassword, "required,gte=8,lte=20"); err != nil {
		return err
	}
	if err := validate.VarCtx(ctx, token, "required"); err != nil {
		return err
	}
	return nil
}
