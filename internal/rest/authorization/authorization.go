package authorization

import (
	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin"
	"github.com/ispras/michman/internal/auth"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/rest/response"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

const (
	admin                 = "admin"
	user                  = "user"
	projectMember         = "project_member"
	authSuccessMessage    = "Authentication success! You are a member of some groups"
	authNoneGroupsMessage = "Authentication success! You aren't member of any group, have readonly rights"
)

type RespData struct {
	Message string `json:"message"`
	Groups  string `json:"groups"`
	UserID  string `json:"user_id"`
}

type AuthorizeClient struct {
	Logger         *logrus.Logger
	Db             database.Database
	Config         utils.Config
	SessionManager *scs.SessionManager
	Auth           auth.Authenticate
	Router         *httprouter.Router
}

// isProjectPath checks if there is a project in the request path
func isProjectPath(path string) bool {
	projectPath := regexp.MustCompile(utils.ProjectPathPattern).MatchString
	if projectPath(path) {
		return true
	}
	return false
}

func getProjectIdOrName(urlPath string) (string, error) {
	urlKeys := strings.Split(urlPath, "/")

	// if length of urlKeys less than 2 -- error
	if len(urlKeys) < 2 {
		return "", ErrNoProjectInURL
	}

	return urlKeys[2], nil
}

// getUserGroups returns a string array of user groups that the user is a member of
func (auth *AuthorizeClient) getUserGroups(r *http.Request, groupKey string) []string {
	groups := auth.SessionManager.GetString(r.Context(), groupKey)
	if groups == "" {
		return nil
	}

	groupsList := strings.Split(groups, ",")
	return groupsList
}

// Authorizer checks whether the user is authorized and allowed to access the requested resource or method
func (auth *AuthorizeClient) Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			request := r.Method + " " + r.URL.Path

			// add userID variable to the request header
			userId := auth.SessionManager.GetString(r.Context(), utils.UserIdKey)
			r.Header.Add(utils.UserIdKey, userId)

			// var for casbin role, set as user because user is default role
			role := user

			groups := auth.getUserGroups(r, utils.GroupKey)
			// check if user is a project member
			// if groups are nil -- role is user
			if groups != nil {
				if isProjectPath(r.URL.Path) {
					// get project which user wants to access
					projectIdOrName, err := getProjectIdOrName(r.URL.Path)
					if err != nil {
						auth.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
						response.Error(w, err)
						return
					}

					project, err := auth.Db.ReadProject(projectIdOrName)
					if err != nil {
						auth.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
						response.Error(w, err)
						return
					}

					// check if one of user groups presents in project
					for _, group := range groups {
						if group == project.GroupID {
							// user is project member
							role = projectMember
							break
						}
					}
				} else {
					// check if user is admin -- admin group must present in groups list
					for _, group := range groups {
						if group == auth.Config.AdminGroup {
							// user is admin
							role = admin
							break
						}
					}
				}
			}

			// casbin enforcer
			res, err := e.EnforceSafe(role, r.URL.Path, r.Method)
			if err != nil {
				auth.Logger.Warn("Request ", request, " failed with an error: ", ErrEnforcerSafe.Error())
				response.Error(w, ErrEnforcerSafe)
				return
			}
			if res {
				next.ServeHTTP(w, r)
			} else {
				auth.Logger.Info(request)
				auth.Logger.Warn("Request ", request, " failed with an error: ", ErrUnauthorized.Error())
				response.Error(w, ErrUnauthorized)
				return
			}
		}

		return http.HandlerFunc(fn)
	}
}

// AuthGet processes a request to authenticate client and show user groups
func (auth *AuthorizeClient) AuthGet(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	request := "GET /auth"
	auth.Logger.Info(request)

	// set auth facts
	err := auth.Auth.SetAuth(auth.SessionManager, r)
	if err != nil {
		auth.Logger.Warn("Request ", request, " failed with an error: ", err.Error())
		response.Error(w, err)
		return
	}

	// get user groups(roles) from request response body
	groups := auth.SessionManager.GetString(r.Context(), utils.GroupKey)

	// get userID from request response body
	userId := auth.SessionManager.GetString(r.Context(), utils.UserIdKey)

	if groups == "" {
		auth.Logger.Info(authNoneGroupsMessage)
		response.Ok(w, RespData{authNoneGroupsMessage, "none", userId}, request)
	} else {
		auth.Logger.Info("Authentication success!")
		auth.Logger.Info("----User groups are: " + groups)
		response.Ok(w, RespData{authSuccessMessage, groups, userId}, request)
	}
}
