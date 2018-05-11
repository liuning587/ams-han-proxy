package electricity

import (
	"context"

	"github.com/golang/protobuf/ptypes"
	influx "github.com/influxdata/influxdb/client/v2"

	api "svenschwermer.de/ams-han-proxy/proto/electricity"
)

type Handler struct {
	influxClient influx.Client
	database     string
}

func NewHandler(influxConfig influx.HTTPConfig, database string) (h *Handler, err error) {
	h = &Handler{
		database: database,
	}
	h.influxClient, err = influx.NewHTTPClient(influxConfig)
	if err != nil {
		return nil, err
	}
	return
}

func (h *Handler) Publish(ctx context.Context, req *api.MeterData) (*api.MeterDataReply, error) {
	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Precision: "ms",
		Database:  h.database,
	})
	// TODO: Properly wrap errors
	t, err := ptypes.Timestamp(req.HostTimestamp)
	if err != nil {
		return nil, err
	}
	p, err := influx.NewPoint("electricity-meter", nil, getFields(req), t)
	if err != nil {
		return nil, err
	}
	bp.AddPoint(p)
	err = h.influxClient.Write(bp)
	if err != nil {
		return nil, err
	}
	return &api.MeterDataReply{}, nil
}

func getFields(req *api.MeterData) map[string]interface{} {
	f := make(map[string]interface{})
	f["active-power-plus"] = req.GetActivePowerPlus()
	// TODO: add other fields
	return f
}
