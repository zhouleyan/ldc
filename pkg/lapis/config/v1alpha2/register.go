package v1alpha2

import (
	"github.com/emicklei/go-restful/v3"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"ldc.io/ldc/pkg/apiserver/runtime"

	ldcconfig "ldc.io/ldc/pkg/apiserver/config"
)

const (
	GroupName = "config.ldc.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, config *ldcconfig.Config) error {
	webservice := runtime.NewWebService(GroupVersion)

	webservice.Route(webservice.GET("/configs/configz").
		Doc("Information about the server configuration").
		To(func(request *restful.Request, response *restful.Response) {
			response.WriteAsJson(config.ToMap())
		}))

	c.Add(webservice)
	return nil
}
