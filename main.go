package main

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	build   string
	commit  string
	version string
	clash   string
	branch  string
	binName string
)

var conf TPClashConf

var rootCmd = &cobra.Command{
	Use:   "tpclash",
	Short: "Transparent proxy tool for Clash",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("%s\nVersion: %s\nBuild: %s\nClash Core: %s\nCommit: %s\n\n", logo, version, build, clash, commit)

		if conf.PrintVersion {
			return
		}

		if conf.Debug {
			h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
			slog.SetDefault(slog.New(h))
		}

		slog.Info("[main] starting tpclash...")

		// Initialize signal control Context
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		defer cancel()

		// Configure Sysctl
		Sysctl()

		// Extract Clash executable and built-in configuration files
		ExtractFiles()

		// Watch config file
		updateCh := WatchConfig(ctx)

		// Wait for the first config to return
		clashConfStr := <-updateCh

		// Check clash config
		cc, err := CheckConfig(clashConfStr)
		if err != nil {
			slog.Error("", err)
		}

		// Copy remote or local clash config file to internal path
		clashConfPath := filepath.Join(conf.ClashHome, InternalConfigName)
		if err = os.WriteFile(clashConfPath, []byte(clashConfStr), 0644); err != nil {
			slog.Error("[main] failed to copy clash config: %v", err)
		}

		// Create child process
		clashBinPath := filepath.Join(conf.ClashHome, InternalClashBinName)
		clashUIPath := filepath.Join(conf.ClashHome, conf.ClashUI)
		cmd := exec.Command(clashBinPath, "-f", clashConfPath, "-d", conf.ClashHome, "-ext-ui", clashUIPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &syscall.SysProcAttr{}
		slog.Info("[main] running cmds: %v", cmd.Args)

		if err = cmd.Start(); err != nil {
			slog.Error("[main] failed to start clash process: %v: %v", err, cmd.Args)
			cancel()
		}
		if cmd.Process == nil {
			slog.Error("[main] failed to start clash process: %v", cmd.Args)
			cancel()
		}

		if err = EnableDockerCompatible(); err != nil {
			slog.Error("[main] failed enable docker compatible: %v", err)
		}

		// Watch clash config changes, and automatically reload the config
		go AutoReload(updateCh, clashConfPath)

		slog.Info("[main] 🍄 提莫队长正在待命...")
		if conf.Test {
			slog.Warn("[main] test mode enabled, tpclash will automatically exit after 5 minutes...")
			go func() {
				<-time.After(5 * time.Minute)
				cancel()
			}()
		}

		if conf.EnableTracing {
			slog.Info("[main] 🔪 永远不要忘记, 吾等为何而战...")
			// always clean tracing containers
			if err = stopTracing(ctx); err != nil {
				slog.Error("[main] ❌ tracing project cleanup failed: %v", err)
			}
			if err = startTracing(ctx, conf, cc); err != nil {
				slog.Error("[main] ❌ tracing project deploy failed: %v", err)
			}
		}

		<-ctx.Done()
		slog.Info("[main] 🛑 TPClash 正在停止...")
		if err = DisableDockerCompatible(); err != nil {
			slog.Error("[main] failed disable docker compatible: %v", err)
		}

		if conf.EnableTracing {
			slog.Info("[main] 🔪 恐惧, 是万敌之首...")

			tracingStopCtx, tracingStopCancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer tracingStopCancel()

			if err = stopTracing(tracingStopCtx); err != nil {
				slog.Error("[main] ❌ tracing project stop failed: %v", err)
			}
		}

		if cmd.Process != nil {
			if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
				slog.Error("", err)
			}
		}

		slog.Info("[main] 🛑 TPClash 已关闭!")
	},
}

func init() {
	cobra.EnableCommandSorting = false

	rootCmd.AddCommand(encCmd, decCmd, installCmd, uninstallCmd, upgradeCmd)

	rootCmd.PersistentFlags().BoolVar(&conf.Debug, "debug", false, "enable debug log")
	rootCmd.PersistentFlags().BoolVar(&conf.Test, "test", false, "enable test mode, tpclash will automatically exit after 5 minutes")
	rootCmd.PersistentFlags().StringVarP(&conf.ClashHome, "home", "d", "/data/clash", "clash home dir")
	rootCmd.PersistentFlags().StringVarP(&conf.ClashConfig, "config", "c", "/etc/clash.yaml", "clash config local path or remote url")
	rootCmd.PersistentFlags().StringVarP(&conf.ClashUI, "ui", "u", "yacd", "clash dashboard(official|yacd)")
	rootCmd.PersistentFlags().DurationVarP(&conf.CheckInterval, "check-interval", "i", 120*time.Second, "remote config check interval")
	rootCmd.PersistentFlags().StringSliceVar(&conf.HttpHeader, "http-header", []string{}, "http header when requesting a remote config(key=value)")
	rootCmd.PersistentFlags().DurationVar(&conf.HttpTimeout, "http-timeout", 10*time.Second, "http request timeout when requesting a remote config")
	rootCmd.PersistentFlags().StringVar(&conf.ConfigEncPassword, "config-password", "", "the password for encrypting the config file")
	rootCmd.PersistentFlags().StringVar(&conf.AutoFixMode, "auto-fix", "", "automatically repair config(tun/ebpf)")
	rootCmd.PersistentFlags().BoolVar(&conf.ForceExtract, "force-extract", false, "extract files force")
	rootCmd.PersistentFlags().BoolVar(&conf.AllowStandardDNSPort, "allow-standard-dns", false, "allow standard DNS port")
	rootCmd.PersistentFlags().BoolVarP(&conf.PrintVersion, "version", "v", false, "version for tpclash")

	if branch == "premium" {
		rootCmd.PersistentFlags().BoolVar(&conf.EnableTracing, "enable-tracing", false, "auto deploy tracing dashboard")
	}
}

func main() {
	cobra.CheckErr(rootCmd.Execute())
}
