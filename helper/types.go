package helper

import "github.com/rohan031/adgytec-api/v1/services"

type Constraint interface {
	services.Newsletter | services.User | services.Project | services.ProjectServiceMap |
		services.ProjectUserMap | services.NewsDelete | services.NewsPut
}
