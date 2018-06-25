package addons

import (
	"encoding/base64"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/jim-minter/azure-helm/pkg/jsonpath"
)

func KeyFunc(gk schema.GroupKind, namespace, name string) string {
	s := gk.String()
	if namespace != "" {
		s += "/" + namespace
	}
	s += "/" + name

	return s
}

type NestedFlags int

const (
	NestedFlagsBase64 NestedFlags = (1 << iota)
)

func Translate(o interface{}, path jsonpath.Path, nestedPath jsonpath.Path, nestedFlags NestedFlags, v string) error {
	var err error

	if nestedPath == nil {
		path.Set(o, v)
		return nil
	}

	nestedBytes := []byte(path.MustGetString(o))

	if nestedFlags&NestedFlagsBase64 != 0 {
		nestedBytes, err = base64.StdEncoding.DecodeString(string(nestedBytes))
		if err != nil {
			return err
		}
	}

	var nestedObject interface{}
	err = yaml.Unmarshal(nestedBytes, &nestedObject)
	if err != nil {
		panic(err)
	}

	nestedPath.Set(nestedObject, v)

	nestedBytes, err = yaml.Marshal(nestedObject)
	if err != nil {
		panic(err)
	}

	if nestedFlags&NestedFlagsBase64 != 0 {
		nestedBytes = []byte(base64.StdEncoding.EncodeToString(nestedBytes))
		if err != nil {
			panic(err)
		}
	}

	path.Set(o, string(nestedBytes))

	return nil
}

