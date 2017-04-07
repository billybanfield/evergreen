package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/evergreen-ci/evergreen"
	"github.com/evergreen-ci/evergreen/model/patch"
	"github.com/evergreen-ci/evergreen/model/user"
	"github.com/evergreen-ci/evergreen/model/version"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (uis *UIServer) timelineJson(w http.ResponseWriter, r *http.Request) {
	projCtx := MustHaveProjectContext(r)

	skip, perPage := getSkipAndLimit(r, DefaultSkip, DefaultLimit)
	data, err := getTimelineData(projCtx.Project.Identifier, evergreen.RepotrackerVersionRequester, skip, perPage)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting timeline data: %v", err.Error()), http.StatusInternalServerError)
		return
	}

	uis.WriteJSON(w, http.StatusOK, data)
}

func (uis *UIServer) timeline(w http.ResponseWriter, r *http.Request) {
	projCtx := MustHaveProjectContext(r)
	if projCtx.Project == nil {
		uis.ProjectNotFound(projCtx, w, r)
		return
	}
	uis.WriteHTML(w, http.StatusOK, struct {
		ProjectData projectContext
		User        *user.DBUser
	}{projCtx, GetUser(r)}, "base", "timeline.html", "base_angular.html", "menu.html")
}

func (uis *UIServer) patchTimeline(w http.ResponseWriter, r *http.Request) {
	uis.patchTimelineWrapper("", w, r)
}

func (uis *UIServer) myPatchesTimeline(w http.ResponseWriter, r *http.Request) {
	author := MustHaveUser(r).Username()
	uis.patchTimelineWrapper(author, w, r)
}

func (uis *UIServer) userPatchesTimeline(w http.ResponseWriter, r *http.Request) {
	author := mux.Vars(r)["user_id"]
	uis.patchTimelineWrapper(author, w, r)
}

func (uis *UIServer) patchTimelineWrapper(author string, w http.ResponseWriter, r *http.Request) {
	projCtx := MustHaveProjectContext(r)

	if projCtx.Project == nil {
		uis.ProjectNotFound(projCtx, w, r)
		return
	}

	uis.WriteHTML(w, http.StatusOK, struct {
		ProjectData projectContext
		User        *user.DBUser
		Author      string
	}{projCtx, GetUser(r), author}, "base", "patches.html", "base_angular.html", "menu.html")
}

func (uis *UIServer) patchTimelineJson(w http.ResponseWriter, r *http.Request) {
	projCtx := MustHaveProjectContext(r)

	pageNum, err := strconv.Atoi(r.FormValue("page"))
	if err != nil {
		pageNum = 0
	}
	skip := pageNum * DefaultLimit

	user := mux.Vars(r)["user_id"]
	var patches []patch.Patch
	if len(user) > 0 {
		patches, err = patch.Find(patch.ByUser(user).
			Project(patch.ExcludePatchDiff).
			Sort([]string{"-" + patch.CreateTimeKey}).
			Skip(skip).Limit(DefaultLimit))
	} else {
		patches, err = patch.Find(patch.ByProject(projCtx.Project.Identifier).
			Sort([]string{"-" + patch.CreateTimeKey}).
			Project(patch.ExcludePatchDiff).
			Skip(skip).Limit(DefaultLimit))
	}
	if err != nil {
		uis.LoggedError(w, r, http.StatusInternalServerError, errors.Wrapf(err,
			"Error fetching patches for %v", projCtx.Project.Identifier))
		return
	}

	versionIds := make([]string, 0, len(patches))
	uiPatches := make([]uiPatch, 0, len(patches))
	for _, patch := range patches {
		if patch.Version != "" {
			versionIds = append(versionIds, patch.Version)
		}
		baseVersion, err := version.FindOne(version.ByProjectIdAndRevision(patch.Project, patch.Githash))
		if err != nil {
			uis.LoggedError(w, r, http.StatusInternalServerError, err)
			return
		}
		var baseVersionId string
		if baseVersion != nil {
			baseVersionId = baseVersion.Id
		}
		patch.Patches = nil
		uiPatches = append(uiPatches, uiPatch{Patch: patch, BaseVersionId: baseVersionId})
	}
	versions, err := version.Find(version.ByIds(versionIds).WithoutFields(version.ConfigKey))
	if err != nil {
		uis.LoggedError(w, r, http.StatusInternalServerError, errors.Wrap(err, "Error fetching versions for patches"))
		return
	}

	versionsMap := map[string]*uiVersion{}
	for _, version := range versions {
		versionUI, err := PopulateUIVersion(&version)
		if err != nil {
			uis.LoggedError(w, r, http.StatusInternalServerError, err)
			return
		}
		versionsMap[version.Id] = versionUI
	}

	data := struct {
		VersionsMap map[string]*uiVersion
		UIPatches   []uiPatch
		PageNum     int
	}{versionsMap, uiPatches, pageNum}

	uis.WriteJSON(w, http.StatusOK, data)
}
