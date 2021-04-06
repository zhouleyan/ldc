package apiserver

import (
	"bytes"
	"fmt"
	"net/http"
	rt "runtime"
	"time"

	restful "github.com/emicklei/go-restful/v3"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog"
	apiserverconfig "ldc.io/ldc/pkg/apiserver/config"
	"ldc.io/ldc/pkg/simple/client/cache"
	utilnet "ldc.io/ldc/pkg/utils/net"

	configv1alpha2 "ldc.io/ldc/pkg/lapis/config/v1alpha2"
)

type APIServer struct {
	// number of ldc apiserver
	ServerCount int

	//
	Server *http.Server

	Config *apiserverconfig.Config

	// webservice container, where all webservice defines
	container *restful.Container

	// cache is used for short lived objects, like session
	CacheClient cache.Interface
}

func (s *APIServer) PrepareRun() error {
	s.container = restful.NewContainer()
	s.container.Filter(logRequestAndResponse)
	s.container.Router(restful.CurlyRouter{})
	s.container.RecoverHandler(func(panicReason interface{}, httpWriter http.ResponseWriter) {
		logStackOnRecover(panicReason, httpWriter)
	})

	s.installLDCAPIs()

	for _, ws := range s.container.RegisteredWebServices() {
		klog.V(2).Infof("%s", ws.RootPath())
	}

	s.Server.Handler = s.container

	// s.buildHandlerChain()

	return nil
}

func (s *APIServer) Run() (err error) {
	klog.V(0).Infof("Start listening on %s", s.Server.Addr)
	if s.Server.TLSConfig != nil {
		err = s.Server.ListenAndServeTLS("", "")
	} else {
		err = s.Server.ListenAndServe()
	}

	return err
}

func (s *APIServer) installLDCAPIs() {
	urlruntime.Must(configv1alpha2.AddToContainer(s.container, s.Config))
}

func logStackOnRecover(panicReason interface{}, w http.ResponseWriter) {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("recover from panic situation: - %v\r\n", panicReason))
	for i := 2; ; i += 1 {
		_, file, line, ok := rt.Caller(i)
		if !ok {
			break
		}
		buffer.WriteString(fmt.Sprintf("    %s:%d\r\n", file, line))
	}
	klog.Errorln(buffer.String())

	headers := http.Header{}
	if ct := w.Header().Get("Content-Type"); len(ct) > 0 {
		headers.Set("Accept", ct)
	}

	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal server error"))
}

func logRequestAndResponse(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	start := time.Now()
	chain.ProcessFilter(req, resp)

	// Always log error response
	logWithVerbose := klog.V(4)
	if resp.StatusCode() > http.StatusBadRequest {
		logWithVerbose = klog.V(0)
	}

	logWithVerbose.Infof("%s - \"%s %s %s\" %d %d %dms",
		utilnet.GetRequestIP(req.Request),
		req.Request.Method,
		req.Request.URL,
		req.Request.Proto,
		resp.StatusCode(),
		resp.ContentLength(),
		time.Since(start)/time.Millisecond,
	)
}

// type errorResponder struct{}

// func (e *errorResponder) Error(w http.ResponseWriter, req *http.Request, err error) {
// 	klog.Error(err)
// 	responsewriters.InternalError(w, req, err)
// }
