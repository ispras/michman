package authorization

import (
	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin"
	"github.com/ispras/michman/internal/auth"
	"github.com/ispras/michman/internal/database"
	"github.com/ispras/michman/internal/rest/handler/response"
	"github.com/ispras/michman/internal/utils"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"net/http"
	"regexp"
	"strings"
)

const (
	admin         = "admin"
	user          = "user"
	projectMember = "project_member"
)

type AuthorizeClient struct {
	Logger         *logrus.Logger
	Db             database.Database
	Config         utils.Config
	SessionManager *scs.SessionManager
	Auth           auth.Authenticate
	Router         *httprouter.Router
}

func isProjectPath(path string) bool {
	projectPath := regexp.MustCompile(utils.ProjectPathPattern).MatchString
	if projectPath(path) {
		return true
	}
	return false
}

func getProjectIdOrName(urlPath string) (string, error) {
	urlKeys := strings.Split(urlPath, "/")

	//if length of urlKeys less then 2 -- error
	if len(urlKeys) < 2 {
		return "", ErrNoProjectInURL
	}

	return urlKeys[2], nil
}

func (auth *AuthorizeClient) getUserGroups(r *http.Request, groupKey string) []string {
	groups := auth.SessionManager.GetString(r.Context(), groupKey)
	if groups == "" {
		return nil
	}

	groupsList := strings.Split(groups, ",")
	return groupsList
}

func (auth *AuthorizeClient) Authorizer(e *casbin.Enforcer) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			request := r.Method + " " + r.URL.Path

			//var for casbin role, set as user because user is default role
			role := user

			groups := auth.getUserGroups(r, utils.GroupKey)
			//check if user is a project member
			//if groups are nil -- role is user
			if groups != nil {
				if isProjectPath(r.URL.Path) {
					//get project which user wants to access
					projectIdOrName, err := getProjectIdOrName(r.URL.Path)
					if err != nil {
						auth.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
						response.InternalError(w, err)
						return
					}

					project, err := auth.Db.ReadProject(projectIdOrName)
					if err != nil {
						auth.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
						response.InternalError(w, err)
						return
					}

					//check if one of user groups presents in project
					projectMemberFlag := false
					for _, g := range groups {
						if g == project.GroupID {
							projectMemberFlag = true
							break
						}
					}

					//user is project member
					if projectMemberFlag {
						role = projectMember
					}
				} else {
					//check if user is admin -- admin group must present in groups list
					adminFlag := false
					for _, g := range groups {
						if g == auth.Config.AdminGroup {
							adminFlag = true
							break
						}
					}

					//user is admin
					if adminFlag {
						role = admin
					}
				}
			}

			// casbin enforcer
			res, err := e.EnforceSafe(role, r.URL.Path, r.Method)
			if err != nil {
				auth.Logger.Warn("Request ", request, " failed with status ", http.StatusInternalServerError, ": ", err.Error())
				response.InternalError(w, err)
				return
			}
			if res {
				next.ServeHTTP(w, r)
			} else {
				auth.Logger.Warn("Request ", request, " failed with status ", http.StatusForbidden, ": ", ErrUnauthorized.Error())
				response.Forbidden(w, ErrUnauthorized)
				return
			}
		}

		return http.HandlerFunc(fn)
	}
}

func (auth *AuthorizeClient) AuthGet(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	request := "GET /auth"
	auth.Logger.Info(request)

	//set auth facts
	err, status := auth.Auth.SetAuth(auth.SessionManager, r)
	if err != nil {
		auth.Logger.Warn("Request ", request, " failed with status ", status, ": ", err.Error())
		switch status {
		case http.StatusBadRequest:
			response.BadRequest(w, err)
			return
		case http.StatusInternalServerError:
			response.InternalError(w, err)
			return
		}
	}

	g := auth.SessionManager.GetString(r.Context(), utils.GroupKey)

	auth.Logger.Info("Authentication success!")
	auth.Logger.Info("----User groups are: " + g)

	var userGroups string
	if g == "" {
		userGroups = "You are not a member of any group."
	} else {
		userGroups = "You are a member of the following groups: " + g
	}

	message := "Authentication success! " + userGroups
	response.Ok(w, message, request)
}
