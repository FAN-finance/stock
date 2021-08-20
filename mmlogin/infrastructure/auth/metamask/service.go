package metamask

import (
	"log"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"stock/mmlogin/domain"
	"stock/mmlogin/domain/auth"
	"stock/mmlogin/library/strutil"
)

const (
	challengeStringLength = 32
)

type service struct {
	secret              string
	tokenExpiryDuration time.Duration
}


func NewServiceWithOutToken() auth.Service {
	return &service{
	}
}

func NewService(secret string, ted time.Duration) auth.Service {
	return &service{
		secret:              secret,
		tokenExpiryDuration: ted,
	}
}

func (s *service) SetUpChallenge(u *domain.User) error {
	u.Challenge = strutil.Rand(challengeStringLength)
	return nil
}

func (s *service) VerifyResponse(u *domain.User, responseBytes []byte) error {
	if responseBytes[domain.SignatureSize-1] >= domain.SignatureRIRangeBase {
		responseBytes[domain.SignatureSize-1] -= domain.SignatureRIRangeBase
	}

	pubkey, err := crypto.SigToPub(
		challenge(u.Challenge).signatureHashBytes(),
		responseBytes,
	)
	if err != nil {
		return err
	}

	address := domain.Address(crypto.PubkeyToAddress(*pubkey))
	if address.Hex() != u.Address.Hex() {
		return domain.ErrInvalidSignature
	}

	return nil
}

func (s *service) IssueToken(u *domain.User) ([]byte, error) {
	if s.secret==""{
		log.Fatal("mmlogin miss secret")
	}
	return newToken(u.Address, s.tokenExpiryDuration).signedBytes(s.secret)
}
