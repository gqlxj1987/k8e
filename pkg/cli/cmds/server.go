package cmds

import (
	"context"

	"github.com/urfave/cli"
	"github.com/xiaods/k8e/pkg/version"
)

const (
	DisableItems = "coredns, servicelb, traefik, local-storage, metrics-server"

	defaultSnapshotRentention    = 5
	defaultSnapshotIntervalHours = 12
)

type Server struct {
	ClusterCIDR    string
	AgentToken     string
	AgentTokenFile string
	Token          string
	TokenFile      string
	ClusterSecret  string
	ServiceCIDR    string
	ClusterDNS     string
	ClusterDomain  string
	// The port which kubectl clients can access k8s
	HTTPSPort int
	// The port which custom k3s API runs on
	SupervisorPort int
	// The port which kube-apiserver runs on
	APIServerPort            int
	APIServerBindAddress     string
	DataDir                  string
	DisableAgent             bool
	KubeConfigOutput         string
	KubeConfigMode           string
	TLSSan                   cli.StringSlice
	BindAddress              string
	ExtraAPIArgs             cli.StringSlice
	ExtraSchedulerArgs       cli.StringSlice
	ExtraControllerArgs      cli.StringSlice
	ExtraCloudControllerArgs cli.StringSlice
	Rootless                 bool
	DatastoreEndpoint        string
	DatastoreCAFile          string
	DatastoreCertFile        string
	DatastoreKeyFile         string
	AdvertiseIP              string
	AdvertisePort            int
	DisableScheduler         bool
	ServerURL                string
	FlannelBackend           string
	DefaultLocalStoragePath  string
	DisableCCM               bool
	DisableNPC               bool
	DisableKubeProxy         bool
	ClusterInit              bool
	ClusterReset             bool
	ClusterResetRestorePath  string
	EncryptSecrets           bool
	StartupHooks             []func(context.Context, <-chan struct{}, string) error
	EtcdDisableSnapshots     bool
	EtcdSnapshotDir          string
	EtcdSnapshotCron         string
	EtcdSnapshotRetention    int
}

var ServerConfig Server

