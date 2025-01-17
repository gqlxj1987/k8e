package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"github.com/xiaods/k8e/pkg/cli/cmds"
	"github.com/xiaods/k8e/pkg/data"
	"github.com/xiaods/k8e/pkg/datadir"
	"github.com/xiaods/k8e/pkg/dataverify"
	"github.com/xiaods/k8e/pkg/flock"
	"github.com/xiaods/k8e/pkg/untar"
	"github.com/xiaods/k8e/pkg/version"
)

func main() {
	if runCLIs() {
		return
	}

	app := cmds.NewApp()
	app.Commands = []cli.Command{
		cmds.NewServerCommand(wrap(version.Program+"-server", os.Args)),
		cmds.NewAgentCommand(wrap(version.Program+"-agent", os.Args)),
		cmds.NewKubectlCommand(externalCLIAction("kubectl")),
		cmds.NewCRICTL(externalCLIAction("crictl")),
		cmds.NewCtrCommand(externalCLIAction("ctr")),
		cmds.NewCheckConfigCommand(externalCLIAction("check-config")),
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}

func runCLIs() bool {
	if os.Getenv("CRI_CONFIG_FILE") == "" {
		os.Setenv("CRI_CONFIG_FILE", datadir.DefaultDataDir+"/agent/etc/crictl.yaml")
	}
	for _, cmd := range []string{"kubectl", "ctr", "crictl"} {
		if filepath.Base(os.Args[0]) == cmd {
			if err := externalCLI(cmd, "", os.Args[1:]); err != nil {
				logrus.Fatal(err)
			}
			return true
		}
	}
	return false
}

func externalCLIAction(cmd string) func(cli *cli.Context) error {
	return func(cli *cli.Context) error {
		return externalCLI(cmd, cli.String("data-dir"), cli.Args())
	}
}

func externalCLI(cli, dataDir string, args []string) error {
	dataDir, err := datadir.Resolve(dataDir)
	if err != nil {
		return err
	}
	return stageAndRun(dataDir, cli, append([]string{cli}, args...))
}

func wrap(cmd string, args []string) func(ctx *cli.Context) error {
	return func(ctx *cli.Context) error {
		return stageAndRunCLI(ctx, cmd, args)
	}
}

func stageAndRunCLI(cli *cli.Context, cmd string, args []string) error {
	dataDir, err := datadir.Resolve(cli.String("data-dir"))
	if err != nil {
		return err
	}

	return stageAndRun(dataDir, cmd, args)
}

func stageAndRun(dataDir string, cmd string, args []string) error {
	dir, err := extract(dataDir)
	if err != nil {
		return errors.Wrap(err, "extracting data")
	}
	logrus.Debugf("Asset dir %s", dir)

	if err := os.Setenv("PATH", filepath.Join(dir, "bin")+":"+os.Getenv("PATH")+":"+filepath.Join(dir, "bin/aux")); err != nil {
		return err
	}
	if err := os.Setenv(version.ProgramUpper+"_DATA_DIR", dir); err != nil {
		return err
	}

	cmd, err = exec.LookPath(cmd)
	if err != nil {
		return err
	}

	logrus.Debugf("Running %s %v", cmd, args)

	return syscall.Exec(cmd, args, os.Environ())
}

func getAssetAndDir(dataDir string) (string, string) {
	asset := data.AssetNames()[0]
	dir := filepath.Join(dataDir, "data", strings.SplitN(filepath.Base(asset), ".", 2)[0])
	return asset, dir
}

func extract(dataDir string) (string, error) {
	// first look for global asset folder so we don't create a HOME version if not needed
	_, dir := getAssetAndDir(datadir.DefaultDataDir)
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}

	asset, dir := getAssetAndDir(dataDir)
	// check if target directory already exists
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}

	// acquire a data directory lock
	os.MkdirAll(filepath.Join(dataDir, "data"), 0755)
	lockFile := filepath.Join(dataDir, "data", ".lock")
	logrus.Infof("Acquiring lock file %s", lockFile)
	lock, err := flock.Acquire(lockFile)
	if err != nil {
		return "", err
	}
	defer flock.Release(lock)

	// check again if target directory exists
	if _, err := os.Stat(dir); err == nil {
		return dir, nil
	}

	logrus.Infof("Preparing data dir %s", dir)

	content, err := data.Asset(asset)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(content)

	tempDest := dir + "-tmp"
	defer os.RemoveAll(tempDest)
	os.RemoveAll(tempDest)

	if err := untar.Untar(buf, tempDest); err != nil {
		return "", err
	}
	if err := dataverify.Verify(filepath.Join(tempDest, "bin")); err != nil {
		return "", err
	}

	currentSymLink := filepath.Join(dataDir, "data", "current")
	previousSymLink := filepath.Join(dataDir, "data", "previous")
	if _, err := os.Lstat(currentSymLink); err == nil {
		if err := os.Rename(currentSymLink, previousSymLink); err != nil {
			return "", err
		}
	}
	if err := os.Symlink(dir, currentSymLink); err != nil {
		return "", err
	}
	return dir, os.Rename(tempDest, dir)
}
