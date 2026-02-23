package otelcontrollers

import (
	"github.com/tracewayapp/traceway/backend/app/models"

	"github.com/google/uuid"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
)

func convertMetricPoints(projectId uuid.UUID, req *colmetricspb.ExportMetricsServiceRequest) []models.MetricPoint {
	var points []models.MetricPoint

	for _, rm := range req.ResourceMetrics {
		resAttrs := rm.GetResource().GetAttributes()
		sn := getStringAttribute(resAttrs, "service.name")

		for _, sm := range rm.ScopeMetrics {
			for _, metric := range sm.Metrics {
				name := metric.Name

				switch data := metric.Data.(type) {
				case *metricspb.Metric_Gauge:
					points = appendNumberDataPoints(points, projectId, name, sn, data.Gauge.GetDataPoints())
				case *metricspb.Metric_Sum:
					points = appendNumberDataPoints(points, projectId, name, sn, data.Sum.GetDataPoints())
				case *metricspb.Metric_Histogram:
					for _, dp := range data.Histogram.GetDataPoints() {
						ts := nanoToTime(dp.TimeUnixNano)
						tags := make(map[string]string)
						if sn != "" {
							tags["server_name"] = sn
						}
						if dp.Count > 0 && dp.Sum != nil {
							points = append(points, models.MetricPoint{
								ProjectId:  projectId,
								Name:       name + ".avg",
								Value:      *dp.Sum / float64(dp.Count),
								Tags:       tags,
								RecordedAt: ts,
							})
						}
						points = append(points, models.MetricPoint{
							ProjectId:  projectId,
							Name:       name + ".count",
							Value:      float64(dp.Count),
							Tags:       tags,
							RecordedAt: ts,
						})
					}
				}
			}
		}
	}
	return points
}

func appendNumberDataPoints(points []models.MetricPoint, projectId uuid.UUID, name, serverName string, dps []*metricspb.NumberDataPoint) []models.MetricPoint {
	for _, dp := range dps {
		var value float64
		switch v := dp.Value.(type) {
		case *metricspb.NumberDataPoint_AsDouble:
			value = v.AsDouble
		case *metricspb.NumberDataPoint_AsInt:
			value = float64(v.AsInt)
		}
		tags := make(map[string]string)
		if serverName != "" {
			tags["server_name"] = serverName
		}
		points = append(points, models.MetricPoint{
			ProjectId:  projectId,
			Name:       name,
			Value:      value,
			Tags:       tags,
			RecordedAt: nanoToTime(dp.TimeUnixNano),
		})
	}
	return points
}
