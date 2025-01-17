package cmds

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"github.com/xiaods/k8e/pkg/version"
)

type Agent struct {
	Token                    string
	TokenFile                string
	ClusterSecret            string
	ServerURL                string
	DisableLoadBalancer      bool
	ResolvConf               string
	DataDir                  string
	NodeIP                   string
	NodeExternalIP           string
	NodeName                 string
	PauseImage               string
	Snapshotter              string
	Docker                   bool
	ContainerRuntimeEndpoint string
	NoFlannel                bool
	FlannelIface             string
	FlannelConf              string
	Debug                    bool
	Rootless                 bool
	RootlessAlreadyUnshared  bool
	WithNodeID               bool
	EnableSELinux            bool
	ProtectKernelDefaults    bool
	AgentShared
	ExtraKubeletArgs   cli.StringSlice
	ExtraKubeProxyArgs cli.StringSlice
	Labels             cli.StringSlice
	Taints             cli.StringSlice
	PrivateRegistry    string
}

type AgentShared struct {
	NodeIP string
}

var (
	appName     = filepath.Base(os.Args[0])
	AgentConfig Agent
	NodeIPFlag  = cli.StringFlag{
		Name:        "node-ip,i",
		Usage:       "(agent/networking) IP address to advertise for node",
		Destination: &AgentConfig.NodeIP,
	}
	NodeExternalIPFlag = cli.StringFlag{
		Name:        "node-external-ip",
		Usage:       "(agent/networking) External IP address to advertise for node",
		Destination: &AgentConfig.NodeExternalIP,
	}
	NodeNameFlag = cli.StringFlag{
		Name:        "node-name",
		Usage:       "(agent/node) Node name",
		EnvVar:      version.ProgramUpper + "_NODE_NAME",
		Destination: &AgentConfig.NodeName,
	}
	WithNodeIDFlag = cli.BoolFlag{
		Name:        "with-node-id",
		Usage:       "(agent/node) Append id to node name",
		Destination: &AgentConfig.WithNodeID,
	}
	DockerFlag = cli.BoolFlag{
		Name:        "docker",
		Usage:       "(agent/runtime) Use docker instead of containerd",
		Destination: &AgentConfig.Docker,
	}
	CRIEndpointFlag = cli.StringFlag{
		Name:        "container-runtime-endpoint",
		Usage:       "(agent/runtime) Disable embedded containerd and use alternative CRI implementation",
		Destination: &AgentConfig.ContainerRuntimeEndpoint,
	}
	PrivateRegistryFlag = cli.StringFlag{
		Name:        "private-registry",
		Usage:       "(agent/runtime) Private registry configuration file",
		Destination: &AgentConfig.PrivateRegistry,
		Value:       "/etc/k8e/" + version.Program + "/registries.yaml",
	}
	PauseImageFlag = cli.StringFlag{
		Name:        "pause-image",
		Usage:       "(agent/runtime) Customized pause image for containerd or docker sandbox",
		Destination: &AgentConfig.PauseImage,
		Value:       "registry.cn-hangzhou.aliyuncs.com/google_containers/pause:3.2",
	}
	SnapshotterFlag = cli.StringFlag{
		Name:        "snapshotter",
		Usage:       "(agent/runtime) Override default containerd snapshotter",
		Destination: &AgentConfig.Snapshotter,
		Value:       "overlayfs",
	}
	FlannelFlag = cli.BoolFlag{
		Name:        "no-flannel",
		Usage:       "(deprecated) use --flannel-backend=none",
		Destination: &AgentConfig.NoFlannel,
	}
	FlannelIfaceFlag = cli.StringFlag{
		Name:        "flannel-iface",
		Usage:       "(agent/networking) Override default flannel interface",
		Destination: &AgentConfig.FlannelIface,
	}
	FlannelConfFlag = cli.StringFlag{
		Name:        "flannel-conf",
		Usage:       "(agent/networking) Override default flannel config file",
		Destination: &AgentConfig.FlannelConf,
	}
	ResolvConfFlag = cli.StringFlag{
		Name:        "resolv-conf",
		Usage:       "(agent/networking) Kubelet resolv.conf file",
		EnvVar:      version.ProgramUpper + "_RESOLV_CONF",
		Destination: &AgentConfig.ResolvConf,
	}
	ExtraKubeletArgs = cli.StringSliceFlag{
		Name:  "kubelet-arg",
		Usage: "(agent/flags) Customized flag for kubelet process",
		Value: &AgentConfig.ExtraKubeletArgs,
	}
	ExtraKubeProxyArgs = cli.StringSliceFlag{
		Name:  "kube-proxy-arg",
		Usage: "(agent/flags) Customized flag for kube-proxy process",
		Value: &AgentConfig.ExtraKubeProxyArgs,
	}
	NodeTaints = cli.StringSliceFlag{
		Name:  "node-taint",
		Usage: "(agent/node) Registering kubelet with set of taints",
		Value: &AgentConfig.Taints,
	}
	NodeLabels = cli.StringSliceFlag{
		Name:  "node-label",
		Usage: "(agent/node) Registering and starting kubelet with set of labels",
		Value: &AgentConfig.Labels,
	}
	DisableSELinuxFlag = cli.BoolTFlag{
		Name:   "disable-selinux",
		Usage:  "(deprecated) Use --selinux to explicitly enable SELinux",
		Hidden: true,
	}
	ProtectKernelDefaultsFlag = cli.BoolFlag{
		Name:        "protect-kernel-defaults",
		Usage:       "(agent/node) Kernel tuning behavior. If set, error if kernel tunables are different than kubelet defaults.",
		Destination: &AgentConfig.ProtectKernelDefaults,
	}
	SELinuxFlag = cli.BoolFlag{
		Name:        "selinux",
		Usage:       "(agent/node) Enable SELinux in containerd",
		Hidden:      false,
		Destination: &AgentConfig.EnableSELinux,
		EnvVar:      version.ProgramUpper + "_SELINUX",
	}
)

