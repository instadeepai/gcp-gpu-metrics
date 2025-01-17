package main

import (
	"flag"
	"fmt"
	"log/syslog"
	"os"
	"strconv"
)

var (
	// flags related

	flagDisplayVersion       bool   = false
	flagServiceAccountPath   string = ""
	flagFetchMetricsInterval uint64 = 10
	flagEnableNvidiasmipm    bool   = false

	envVarPrefix = "GGM_"

	// Version represents gcp-gpu-metrics version
	Version string
	// Commit represents gcp-gpu-metrics build commit hash
	Commit string
)

func newSyslogger() (*syslog.Writer, error) {
	return syslog.New(syslog.LOG_INFO|syslog.LOG_SYSLOG, "gcp-gpu-metrics")
}

func evaluateEnvVars() {
	tmpSAP := os.Getenv(envVarPrefix + "SERVICE_ACCOUNT_PATH")
	if tmpSAP != "" {
		flagServiceAccountPath = tmpSAP
	}

	tmpFM := os.Getenv(envVarPrefix + "METRICS_INTERVAL")
	if tmpFM != "" {
		v, err := strconv.ParseUint(tmpFM, 10, 64)
		if err == nil {
			flagFetchMetricsInterval = v
		}
	}

	tmpNPM := os.Getenv(envVarPrefix + "ENABLE_NVIDIASMI_PM")
	if tmpNPM != "" {
		v, err := strconv.ParseBool(tmpNPM)
		if err == nil {
			flagEnableNvidiasmipm = v
		}
	}
}

func main() {
	evaluateEnvVars()

	flag.BoolVar(&flagDisplayVersion, "version", flagDisplayVersion, "Display current version/release and commit hash.")
	flag.StringVar(&flagServiceAccountPath, "service-account-path", flagServiceAccountPath, "GCP service account path.")
	flag.Uint64Var(&flagFetchMetricsInterval, "metrics-interval", flagFetchMetricsInterval, "Fetch metrics interval in seconds.")
	flag.BoolVar(&flagEnableNvidiasmipm, "enable-nvidiasmi-pm", flagEnableNvidiasmipm, "Enable persistant mod for nvidia-smi.")
	flag.Parse()

	if flagDisplayVersion {
		fmt.Printf("Current version: %s\n", Version)
		fmt.Printf("Current commit: %s\n", Commit)
		os.Exit(0)
	}

	// init syslogger
	slog, err := newSyslogger()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// check if nvidia-smi binary is present on the instance
	if err := isNvidiasmiExist(); err != nil {
		_ = slog.Err(err.Error())
		os.Exit(1)
	}
	_ = slog.Info("nvidia-smi detected")

	// enable nvidia-smi persistence mod
	if flagEnableNvidiasmipm {
		if err := enablePMNvidiasmi(); err != nil {
			_ = slog.Info(err.Error())
		} else {
			_ = slog.Info("nvidia-smi persistence mod enabled")
		}
	}

	// get GPU amount on the instance
	gpuAmount, err := getGPUAmount()
	if err != nil {
		_ = slog.Err(err.Error())
		os.Exit(1)
	}
	_ = slog.Info(fmt.Sprintf("%d GPU(s) detected\n", gpuAmount))

	// create a new GCP auth service
	s, err := newService(slog)
	if err != nil {
		_ = slog.Err(err.Error())
		os.Exit(1)
	}

	defer s.Close()

	// creation loop of metrics descriptors
	if err := s.createMetricsDescriptors(); err != nil {
		_ = slog.Err(err.Error())
		os.Exit(1)
	}

	// fetch metrics infinite loop
	s.fetchMetrics(gpuAmount)

	os.Exit(0)
}
