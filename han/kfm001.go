package han

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"svenschwermer.de/ams-han-proxy/cosem"
	api "svenschwermer.de/ams-han-proxy/proto/electricity"
)

type Handler struct {
	sink api.MeterSinkClient
}

func NewHandler(cc *grpc.ClientConn) *Handler {
	return &Handler{sink: api.NewMeterSinkClient(cc)}
}

func (h *Handler) DecodeLLCPayload(data []byte) error {
	t, err := cosem.DecodeTelegram(data)
	if err != nil {
		return fmt.Errorf("Failed to decode telegram (%s): %s", hex.EncodeToString(data), err)
	}
	if t.NumItems() != 6 {
		return fmt.Errorf("Expected 6 items, got %d: %v", t.NumItems(), t)
	}
	s, ok := t.Item(5).(cosem.Structure)
	if !ok {
		return fmt.Errorf("Expected structure, got %T (Telegram: %v)", t.Item(5), t)
	}

	md := &api.MeterData{
		HostTimestamp: ptypes.TimestampNow(),
		Header:        &api.MeterDataHeader{},
	}
	if i, ok := t.Item(0).(cosem.Integer); ok {
		md.Header.X = int32(i)
	}
	if err := getTime(t.Item(4), &md.Header.Timestamp); err != nil {
		return err
	}

	switch s.NumItems() {
	case 1:
		err = handleList1(s, md)
	case 9:
		md.List2 = &api.MeterDataList2{}
		err = handleList2(s, md)
	case 14:
		md.List3 = &api.MeterDataList3{}
		err = handleList3(s, md)
	default:
		err = fmt.Errorf("Unexpected structure in telegram (%v): %v", t, s)
	}
	if err != nil {
		return err
	}
	log.Debugf("Publishing %+v", *md)
	_, err = h.sink.Publish(context.Background(), md)
	return err
}

func handleList1(s cosem.Structure, md *api.MeterData) error {
	return getFloat64(s.Item(0), &md.ActivePowerPlus)
}

func handleList2(s cosem.Structure, md *api.MeterData) error {
	vars := []func() error{
		func() error { return getString(s.Item(0), &md.List2.ObisListVersionId) },
		func() error { return getString(s.Item(1), &md.List2.MeterId) },
		func() error { return getString(s.Item(2), &md.List2.MeterType) },
		func() error { return getFloat64(s.Item(3), &md.ActivePowerPlus) },
		func() error { return getFloat64(s.Item(4), &md.List2.ActivePowerMinus) },
		func() error { return getFloat64(s.Item(5), &md.List2.ReactivePowerPlus) },
		func() error { return getFloat64(s.Item(6), &md.List2.ReactivePowerMinus) },
		func() error { return getFloat64(s.Item(7), &md.List2.PhaseCurrent) },
		func() error { return getFloat64(s.Item(8), &md.List2.PhaseVoltage) },
	}
	for _, v := range vars {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}

func handleList3(s cosem.Structure, md *api.MeterData) error {
	if err := handleList2(s, md); err != nil {
		return err
	}
	vars := []func() error{
		func() error { return getTime(s.Item(9), &md.List3.MeterTimestamp) },
		func() error { return getFloat64(s.Item(10), &md.List3.CumulativeHourlyActiveImportEnergy) },
		func() error { return getFloat64(s.Item(11), &md.List3.CumulativeHourlyActiveExportEnergy) },
		func() error { return getFloat64(s.Item(12), &md.List3.CumulativeHourlyReactiveImportEnergy) },
		func() error { return getFloat64(s.Item(13), &md.List3.CumulativeHourlyReactiveExportEnergy) },
	}
	for _, v := range vars {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}

func getFloat64(src cosem.Data, dest *float64) error {
	i, ok := src.(cosem.Integer)
	if !ok {
		return fmt.Errorf("Item is no integer, but %T", src)
	}
	*dest = float64(i)
	return nil
}

func getString(src cosem.Data, dest *string) error {
	s, ok := src.(cosem.OctetString)
	if !ok {
		return fmt.Errorf("Item is no octet string, but %T", src)
	}
	*dest = string(s)
	return nil
}

func getTime(src cosem.Data, dest **timestamp.Timestamp) error {
	os, ok := src.(cosem.OctetString)
	if !ok {
		return fmt.Errorf("Item is no octet string (for conversion to timestamp), but %T", src)
	}
	t, err := os.AsDateTime()
	if err != nil {
		return err
	}
	*dest, err = ptypes.TimestampProto(time.Time(t))
	return err
}
