package api

import (
	"github.com/reeceappling/goUtils/v2/utils"
	sliceutils "github.com/reeceappling/goUtils/v2/utils/slices"
	"golang.org/x/exp/maps"
	"net/http"
)

type Permissioned interface {
	Permissions() ACL
}

var SessionUserProjectsHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	complete := queryParams.Get("complete")
	var getAllProjectsCompleteArg *bool = nil
	if complete != "" {
		if complete == "true" {
			getAllProjectsCompleteArg = utils.Pointer(true)
		} else if complete == "false" {
			getAllProjectsCompleteArg = utils.Pointer(false)
		}
	}

	ctx := r.Context()
	user, err := GetAuthInfo(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	var projectsToReturn []projectName
	if user.IsAdmin() {
		allProjects, err := GetAllProjects(ctx, getAllProjectsCompleteArg)
		if err != nil {
			http.Error(w, "failed to get all incomplete projects: "+err.Error(), http.StatusInternalServerError)
			return
		}
		projectsToReturn = sliceutils.Map(allProjects, func(proj Project) projectName {
			return proj.Name
		})
	} else {
		if user.Projects != nil {
			projectsToReturn = maps.Keys(user.Projects)
		} else {
			projectsToReturn = []projectName{}
		}

	}
	MarshalAndReturn(ctx, w, projectsToReturn)
}