func NewServerCommand(action func(*cli.Context) error) cli.Command {
	return cli.Command{
		Name:      "server",
		Usage:     "Run management server",
		UsageText: appName + " server [OPTIONS]",
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
				Name:        "bind-address",
				Usage:       "(listener) " + version.Program + " bind address (default: 0.0.0.0)",
				Destination: &ServerConfig.BindAddress,
			},
			cli.IntFlag{
				Name:        "https-listen-port",
				Usage:       "(listener) HTTPS listen port",
				Value:       6443,
				Destination: &ServerConfig.HTTPSPort,
			},
			cli.StringFlag{
				Name:        "advertise-address",
				Usage:       "(listener) IP address that apiserver uses to advertise to members of the cluster (default: node-external-ip/node-ip)",
				Destination: &ServerConfig.AdvertiseIP,
			},
			cli.IntFlag{
				Name:        "advertise-port",
				Usage:       "(listener) Port that apiserver uses to advertise to members of the cluster (default: listen-port)",
				Destination: &ServerConfig.AdvertisePort,
			},
			cli.StringSliceFlag{
				Name:  "tls-san",
				Usage: "(listener) Add additional hostname or IP as a Subject Alternative Name in the TLS cert",
				Value: &ServerConfig.TLSSan,
			},
			cli.StringFlag{
				Name:        "data-dir,d",
				Usage:       "(data) Folder to hold state default /var/lib/k8e/" + version.Program + " or ${HOME}/.k8e/" + version.Program + " if not root",
				Destination: &ServerConfig.DataDir,
			},
			cli.StringFlag{
				Name:        "cluster-cidr",
				Usage:       "(networking) Network CIDR to use for pod IPs",
				Destination: &ServerConfig.ClusterCIDR,
				Value:       "10.42.0.0/16",
			},
			cli.StringFlag{
				Name:        "service-cidr",
				Usage:       "(networking) Network CIDR to use for services IPs",
				Destination: &ServerConfig.ServiceCIDR,
				Value:       "10.43.0.0/16",
			},
			cli.StringFlag{
				Name:        "cluster-dns",
				Usage:       "(networking) Cluster IP for coredns service. Should be in your service-cidr range (default: 10.43.0.10)",
				Destination: &ServerConfig.ClusterDNS,
				Value:       "",
			},
			cli.StringFlag{
				Name:        "cluster-domain",
				Usage:       "(networking) Cluster Domain",
				Destination: &ServerConfig.ClusterDomain,
				Value:       "cluster.local",
			},
			cli.StringFlag{
				Name:        "flannel-backend",
				Usage:       "(networking) One of 'none', 'vxlan', 'ipsec', 'host-gw', or 'wireguard'",
				Destination: &ServerConfig.FlannelBackend,
				Value:       "vxlan",
			},
			cli.StringFlag{
				Name:        "token,t",
				Usage:       "(cluster) Shared secret used to join a server or agent to a cluster",
				Destination: &ServerConfig.Token,
				EnvVar:      version.ProgramUpper + "_TOKEN",
			},
			cli.StringFlag{
				Name:        "token-file",
				Usage:       "(cluster) File containing the cluster-secret/token",
				Destination: &ServerConfig.TokenFile,
				EnvVar:      version.ProgramUpper + "_TOKEN_FILE",
			},
			cli.StringFlag{
				Name:        "write-kubeconfig,o",
				Usage:       "(client) Write kubeconfig for admin client to this file",
				Destination: &ServerConfig.KubeConfigOutput,
				EnvVar:      version.ProgramUpper + "_KUBECONFIG_OUTPUT",
			},
			cli.StringFlag{
				Name:        "write-kubeconfig-mode",
				Usage:       "(client) Write kubeconfig with this mode",
				Destination: &ServerConfig.KubeConfigMode,
				EnvVar:      version.ProgramUpper + "_KUBECONFIG_MODE",
			},
			cli.StringSliceFlag{
				Name:  "kube-apiserver-arg",
				Usage: "(flags) Customized flag for kube-apiserver process",
				Value: &ServerConfig.ExtraAPIArgs,
			},
			cli.StringSliceFlag{
				Name:  "kube-scheduler-arg",
				Usage: "(flags) Customized flag for kube-scheduler process",
				Value: &ServerConfig.ExtraSchedulerArgs,
			},
			cli.StringSliceFlag{
				Name:  "kube-controller-manager-arg",
				Usage: "(flags) Customized flag for kube-controller-manager process",
				Value: &ServerConfig.ExtraControllerArgs,
			},
			cli.StringSliceFlag{
				Name:  "kube-cloud-controller-manager-arg",
				Usage: "(flags) Customized flag for kube-cloud-controller-manager process",
				Value: &ServerConfig.ExtraCloudControllerArgs,
			},
			cli.StringFlag{
				Name:        "datastore-endpoint",
				Usage:       "(db) Specify etcd, Mysql, Postgres, or Sqlite (default) data source name",
				Destination: &ServerConfig.DatastoreEndpoint,
				EnvVar:      version.ProgramUpper + "_DATASTORE_ENDPOINT",
			},
			cli.StringFlag{
				Name:        "datastore-cafile",
				Usage:       "(db) TLS Certificate Authority file used to secure datastore backend communication",
				Destination: &ServerConfig.DatastoreCAFile,
				EnvVar:      version.ProgramUpper + "_DATASTORE_CAFILE",
			},
			cli.StringFlag{
				Name:        "datastore-certfile",
				Usage:       "(db) TLS certification file used to secure datastore backend communication",
				Destination: &ServerConfig.DatastoreCertFile,
				EnvVar:      version.ProgramUpper + "_DATASTORE_CERTFILE",
			},
			cli.StringFlag{
				Name:        "datastore-keyfile",
				Usage:       "(db) TLS key file used to secure datastore backend communication",
				Destination: &ServerConfig.DatastoreKeyFile,
				EnvVar:      version.ProgramUpper + "_DATASTORE_KEYFILE",
			},
			&cli.BoolFlag{
				Name:        "etcd-disable-snapshots",
				Usage:       "(db) Disable automatic etcd snapshots",
				Destination: &ServerConfig.EtcdDisableSnapshots,
			},
			&cli.StringFlag{
				Name:        "etcd-snapshot-schedule-cron",
				Usage:       "(db) Snapshot interval time in cron spec. eg. every 5 hours '* */5 * * *'",
				Destination: &ServerConfig.EtcdSnapshotCron,
				Value:       "0 */12 * * *",
			},
			&cli.IntFlag{
				Name:        "etcd-snapshot-retention",
				Usage:       "(db) Number of snapshots to retain",
				Destination: &ServerConfig.EtcdSnapshotRetention,
				Value:       defaultSnapshotRentention,
			},
			&cli.StringFlag{
				Name:        "etcd-snapshot-dir",
				Usage:       "(db) Directory to save db snapshots. (Default location: ${data-dir}/db/snapshots)",
				Destination: &ServerConfig.EtcdSnapshotDir,
			},
			cli.StringFlag{
				Name:        "default-local-storage-path",
				Usage:       "(storage) Default local storage path for local provisioner storage class",
				Destination: &ServerConfig.DefaultLocalStoragePath,
			},
			cli.StringSliceFlag{
				Name:  "disable",
				Usage: "(components) Do not deploy packaged components and delete any deployed components (valid items: " + DisableItems + ")",
			},
			cli.BoolFlag{
				Name:        "disable-scheduler",
				Usage:       "(components) Disable Kubernetes default scheduler",
				Destination: &ServerConfig.DisableScheduler,
			},
			cli.BoolFlag{
				Name:        "disable-cloud-controller",
				Usage:       "(components) Disable " + version.Program + " default cloud controller manager",
				Destination: &ServerConfig.DisableCCM,
			},
			cli.BoolFlag{
				Name:        "disable-kube-proxy",
				Usage:       "(components) Disable running kube-proxy",
				Destination: &ServerConfig.DisableKubeProxy,
			},
			cli.BoolFlag{
				Name:        "disable-network-policy",
				Usage:       "(components) Disable " + version.Program + " default network policy controller",
				Destination: &ServerConfig.DisableNPC,
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
				Destination: &ServerConfig.Rootless,
			},
			cli.StringFlag{
				Name:        "agent-token",
				Usage:       "(experimental/cluster) Shared secret used to join agents to the cluster, but not servers",
				Destination: &ServerConfig.AgentToken,
				EnvVar:      version.ProgramUpper + "_AGENT_TOKEN",
			},
			cli.StringFlag{
				Name:        "agent-token-file",
				Usage:       "(experimental/cluster) File containing the agent secret",
				Destination: &ServerConfig.AgentTokenFile,
				EnvVar:      version.ProgramUpper + "_AGENT_TOKEN_FILE",
			},
			cli.StringFlag{
				Name:        "server,s",
				Hidden:      hideClusterFlags,
				Usage:       "(experimental/cluster) Server to connect to, used to join a cluster",
				EnvVar:      version.ProgramUpper + "_URL",
				Destination: &ServerConfig.ServerURL,
			},
			cli.BoolFlag{
				Name:        "cluster-init",
				Hidden:      hideClusterFlags,
				Usage:       "(experimental/cluster) Initialize a new cluster",
				EnvVar:      version.ProgramUpper + "_CLUSTER_INIT",
				Destination: &ServerConfig.ClusterInit,
			},
			cli.BoolFlag{
				Name:        "cluster-reset",
				Hidden:      hideClusterFlags,
				Usage:       "(experimental/cluster) Forget all peers and become sole member of a new cluster",
				EnvVar:      version.ProgramUpper + "_CLUSTER_RESET",
				Destination: &ServerConfig.ClusterReset,
			},
			&cli.StringFlag{
				Name:        "cluster-reset-restore-path",
				Usage:       "(db) Path to snapshot file to be restored",
				Destination: &ServerConfig.ClusterResetRestorePath,
			},
			cli.BoolFlag{
				Name:        "secrets-encryption",
				Usage:       "(experimental) Enable Secret encryption at rest",
				Destination: &ServerConfig.EncryptSecrets,
			},
			&SELinuxFlag,

			// Hidden/Deprecated flags below

			&DisableSELinuxFlag,
			FlannelFlag,
			cli.StringSliceFlag{
				Name:  "no-deploy",
				Usage: "(deprecated) Do not deploy packaged components (valid items: " + DisableItems + ")",
			},
			cli.StringFlag{
				Name:        "cluster-secret",
				Usage:       "(deprecated) use --token",
				Destination: &ServerConfig.ClusterSecret,
				EnvVar:      version.ProgramUpper + "_CLUSTER_SECRET",
			},
			cli.BoolFlag{
				Name:        "disable-agent",
				Usage:       "Do not run a local agent and register a local kubelet",
				Hidden:      true,
				Destination: &ServerConfig.DisableAgent,
			},
			cli.StringSliceFlag{
				Hidden: true,
				Name:   "kube-controller-arg",
				Usage:  "(flags) Customized flag for kube-controller-manager process",
				Value:  &ServerConfig.ExtraControllerArgs,
			},
			cli.StringSliceFlag{
				Hidden: true,
				Name:   "kube-cloud-controller-arg",
				Usage:  "(flags) Customized flag for kube-cloud-controller-manager process",
				Value:  &ServerConfig.ExtraCloudControllerArgs,
			},
		},
	}
}
