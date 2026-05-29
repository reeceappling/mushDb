package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type SensorData struct {
	Time     float64 `json:"time"` // EpochTimestampMs
	Temp     float64 `json:"temp"`
	RH       float64 `json:"rh"`
	Lux      float64 `json:"lux"`      // TODO: is this ok?
	Airspeed float64 `json:"airspeed"` // TODO: ft/s? m/s
}

type GetSensorDataReq struct {
	// TODO: this
}
type GetSensorDataResp struct {
	IsAggregate bool // If true, pass min/max/avg
	Data        any  // TODO: fix
}

type AddSensorDataReq struct {
	auth string       // TODO: ok????
	Data []SensorData // TODO: fix
}

func GetSensorDataHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nodeName := r.PathValue("nodeName")
		// TODO: this
		// TODO: check timespan/discretization
		// TODO: get datapoints
		// TODO: discretize if needed
	})
}

func GetSensorDataSinceHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//nodeName := r.PathValue("nodeName")
		// TODO: this
		// TODO: get all datapoints since the last one
		// TODO: return all those datapoints
	})
}

func AddSensorDataHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		bs, err := io.ReadAll(r.Body)
		if err != nil {
			// TODO: something
		}
		var req AddSensorDataReq
		err = json.Unmarshal(bs, &req)
		if err != nil {
			// TODO: something
		}
		// TODO: validate caller is authorized to add Data
		nodeName := r.PathValue("nodeName")
		err = addSensorDataFor(nodeName, req.Data)
		if err != nil {
			// TODO: something here
		}
		_, err = w.Write([]byte("OK"))
		if err != nil {
			// TODO: something here
		}
	})
}

func addSensorDataFor(nodeName string, data []SensorData) error {

	// TODO: this! Add sensor Data to db (or file?)
	return errors.New("NOT IMPLEMENTED")
}