func CheckSELinuxFlags(ctx *cli.Context) error {
	disable, enable := DisableSELinuxFlag.Name, SELinuxFlag.Name
	switch {
	case ctx.IsSet(disable) && ctx.IsSet(enable):
		return errors.Errorf("--%s is deprecated in favor of --%s to affirmatively enable it in containerd", disable, enable)
	case ctx.IsSet(disable):
		AgentConfig.EnableSELinux = !ctx.Bool(disable)
	}
	return nil
}
func NewAgentCommand(action func(ctx *cli.Context) error) cli.Command {
	return cli.Command{
		Name:      "agent",
		Usage:     "Run node agent",
		UsageText: appName + " agent [OPTIONS]",
		Before:    SetupDebug(CheckSELinuxFlags),
		Action:    action,
		Flags: []cli.Flag{
			ConfigFlag,
			DebugFlag,
			VLevel,
			VModule,
			LogFile,
			AlsoLogToStderr,
			cli.StringFlag{
				Name:        "token,t",
				Usage:       "(cluster) Token to use for authentication",
				EnvVar:      version.ProgramUpper + "_TOKEN",
				Destination: &AgentConfig.Token,
			},
			cli.StringFlag{
				Name:        "token-file",
				Usage:       "(cluster) Token file to use for authentication",
				EnvVar:      version.ProgramUpper + "_TOKEN_FILE",
				Destination: &AgentConfig.TokenFile,
			},
			cli.StringFlag{
				Name:        "server,s",
				Usage:       "(cluster) Server to connect to",
				EnvVar:      version.ProgramUpper + "_URL",
				Destination: &AgentConfig.ServerURL,
			},
			cli.StringFlag{
				Name:        "data-dir,d",
				Usage:       "(agent/data) Folder to hold state",
				Destination: &AgentConfig.DataDir,
				Value:       "/var/lib/k8e/" + version.Program + "",
			},
			NodeNameFlag,
			WithNodeIDFlag,
			NodeLabels,
			NodeTaints,
			DockerFlag,
			CRIEndpointFlag,
			PauseImageFlag,
			SnapshotterFlag,
			PrivateRegistryFlag,
			NodeIPFlag,
			NodeExternalIPFlag,
			ResolvConfFlag,
			FlannelIfaceFlag,
			FlannelConfFlag,
			ExtraKubeletArgs,
			ExtraKubeProxyArgs,
			ProtectKernelDefaultsFlag,
			cli.BoolFlag{
				Name:        "rootless",
				Usage:       "(experimental) Run rootless",
				Destination: &AgentConfig.Rootless,
			},
			&SELinuxFlag,

			// Deprecated/hidden below

			&DisableSELinuxFlag,
			FlannelFlag,
			cli.StringFlag{
				Name:        "cluster-secret",
				Usage:       "(deprecated) use --token",
				Destination: &AgentConfig.ClusterSecret,
				EnvVar:      version.ProgramUpper + "_CLUSTER_SECRET",
			},
		},
	}
}
