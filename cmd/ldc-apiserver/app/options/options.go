package options

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"strings"

	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
	"ldc.io/ldc/pkg/apiserver"
	apiserverconfig "ldc.io/ldc/pkg/apiserver/config"
	genericoptions "ldc.io/ldc/pkg/server/options"
	"ldc.io/ldc/pkg/simple/client/cache"
)

type ServerRunOptions struct {
	ConfigFile              string
	GenericServerRunOptions *genericoptions.ServerRunOptions
	*apiserverconfig.Config

	DebugMode bool
}

func NewServerRunOptions() *ServerRunOptions {
	s := &ServerRunOptions{
		GenericServerRunOptions: genericoptions.NewServerRunOptions(),
		Config:                  apiserverconfig.New(),
	}

	return s
}

func (s *ServerRunOptions) Flags() (fss cliflag.NamedFlagSets) {
	fs := fss.FlagSet("generic")
	fs.BoolVar(&s.DebugMode, "debug", false, "Don't enable this if you don't know what it means.")
	s.GenericServerRunOptions.AddFlags(fs, s.GenericServerRunOptions)
	s.RedisOptions.AddFlags(fss.FlagSet("redis"), s.RedisOptions)

	fs = fss.FlagSet("klog")
	local := flag.NewFlagSet("klog", flag.ExitOnError)
	klog.InitFlags(local)
	local.VisitAll(func(fl *flag.Flag) {
		fl.Name = strings.Replace(fl.Name, "_", "-", -1)
		fs.AddGoFlag(fl)
	})

	return fss
}

const fakeInterface string = "FAKE"

// NewAPIServer creates an APIServer instance using given options
func (s *ServerRunOptions) NewAPIServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{
		Config: s.Config,
	}
	var err error
	var cacheClient cache.Interface
	if s.RedisOptions != nil && len(s.RedisOptions.Host) != 0 {
		if s.RedisOptions.Host == fakeInterface && s.DebugMode {
			apiServer.CacheClient = cache.NewSimpleCache()
		} else {
			cacheClient, err = cache.NewRedisClient(s.RedisOptions, stopCh)
			if err != nil {
				return nil, fmt.Errorf("failed to connect to redis service, please check redis status, err: %v", err)
			}
			apiServer.CacheClient = cacheClient
		}
	} else {
		klog.Warning("ldc-apiserver starts without redis provided, it will use in memory cache. " +
			"This may cause inconsistencies when running ldc-apiserver with multiple replicas.")
		apiServer.CacheClient = cache.NewSimpleCache()
	}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.GenericServerRunOptions.InsecurePort),
	}

	if s.GenericServerRunOptions.SecurePort != 0 {
		certificate, err := tls.LoadX509KeyPair(s.GenericServerRunOptions.TLSCertFile, s.GenericServerRunOptions.TLSPrivateKey)
		if err != nil {
			return nil, err
		}
		server.TLSConfig.Certificates = []tls.Certificate{certificate}
	}

	apiServer.Server = server

	return apiServer, nil
}
