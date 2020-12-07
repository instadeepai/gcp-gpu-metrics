package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log/syslog"
	"net/http"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"google.golang.org/api/option"
	metric "google.golang.org/genproto/googleapis/api/metric"
	monitoredres "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	metadataServer = "http://metadata/computeMetadata/v1/instance/"
)

type service struct {
	*monitoring.MetricClient
	zone       string
	projectID  string
	instanceID string
	slog       *syslog.Writer
}

func newService(slog *syslog.Writer) (*service, error) {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx, option.WithCredentialsFile(flagServiceAccountPath))
	if err != nil {
		return nil, err
	}

	s := &service{
		MetricClient: client,
		slog:         slog,
	}

	// Get projectID and zone by querying zone metadata server
	mzone, err := retrieveInstanceMetadata("zone")
	if err != nil {
		return nil, err
	}

	s.zone = strings.Split(mzone, "/")[3]
	s.projectID = strings.Split(mzone, "/")[1]

	// Get instanceID by querying id metadata server
	mid, err := retrieveInstanceMetadata("id")
	if err != nil {
		return nil, err
	}

	s.instanceID = mid

	return s, nil
}

func retrieveInstanceMetadata(mpath string) (string, error) {
	httpClient := &http.Client{}

	req, _ := http.NewRequest("GET", metadataServer+mpath, nil)
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return strings.Split(string(b), "\n")[0], nil
}

func (s *service) createMetricDescriptor() error {
	for _, query := range nvidiasmiQueries {
		fquery := query.gcpFormat()

		req := &monitoringpb.CreateMetricDescriptorRequest{
			Name: "projects/" + s.projectID,
			MetricDescriptor: &metric.MetricDescriptor{
				Name:        fquery,
				Type:        "custom.googleapis.com/gpu/" + fquery,
				MetricKind:  query.Kind,
				ValueType:   query.Type,
				Unit:        query.Unit,
				Description: "gcp_gpu_metrics for " + fquery + " nvidia-smi query",
				DisplayName: fquery,
			},
		}

		ctx := context.Background()

		resp, err := s.CreateMetricDescriptor(ctx, req)
		if err != nil {
			return fmt.Errorf("%s - %s", resp, err.Error())
		}

		_ = s.slog.Info("Metric descriptor created for " + fquery)
	}

	return nil
}

func (s *service) fetchMetricsLoop() {
	fmi := flagFetchMetricsInterval
	_ = s.slog.Info(fmt.Sprintf("Start fetching metrics every %d seconds", fmi))

	for {
		for _, query := range nvidiasmiQueries {
			value, _, err := getGPUsMetric(query.Name)
			if err != nil {
				fmt.Println(err)
			}

			s.createTimeSeries(value, &query)
		}

		time.Sleep(time.Duration(flagFetchMetricsInterval) * time.Second)
	}
}

func (s *service) createTimeSeries(value int64, q *nvidiasmiQuery) {
	now := time.Now()

	fquery := q.gcpFormat()

	req := &monitoringpb.CreateTimeSeriesRequest{
		Name: "projects/" + s.projectID,
		TimeSeries: []*monitoringpb.TimeSeries{
			{
				Metric: &metric.Metric{
					Type: "custom.googleapis.com/gpu/" + fquery,
				},
				Resource: &monitoredres.MonitoredResource{
					Type: "gce_instance",
					Labels: map[string]string{
						"instance_id": s.instanceID,
						"zone":        s.zone,
						"project_id":  s.projectID,
					},
				},
				MetricKind: q.Kind,
				ValueType:  q.Type,
				Points: []*monitoringpb.Point{
					{
						Interval: &monitoringpb.TimeInterval{
							EndTime: &timestamppb.Timestamp{
								Seconds: int64(now.Unix()),
								Nanos:   int32(now.Nanosecond()),
							},
						},
						Value: &monitoringpb.TypedValue{
							Value: &monitoringpb.TypedValue_Int64Value{
								Int64Value: value,
							},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()

	err := s.CreateTimeSeries(ctx, req)
	if err != nil {
		_ = s.slog.Err(err.Error())
	}
}
