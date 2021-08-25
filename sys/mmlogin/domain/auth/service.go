package auth

import "stock/sys/mmlogin/domain"

type Service interface {
	SetUpChallenge(u *domain.User) error
	VerifyResponse(u *domain.User, responseBytes []byte) error
	IssueToken(u *domain.User) ([]byte, error)
}
