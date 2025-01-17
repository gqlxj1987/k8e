package templates

import (
	"bytes"
	"text/template"

	"github.com/xiaods/k8e/pkg/daemons/config"
)

type ContainerdConfig struct {
	NodeConfig            *config.Node
	IsRunningInUserNS     bool
	PrivateRegistryConfig *Registry
}

const ContainerdConfigTemplate = `
[plugins.opt]
  path = "{{ .NodeConfig.Containerd.Opt }}"

[plugins.cri]
  stream_server_address = "127.0.0.1"
  stream_server_port = "10010"
  enable_selinux = {{ .NodeConfig.SELinux }}

{{- if .IsRunningInUserNS }}
  disable_cgroup = true
  disable_apparmor = true
  restrict_oom_score_adj = true
{{end}}

{{- if .NodeConfig.AgentConfig.PauseImage }}
  sandbox_image = "{{ .NodeConfig.AgentConfig.PauseImage }}"
{{end}}

{{- if .NodeConfig.AgentConfig.Snapshotter }}
[plugins.cri.containerd]
  snapshotter = "{{ .NodeConfig.AgentConfig.Snapshotter }}"
{{end}}

{{- if not .NodeConfig.NoFlannel }}
[plugins.cri.cni]
  bin_dir = "{{ .NodeConfig.AgentConfig.CNIBinDir }}"
  conf_dir = "{{ .NodeConfig.AgentConfig.CNIConfDir }}"
{{end}}

[plugins.cri.containerd.runtimes.runc]
  runtime_type = "io.containerd.runc.v2"

{{ if .PrivateRegistryConfig }}
{{ if .PrivateRegistryConfig.Mirrors }}
[plugins.cri.registry.mirrors]{{end}}
{{range $k, $v := .PrivateRegistryConfig.Mirrors }}
[plugins.cri.registry.mirrors."{{$k}}"]
  endpoint = [{{range $i, $j := $v.Endpoints}}{{if $i}}, {{end}}{{printf "%q" .}}{{end}}]
{{end}}

{{range $k, $v := .PrivateRegistryConfig.Configs }}
{{ if $v.Auth }}
[plugins.cri.registry.configs."{{$k}}".auth]
  {{ if $v.Auth.Username }}username = {{ printf "%q" $v.Auth.Username }}{{end}}
  {{ if $v.Auth.Password }}password = {{ printf "%q" $v.Auth.Password }}{{end}}
  {{ if $v.Auth.Auth }}auth = {{ printf "%q" $v.Auth.Auth }}{{end}}
  {{ if $v.Auth.IdentityToken }}identitytoken = {{ printf "%q" $v.Auth.IdentityToken }}{{end}}
{{end}}
{{ if $v.TLS }}
[plugins.cri.registry.configs."{{$k}}".tls]
  {{ if $v.TLS.CAFile }}ca_file = "{{ $v.TLS.CAFile }}"{{end}}
  {{ if $v.TLS.CertFile }}cert_file = "{{ $v.TLS.CertFile }}"{{end}}
  {{ if $v.TLS.KeyFile }}key_file = "{{ $v.TLS.KeyFile }}"{{end}}
  {{ if $v.TLS.InsecureSkipVerify }}insecure_skip_verify = true{{end}}
{{end}}
{{end}}
{{end}}
`

func ParseTemplateFromConfig(templateBuffer string, config interface{}) (string, error) {
	out := new(bytes.Buffer)
	t := template.Must(template.New("compiled_template").Parse(templateBuffer))
	if err := t.Execute(out, config); err != nil {
		return "", err
	}
	return out.String(), nil
}
