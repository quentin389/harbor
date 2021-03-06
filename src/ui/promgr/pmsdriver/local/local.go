// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package local

import (
	"fmt"
	"regexp"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	errutil "github.com/vmware/harbor/src/common/utils/error"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/promgr/pmsdriver"
)

const dupProjectPattern = `Duplicate entry '\w+' for key 'name'`

type driver struct {
}

// NewDriver returns an instance of driver
func NewDriver() pmsdriver.PMSDriver {
	return &driver{}
}

// Get ...
func (d *driver) Get(projectIDOrName interface{}) (
	*models.Project, error) {
	id, name, err := utils.ParseProjectIDOrName(projectIDOrName)
	if err != nil {
		return nil, err
	}

	if id > 0 {
		return dao.GetProjectByID(id)
	}

	return dao.GetProjectByName(name)
}

// Create ...
func (d *driver) Create(project *models.Project) (int64, error) {
	if project == nil {
		return 0, fmt.Errorf("project is nil")
	}

	if len(project.Name) == 0 {
		return 0, fmt.Errorf("project name is nil")
	}

	if project.OwnerID == 0 {
		if len(project.OwnerName) == 0 {
			return 0, fmt.Errorf("owner ID and owner name are both nil")
		}

		user, err := dao.GetUser(models.User{
			Username: project.OwnerName,
		})
		if err != nil {
			return 0, err
		}
		if user == nil {
			return 0, fmt.Errorf("can not get owner whose name is %s", project.OwnerName)
		}
		project.OwnerID = user.UserID
	}

	t := time.Now()
	pro := &models.Project{
		Name:         project.Name,
		Public:       project.Public,
		OwnerID:      project.OwnerID,
		CreationTime: t,
		UpdateTime:   t,
	}

	id, err := dao.AddProject(*pro)
	if err != nil {
		dup, e := regexp.MatchString(dupProjectPattern, err.Error())
		if e != nil {
			log.Errorf("failed to match duplicate project pattern: %v", e)
		}

		if dup {
			err = errutil.ErrDupProject
		}

		return 0, err
	}

	return id, nil
}

// Delete ...
func (d *driver) Delete(projectIDOrName interface{}) error {
	id, name, err := utils.ParseProjectIDOrName(projectIDOrName)
	if err != nil {
		return err
	}
	if len(name) > 0 {
		project, err := dao.GetProjectByName(name)
		if err != nil {
			return err
		}
		id = project.ProjectID
	}

	return dao.DeleteProject(id)
}

// Update ...
func (d *driver) Update(projectIDOrName interface{},
	project *models.Project) error {
	// nil implement
	return nil
}

// TODO remove base
// List returns a project list according to the query parameters
func (d *driver) List(query *models.ProjectQueryParam,
	base ...*models.BaseProjectCollection) (
	*models.ProjectQueryResult, error) {
	total, err := dao.GetTotalOfProjects(query, base...)
	if err != nil {
		return nil, err
	}
	projects, err := dao.GetProjects(query, base...)
	if err != nil {
		return nil, err
	}
	return &models.ProjectQueryResult{
		Total:    total,
		Projects: projects,
	}, nil
}

func (d *driver) EnableExternalMetaMgr() bool {
	return true
}
