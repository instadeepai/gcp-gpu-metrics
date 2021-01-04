# gcp-gpu-metrics

![CI](https://github.com/instadeepai/gcp-gpu-metrics/workflows/CI/badge.svg?branch=master)

Tiny Go binary that aims to export Nvidia GPU metrics to GCP monitoring, based on nvidia-smi.

## Requirements ⚓

* Your machine must be a GCE (Google Compute Engine) instance.
* Your Service Account must have the `Monitoring Metric Writer` permission.
* You need the `nvidia-smi` binary installed on your GCE instance.


Protip: You can use a [machine learning image](https://cloud.google.com/ai-platform/deep-learning-vm/docs/images) provided by GCP as default image.

## Install ⏬

If you're root, you can install the latest binary version using the following script:
```bash
$ bash -e < <(curl -sSL https://raw.githubusercontent.com/instadeepai/gcp-gpu-metrics/master/install-latest.sh)
```

Or, you can download a [release/binary from this page](https://github.com/instadeepai/gcp-gpu-metrics/releases) and install it manually.

## Usage 💻

gcp-gpu-metrics is an UNIX compliant and very simple CLI, you just have to use it as usual:

```bash
$ gcp-gpu-metrics
```

Available flags:

* `--service-account-path string` | GCP service account path. (default "./service-account.json")
* `--metrics-interval uint` | Fetch metrics interval in seconds. (default 10)
* `--enable-nvidiasmi-pm` | Enable persistence mod for nvidia-smi. (default false)
* `--version` | Display current version/release and commit hash.

Available env variables:
* `GGM_SERVICE_ACCOUNT_PATH=./service-account.json` linked to `--service-account-path` flag.
* `GGM_METRICS_INTERVAL=10` linked to `--metrics-interval` flag.
* `GGM_ENABLE_NVIDIASMI_PM=true` linked to `--enable-nvidiasmi-pm` flag.

Priority order is `binary flag` ➡️ `env var` ➡️ `default value`.

Nvidia-smi persistence mod is very useful, the option permits to run `nvidia-smi` as a daemon in background to prevent 100% of GPU load at each request. Enabling this option requires root.

## Metrics 📈

There are 6 differents metrics fetched, this number will grow in the future.

* `temperature.gpu` as `custom.googleapis.com/gpu/temperature_gpu` | Core GPU temperature. in degrees C.
* `utilization.gpu` as `custom.googleapis.com/gpu/utilization_gpu` | Percent of time over the past sample period during which one or more kernels were executed on the GPU.
* `utilization.memory` as `custom.googleapis.com/gpu/utilization_memory` | Percent of time over the past sample period during which global (device) memory was being read or written.
* `memory.total` as `custom.googleapis.com/gpu/memory_total` | Total installed GPU memory.
* `memory.free` as `custom.googleapis.com/gpu/memory_free` | Total GPU free memory.
* `memory.used` as `custom.googleapis.com/gpu/memory_used` | Total memory allocated by active contexts.

For the moment, gcp-gpu-metrics sends an average of any metrics if you have more than 1 GPU.

## Compile gcp-gpu-metrics ⚙

There is a `re` command in the Makefile.

```bash
$ make re
```

gcp-gpu-metrics has been tested with `go1.15` and use go modules.

## Report an issue 📢

Feel free to open a GitHub issue on this project 🚀

## License 🔑

See [LICENSE](LICENSE).