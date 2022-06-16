package authorization

func (auth *AuthorizeClient) CreateRoutes() {
	auth.Router.Handle("GET", "/auth", auth.AuthGet)
}
