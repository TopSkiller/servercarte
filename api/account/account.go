package account

import (
	"context"
	"errors"
	"time"

	"github.com/CaninoDev/gastro/server/api"
	"github.com/CaninoDev/gastro/server/api/authentication"
	"github.com/CaninoDev/gastro/server/api/security"
	"github.com/CaninoDev/gastro/server/api/user"

	"github.com/google/uuid"
)

// Account are the contracted methods to interact with GORM
type Account struct {
	accountRepo Repository
	userRepo    user.Repository
	secSvc      security.Service
	authSvc     authentication.Service
}

func Bind(accountRepo Repository, userRepo user.Repository, secSvc security.Service,
	authSvc authentication.Service) *Account {
	return &Account{
		accountRepo, userRepo, secSvc, authSvc,
	}
}

func Initialize(accountRepo Repository, userRepo user.Repository, secSvc security.Service,
	authSvc authentication.Service) *Account {
	return Bind(accountRepo, userRepo, secSvc, authSvc)
}

func (a *Account) New(ctx context.Context, req NewAccountRequest) error {
	var newUser api.User
	newUser.FirstName = req.FirstName
	newUser.LastName = req.LastName
	newUser.Address1 = req.Address1
	newUser.Address2 = *req.Address2
	newUser.ZipCode = req.ZipCode
	newUser.Email = req.Email
	if err := a.userRepo.Search(ctx, &newUser); err == nil {
		return errors.New("account already exists")
	}

	if !a.secSvc.ConfirmationChecker(ctx, req.Password, req.PasswordConfirm) {
		return errors.New("passwords don't match")
	}

	if err := a.secSvc.IsValid(ctx, req.Password); err != nil {
		return err
	}

	var newAccount api.Account
	newAccount.Username = req.Username
	if err := a.accountRepo.Find(ctx, &newAccount); err == nil {
		return errors.New("username already exists")
	}

	newAccount.Password = a.secSvc.Hash(ctx, req.Password)

	if err := a.userRepo.Create(ctx, &newUser); err != nil {
		return err
	}
	newAccount.UserID = newUser.ID

	if err := a.accountRepo.Create(ctx, &newAccount); err != nil {
		return err
	}
	return nil

}

func (a *Account) Authenticate(ctx context.Context, username, password string) (string, error) {
	var acct api.Account
	acct.Username = username

	if err := a.accountRepo.Find(ctx, &acct); err != nil {
		return "", err
	}

	if !a.secSvc.VerifyPasswordMatches(ctx, acct.Password, password) {
		return "", errors.New("unauthorized")
	}

	token, err := a.authSvc.GenerateToken(ctx, &acct)
	if err != nil {
		return "", err
	}

	acct.LastLogin = time.Now()
	acct.Token = token
	if err := a.accountRepo.Update(ctx, &acct); err != nil {
		return "", err
	}
	return token, nil
}

func (a *Account) FindByUsername(ctx context.Context, username string) (*api.Account, error) {
	var acct api.Account
	acct.Username = username
	if err := a.accountRepo.Find(ctx, &acct); err != nil {
		return &api.Account{}, err
	}
	return &acct, nil
}

func (a *Account) FindByToken(ctx context.Context, token string) (*api.Account, error) {
	var acct api.Account
	acct.Token = token
	if err := a.accountRepo.Find(ctx, &acct); err != nil {
		return &api.Account{}, err
	}
	return &acct, nil
}

func (a *Account) ChangePassword(ctx context.Context, username, oldPassword, newPassword, confirmNewPassword string) error {
	var acct api.Account
	acct.Username = username
	if newPassword != confirmNewPassword {
		return errors.New("passwords don't match")
	}

	if err := a.accountRepo.Find(ctx, &acct); err != nil {
		return err
	}
	if !a.secSvc.VerifyPasswordMatches(ctx, acct.Password, oldPassword) {
		return errors.New("password incorrect")
	}

	encryptedPW := a.secSvc.Hash(ctx, newPassword)
	acct.Password = encryptedPW
	if err := a.accountRepo.Update(ctx, &acct); err != nil {
		return err
	}
	return nil
}

// Delete will delete the intended account
func (a *Account) Delete(ctx context.Context, id uuid.UUID, passWord string) error {
	var acct api.Account
	acct.ID = id
	if err := a.accountRepo.Find(ctx, &acct); err != nil {
		return err
	}

	if !a.secSvc.VerifyPasswordMatches(ctx, acct.Password, passWord) {
		return errors.New("password incorrect")
	}

	if err := a.accountRepo.Delete(ctx, &acct); err != nil {
		return err
	}
	return nil
}

func (a *Account) RefreshAuthorization(ctx context.Context) error {
	userN := ctx.Value("username")
	var acct api.Account
	acct.Username = userN.(string)
	if err := a.accountRepo.Find(ctx, &acct); err != nil {
		return err
	}

	token, err := a.authSvc.GenerateToken(ctx, &acct)
	if err == nil {
		return err
	}
	acct.Token = token
	if err := a.accountRepo.Update(ctx, &acct); err != nil {
		return err
	}
	return errors.New("unauthorized; please re-login")
}

func (a *Account) List(ctx context.Context) (*[]api.Account, error) {
	var accounts []api.Account
	if err := a.accountRepo.All(ctx, &accounts); err != nil {
		return &accounts, err
	}
	return &accounts, nil
}

func (a *Account) Update(ctx context.Context, request UpdateAccountRequest) error {
	var account api.Account
	account.ID = request.ID
	if err := a.accountRepo.Find(ctx, &account); err != nil {
		return err
	}
	if err := a.accountRepo.Update(ctx, &account); err != nil {
		return err
	}
	var updateUser api.User
	updateUser.ID = account.UserID
	if err := a.userRepo.View(ctx, &updateUser); err != nil {
		return err
	}
	updateUser.Address1 = request.Address1
	if request.Address2 != nil {
		updateUser.Address2 = *request.Address2
	}
	updateUser.ZipCode = request.ZipCode
	updateUser.Email = request.Email
	if err := a.userRepo.Update(ctx, &updateUser); err != nil {
		return err
	}
	return nil
}
