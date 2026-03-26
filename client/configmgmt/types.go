package configmgmt

import (
	"errors"
	"time"

	"github.com/fatedier/frp/client/http/model"
	"github.com/fatedier/frp/client/proxy"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
	ErrStoreDisabled   = errors.New("store disabled")
	ErrApplyConfig     = errors.New("apply config failed")
)

type ConfigManager interface {
	ReloadFromFile(strict bool) error

	ReadConfigFile() (string, error)
	WriteConfigFile(content []byte) error
	GetSettings() (model.ClientSettings, error)
	UpdateSettings(model.ClientSettings) error
	UploadFile(targetPath string, filename string, content []byte) (string, error)

	GetProxyStatus() []*proxy.WorkingStatus
	IsStoreProxyEnabled(name string) bool
	StoreEnabled() bool

	GetProxyConfig(name string) (v1.ProxyConfigurer, bool)
	GetVisitorConfig(name string) (v1.VisitorConfigurer, bool)
	ListConfigProxies() ([]v1.ProxyConfigurer, error)
	ListConfigVisitors() ([]v1.VisitorConfigurer, error)

	ListStoreProxies() ([]v1.ProxyConfigurer, error)
	GetStoreProxy(name string) (v1.ProxyConfigurer, error)
	CreateStoreProxy(cfg v1.ProxyConfigurer) (v1.ProxyConfigurer, error)
	UpdateStoreProxy(name string, cfg v1.ProxyConfigurer) (v1.ProxyConfigurer, error)
	DeleteStoreProxy(name string) error

	ListStoreVisitors() ([]v1.VisitorConfigurer, error)
	GetStoreVisitor(name string) (v1.VisitorConfigurer, error)
	CreateStoreVisitor(cfg v1.VisitorConfigurer) (v1.VisitorConfigurer, error)
	UpdateStoreVisitor(name string, cfg v1.VisitorConfigurer) (v1.VisitorConfigurer, error)
	DeleteStoreVisitor(name string) error

	GracefulClose(d time.Duration)
}
