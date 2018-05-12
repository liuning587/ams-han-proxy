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
	if l2 := req.GetList2(); l2 != nil {
		f["active-power-minus"] = l2.GetActivePowerMinus()
		f["reactive-power-plus"] = l2.GetReactivePowerPlus()
		f["reactive-power-minus"] = l2.GetReactivePowerMinus()
		f["phase-current"] = l2.GetPhaseCurrent()
		f["phase-voltage"] = l2.GetPhaseVoltage()
	}
	if l3 := req.GetList3(); l3 != nil {
		f["cumulative-hourly-active-import-energy"] = l3.GetCumulativeHourlyActiveImportEnergy()
		f["cumulative-hourly-active-export-energy"] = l3.GetCumulativeHourlyActiveExportEnergy()
		f["cumulative-hourly-reactive-import-energy"] = l3.GetCumulativeHourlyReactiveImportEnergy()
		f["cumulative-hourly-reactive-export-energy"] = l3.GetCumulativeHourlyReactiveExportEnergy()
	}
	return f
}
