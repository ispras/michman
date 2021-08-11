package authorization

import (
	"errors"
	"github.com/alexedwards/scs/v2"
	"github.com/casbin/casbin"
	"github.com/ispras/michman/internal/utils"
	"github.com/google/uuid"
	"github.com/ispras/michman/internal/database"
	proto "github.com/ispras/michman/internal/protobuf"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const (
	admin = "admin"
	user = "user"
	projectMember = "project_member"
)

type AuthorizeClient struct {
	Logger *log.Logger
	Db     database.Database
	Config utils.Config
	SessionManager *scs.SessionManager
}

func isProjectPath(path string) bool  {
	projectPath := regexp.MustCompile(`^/projects/`).MatchString
	if projectPath(path) {
		return true
	}
	return false
}

func getProjectIdOrName(urlPath string) (string, error) {
	urlKeys := strings.Split(urlPath, "/")

	//if length of urlKeys less then 2 -- error
	if len(urlKeys) < 2 {
		return "", errors.New("ERROR: no project ID or name in URL path")
	}

	return urlKeys[2], nil
}

func (auth *AuthorizeClient) getProject(idORname string) (*proto.Project, error) {
	is_uuid := true
	_, err := uuid.Parse(idORname)
	if err != nil {
		is_uuid = false
	}

	var project *proto.Project

	if is_uuid {
		project, err = auth.Db.ReadProject(idORname)
	} else {
		project, err = auth.Db.ReadProjectByName(idORname)
	}

	return project, err
}

func (auth *AuthorizeClient) getUserGroups(r *http.Request, groupKey string) []string{
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
						auth.Logger.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("ERROR"))
						return
					}

					project, err := auth.getProject(projectIdOrName)
					if err != nil {
						auth.Logger.Print(err)
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("ERROR"))
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
				log.Print("ERROR: ", err.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("ERROR"))
				return
			}
			if res {
				next.ServeHTTP(w, r)
			} else {
				log.Print("ERROR: ", "unauthorized")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("FORBIDDEN"))
				return
			}
		}

		return http.HandlerFunc(fn)
	}
}
