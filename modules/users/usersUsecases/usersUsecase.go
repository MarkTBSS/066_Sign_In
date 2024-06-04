package usersUsecases

import (
	"fmt"

	"github.com/MarkTBSS/066_Sign_In/config"
	"github.com/MarkTBSS/066_Sign_In/modules/users"
	"github.com/MarkTBSS/066_Sign_In/modules/users/usersRepositories"
	"github.com/MarkTBSS/066_Sign_In/pkg/kawaiiauth"
	"golang.org/x/crypto/bcrypt"
)

type IUsersUsecase interface {
	InsertCustomer(req *users.UserRegisterReq) (*users.UserPassport, error)
	GetPassport(req *users.UserCredential) (*users.UserPassport, error)
}

type usersUsecase struct {
	cfg             config.IConfig
	usersRepository usersRepositories.IUsersRepository
}

func UsersUsecase(cfg config.IConfig, usersRepository usersRepositories.IUsersRepository) IUsersUsecase {
	return &usersUsecase{
		cfg:             cfg,
		usersRepository: usersRepository,
	}
}

func (u *usersUsecase) GetPassport(req *users.UserCredential) (*users.UserPassport, error) {
	// Find user
	user, err := u.usersRepository.FindOneUserByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("password is invalid")
	}

	// Sign token
	accessToken, _ := kawaiiauth.NewKawaiiAuth(kawaiiauth.Access, u.cfg.Jwt(),
		&users.UserClaims{
			Id:     user.Id,
			RoleId: user.RoleId,
		})
	refreshToken, _ := kawaiiauth.NewKawaiiAuth(kawaiiauth.Refresh, u.cfg.Jwt(),
		&users.UserClaims{
			Id:     user.Id,
			RoleId: user.RoleId,
		})

	// Set passport
	passport := &users.UserPassport{
		User: &users.User{
			Id:       user.Id,
			Email:    user.Email,
			Username: user.Username,
			RoleId:   user.RoleId,
		},
		Token: &users.UserToken{
			AccessToken:  accessToken.SignToken(),
			RefreshToken: refreshToken.SignToken(),
		},
	}
	err = u.usersRepository.InsertOauth(passport)
	if err != nil {
		return nil, err
	}
	return passport, nil
}

func (u *usersUsecase) InsertCustomer(req *users.UserRegisterReq) (*users.UserPassport, error) {
	// Hashing a password
	if err := req.BcryptHashing(); err != nil {
		return nil, err
	}
	// Insert user
	result, err := u.usersRepository.InsertUser(req, false)
	if err != nil {
		return nil, err
	}
	return result, nil
}
