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
	Name string
	Kind metric.MetricDescriptor_MetricKind
	Type metric.MetricDescriptor_ValueType
	Unit string
}

var (
	// Units format from https://ucum.org/ucum.html

	nvidiasmiQueries = []nvidiasmiQuery{
		{
			Name: "temperature.gpu",
			Kind: metric.MetricDescriptor_GAUGE,
			Type: metric.MetricDescriptor_INT64,
			Unit: "1/{degres C}",
		},
		{
			Name: "utilization.gpu",
			Kind: metric.MetricDescriptor_GAUGE,
			Type: metric.MetricDescriptor_INT64,
			Unit: "%",
		},
		{
			Name: "utilization.memory",
			Kind: metric.MetricDescriptor_GAUGE,
			Type: metric.MetricDescriptor_INT64,
			Unit: "%",
		},
		{
			Name: "memory.total",
			Kind: metric.MetricDescriptor_GAUGE,
			Type: metric.MetricDescriptor_INT64,
			Unit: "MiBy",
		},
		{
			Name: "memory.free",
			Kind: metric.MetricDescriptor_GAUGE,
			Type: metric.MetricDescriptor_INT64,
			Unit: "MiBy",
		},
		{
			Name: "memory.used",
			Kind: metric.MetricDescriptor_GAUGE,
			Type: metric.MetricDescriptor_INT64,
			Unit: "MiBy",
		},
	}
)

func (q *nvidiasmiQuery) gcpFormat() string {
	return strings.ReplaceAll(q.Name, ".", "_")
}

func getGPUAmount() (int, error) {
	o, err := exec.Command("/bin/sh",
		"-c",
		"nvidia-smi --query-gpu=index -u --format=csv,noheader",
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

func getGPUMetric(query string, id int) (int64, string, error) {
	var cmd string

	if id >= 0 {
		cmd = fmt.Sprintf("nvidia-smi --id=%d --query-gpu=%s -u --format=csv,noheader",
			id, query)
	} else {
		cmd = fmt.Sprintf("nvidia-smi --query-gpu=%s -u --format=csv,noheader",
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