var Translations = map[string][]struct {
	Path        jsonpath.Path
	NestedPath  jsonpath.Path
	NestedFlags NestedFlags
	Template    string
}{
	"APIService.apiregistration.k8s.io/v1beta1.servicecatalog.k8s.io": {
		{
			Path:     jsonpath.MustCompile("$.spec.caBundle"),
			Template: "{{ Base64Encode (CertAsBytes .Config.ServiceCatalogCaCert) }}",
		},
	},
	"ClusterServiceBroker.servicecatalog.k8s.io/ansible-service-broker": {
		{
			Path:     jsonpath.MustCompile("$.spec.caBundle"),
			Template: "{{ Base64Encode (CertAsBytes .Config.ServiceSigningCaCert) }}",
		},
	},
	"ClusterServiceBroker.servicecatalog.k8s.io/template-service-broker": {
		{
			Path:     jsonpath.MustCompile("$.spec.caBundle"),
			Template: "{{ Base64Encode (CertAsBytes .Config.ServiceSigningCaCert) }}",
		},
	},
	"ConfigMap/kube-service-catalog/cluster-info": {
		{
			Path:     jsonpath.MustCompile("$.data.id"),
			Template: "{{ .Config.ServiceCatalogClusterID }}",
		},
	},
	"ConfigMap/kube-system/extension-apiserver-authentication": {
		{
			Path:     jsonpath.MustCompile("$.data.'client-ca-file'"),
			Template: "{{ String (CertAsBytes .Config.CaCert) }}",
		},
		{
			Path:     jsonpath.MustCompile("$.data.'requestheader-client-ca-file'"),
			Template: "{{ String (CertAsBytes .Config.FrontProxyCaCert) }}",
		},
	},
	"ConfigMap/openshift-web-console/webconsole-config": {
		{
			Path:       jsonpath.MustCompile("$.data.'webconsole-config.yaml'"),
			NestedPath: jsonpath.MustCompile("$.clusterInfo.consolePublicURL"),
			Template:   "https://{{ .Config.PublicHostname }}/console/",
		},
		{
			Path:       jsonpath.MustCompile("$.data.'webconsole-config.yaml'"),
			NestedPath: jsonpath.MustCompile("$.clusterInfo.masterPublicURL"),
			Template:   "https://{{ .Config.PublicHostname }}.cloudapp.azure.com",
		},
	},
	"DaemonSet.apps/kube-service-catalog/apiserver": {
		{
			Path:     jsonpath.MustCompile("$.spec.template.spec.containers[0].args[6]"),
			Template: "https://{{ .Config.EtcdHostname }}:2379",
		},
	},
	"DeploymentConfig.apps.openshift.io/default/docker-registry": {
		{
			Path:     jsonpath.MustCompile("$.spec.template.spec.containers[0].env[?(@.name='REGISTRY_HTTP_SECRET')].value"),
			Template: "{{ Base64Encode .Config.RegistryHTTPSecret }}",
		},
	},
	"DeploymentConfig.apps.openshift.io/default/registry-console": {
		{
			Path:     jsonpath.MustCompile("$.spec.template.spec.containers[0].env[?(@.name='OPENSHIFT_OAUTH_PROVIDER_URL')].value"),
			Template: "https://{{ .Config.PublicHostname }}.cloudapp.azure.com",
		},
		{
			Path:     jsonpath.MustCompile("$.spec.template.spec.containers[0].env[?(@.name='REGISTRY_HOST')].value"),
			Template: "docker-registry-default.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"OAuthClient.oauth.openshift.io/cockpit-oauth-client": {
		{
			Path:     jsonpath.MustCompile("$.redirectURIs[0]"),
			Template: "https://registry-console-default.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/default/docker-registry": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "docker-registry-default.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/default/registry-console": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "registry-console-default.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/kube-service-catalog/apiserver": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "apiserver-kube-service-catalog.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/openshift-ansible-service-broker/asb-1338": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "asb-1338-openshift-ansible-service-broker.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/openshift-metrics/alertmanager": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "alertmanager-openshift-metrics.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/openshift-metrics/alerts": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "alerts-openshift-metrics.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Route.route.openshift.io/openshift-metrics/prometheus": {
		{
			Path:     jsonpath.MustCompile("$.spec.host"),
			Template: "prometheus-openshift-metrics.{{ .Config.RouterIP }}.nip.io",
		},
	},
	"Secret/default/registry-certificates": {
		{
			Path:     jsonpath.MustCompile("$.data.'registry.crt'"),
			Template: "{{ Base64Encode (JoinBytes (CertAsBytes .Config.RegistryCert) (CertAsBytes .Config.CaCert)) }}",
		},
		{
			Path:     jsonpath.MustCompile("$.data.'registry.key'"),
			Template: "{{ Base64Encode (PrivateKeyAsBytes .Config.RegistryKey) }}",
		},
	},
	"Secret/default/registry-config": {
		{
			Path:        jsonpath.MustCompile("$.data.'config.yml'"),
			NestedPath:  jsonpath.MustCompile("$.storage.azure.accountname"),
			NestedFlags: NestedFlagsBase64,
			Template:    "{{ .Config.RegistryStorageAccount }}",
		},
		{
			Path:        jsonpath.MustCompile("$.data.'config.yml'"),
			NestedPath:  jsonpath.MustCompile("$.storage.azure.accountkey"),
			NestedFlags: NestedFlagsBase64,
			Template:    "{{ .Config.RegistryAccountKey }}",
		},
	},
	"Secret/default/router-certs": {
		{
			Path:     jsonpath.MustCompile("$.data.'tls.crt'"),
			Template: "{{ Base64Encode (JoinBytes (CertAsBytes .Config.RouterCert) (CertAsBytes .Config.CaCert) (PrivateKeyAsBytes .Config.RouterKey)) }}",
		},
		{
			Path:     jsonpath.MustCompile("$.data.'tls.key'"),
			Template: "{{ Base64Encode (PrivateKeyAsBytes .Config.RouterKey) }}",
		},
	},
	"Secret/kube-service-catalog/apiserver-ssl": {
		{
			Path:     jsonpath.MustCompile("$.data.'tls.crt'"),
			Template: "{{ Base64Encode (JoinBytes (CertAsBytes .Config.ServiceCatalogServerCert) (CertAsBytes .Config.ServiceCatalogCaCert)) }}",
		},
		{
			Path:     jsonpath.MustCompile("$.data.'tls.key'"),
			Template: "{{ Base64Encode (PrivateKeyAsBytes .Config.ServiceCatalogServerKey) }}",
		},
	},
	"Secret/openshift-metrics/alertmanager-proxy": {
		{
			Path:     jsonpath.MustCompile("$.data.'session_secret'"),
			Template: "{{ Base64Encode (Bytes (Base64Encode .Config.AlertManagerProxySessionSecret)) }}",
		},
	},
	"Secret/openshift-metrics/alerts-proxy": {
		{
			Path:     jsonpath.MustCompile("$.data.'session_secret'"),
			Template: "{{ Base64Encode (Bytes (Base64Encode .Config.AlertsProxySessionSecret)) }}",
		},
	},
	"Secret/openshift-metrics/prometheus-proxy": {
		{
			Path:     jsonpath.MustCompile("$.data.'session_secret'"),
			Template: "{{ Base64Encode (Bytes (Base64Encode .Config.PrometheusProxySessionSecret)) }}",
		},
	},
	"Service/default/docker-registry": {
		{
			Path:     jsonpath.MustCompile("$.spec.clusterIP"),
			Template: "{{ .Config.RegistryServiceIP }}",
		},
	},
}
