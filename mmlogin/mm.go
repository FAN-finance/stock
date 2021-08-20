package mmlogin

import (
	"stock/mmlogin/application"
	"stock/mmlogin/infrastructure/auth/metamask"
	cacheUser "stock/mmlogin/infrastructure/cache/user"
	appAuth "stock/mmlogin/application/auth"
	appUser "stock/mmlogin/application/user"
)

func newAppCore(conf *application.Config) *application.Core {
	return &application.Core{
		Services: &application.Services{
			Auth: metamask.NewServiceWithOutToken(
			),
		},
		Repositories: &application.Repositories{
			User: cacheUser.NewRepository(),
		},
	}
}
type apps struct {
	Auth appAuth.Application
	User appUser.Application
}
var Apps *apps
func InitMMLogin(){
	appCore := newAppCore(nil)
	Apps=&apps{
		appAuth.NewApplication(appCore),
		appUser.NewApplication(appCore),
	}
}

