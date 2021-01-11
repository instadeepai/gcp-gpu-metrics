package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	metric "google.golang.org/genproto/googleapis/api/metric"
)

type nvidiasmiQuery struct {
	Name        string
	DisplayName string
	Kind        metric.MetricDescriptor_MetricKind
	Type        metric.MetricDescriptor_ValueType
	Unit        string
}

var (
	// Units format from https://ucum.org/ucum.html

	nvidiasmiQueries = []nvidiasmiQuery{
		{
			Name:        "temperature.gpu",
			DisplayName: "Temperature GPU",
			Kind:        metric.MetricDescriptor_GAUGE,
			Type:        metric.MetricDescriptor_INT64,
			Unit:        "1/{degres C}",
		},
		{
			Name:        "utilization.gpu",
			DisplayName: "Utilization GPU",
			Kind:        metric.MetricDescriptor_GAUGE,
			Type:        metric.MetricDescriptor_INT64,
			Unit:        "%",
		},
		{
			Name:        "utilization.memory",
			DisplayName: "Utilization Memory GPU",
			Kind:        metric.MetricDescriptor_GAUGE,
			Type:        metric.MetricDescriptor_INT64,
			Unit:        "%",
		},
		{
			Name:        "memory.total",
			DisplayName: "Memory Total GPU",
			Kind:        metric.MetricDescriptor_GAUGE,
			Type:        metric.MetricDescriptor_INT64,
			Unit:        "MiBy",
		},
		{
			Name:        "memory.free",
			DisplayName: "Memory Free GPU",
			Kind:        metric.MetricDescriptor_GAUGE,
			Type:        metric.MetricDescriptor_INT64,
			Unit:        "MiBy",
		},
		{
			Name:        "memory.used",
			DisplayName: "Memory Used GPU",
			Kind:        metric.MetricDescriptor_GAUGE,
			Type:        metric.MetricDescriptor_INT64,
			Unit:        "MiBy",
		},
	}
)

func (q *nvidiasmiQuery) gcpFormat() string {
	return strings.ReplaceAll(q.Name, ".", "_")
}

const (
	queryFormat string = "-u --format=csv,noheader"
)

func getGPUAmount() (int, error) {
	o, err := exec.Command("/bin/sh",
		"-c",
		"nvidia-smi --query-gpu=index "+queryFormat,
	).Output()
	if err != nil {
		return 0, fmt.Errorf("%s - %s", err.Error(), string(o))
	}

	amount := len(strings.Split(string(o), "\n")) - 1
	if amount == 0 {
		return 0, errors.New("Can't fetch metrics on 0 GPUs")
	}

	return amount, nil
}

func getGPUbusID(id int) (string, error) {
	o, err := exec.Command("/bin/sh",
		"-c",
		fmt.Sprintf("nvidia-smi --query-gpu=pci.bus_id --id=%d "+queryFormat,
			id),
	).Output()
	if err != nil {
		return "", fmt.Errorf("%s - %s", err.Error(), string(o))
	}

	return strings.Split(string(o), "\n")[0], nil
}

func getGPUMetric(query string, id int) (int64, string, error) {
	var cmd string

	if id >= 0 {
		cmd = fmt.Sprintf("nvidia-smi --id=%d --query-gpu=%s "+queryFormat,
			id, query)
	} else {
		cmd = fmt.Sprintf("nvidia-smi --query-gpu=%s "+queryFormat,
			query)
	}

	o, err := exec.Command("/bin/sh",
		"-c",
		cmd,
	).Output()
	if err != nil {
		return 0, "", err
	}

	lines := strings.Split(string(o), "\n")

	amount := int64(len(lines) - 1)

	if amount == 0 {
		return 0, "", errors.New("Can't fetch metrics on 0 GPUs")
	}

	//delete last blank line
	lines = lines[:amount]

	sumValues := int64(0)
	unit := ""

	for _, line := range lines {
		elem := strings.Split(line, " ")
		if len(elem) >= 2 {
			unit = elem[1]
		}
		v, err := strconv.ParseInt(elem[0], 10, 64)
		if err != nil {
			v = 0
		}
		sumValues += v
	}

	value := sumValues / amount

	return value, unit, nil
}

func isNvidiasmiExist() error {
	o, err := exec.Command("/bin/sh",
		"-c",
		"nvidia-smi --list-gpus",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s - %s", err.Error(), string(o))
	}

	return nil
}

// enablePMNvidiasmi aims to enable persistence mod on nvidia smi
// to prevent 100% gpu usage on one GPU at each query
func enablePMNvidiasmi() error {
	o, err := exec.Command("/bin/sh",
		"-c",
		"sudo nvidia-smi -pm 1",
	).Output()
	if err != nil {
		return fmt.Errorf("%s - %s", err.Error(), string(o))
	}

	return nil
}
