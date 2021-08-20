package application

import (
	"stock/mmlogin/domain/auth"
	"stock/mmlogin/domain/user"
)

type Core struct {
	Config       *Config
	Services     *Services
	Repositories *Repositories
}

type Services struct {
	Auth auth.Service
}

type Repositories struct {
	User user.Repository
}
