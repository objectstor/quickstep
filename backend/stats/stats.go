package qstats

import (
	"errors"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

//Qmetric - metric struct
type Qmetric struct {
	Gauge      *prometheus.Gauge
	Counter    *prometheus.Counter
	Registered bool
}

// QStat stat structure
type QStat struct {
	Metrics map[string]Qmetric
}

//New Qstat
func New() *QStat {
	q := new(QStat)
	q.Metrics = make(map[string]Qmetric)
	return q
}

//AddGauge add Gauge to metrics
func (q *QStat) AddGauge(ShortName string, Name string, Help string) error {
	_, ok := q.Metrics[ShortName]
	if ok {
		msg := fmt.Sprintf("Metrics %s not exists", ShortName)
		return errors.New(msg)
	}

	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: Name,
		Help: Help,
	})
	metricts := Qmetric{&gauge, nil, false}
	q.Metrics[ShortName] = metricts
	return nil
}

//AddCounter to metrics
func (q *QStat) AddCounter(ShortName string, Name string, Help string) error {
	_, ok := q.Metrics[ShortName]
	if ok {
		msg := fmt.Sprintf("Metrics %s not exists", ShortName)
		return errors.New(msg)
	}
	counter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: Name,
		Help: Help,
	})
	metricts := Qmetric{nil, &counter, false}
	q.Metrics[ShortName] = metricts
	return nil
}
func (q *QStat) checkShortName(ShortName string) (Qmetric, error) {
	metric, ok := q.Metrics[ShortName]
	if !ok {
		msg := fmt.Sprintf("Metrics %s not exists", ShortName)
		return Qmetric{nil, nil, false}, errors.New(msg)
	}
	return metric, nil
}

//Register - registetr counters
func (q *QStat) Register() (int, error) {
	counter := 0
	for _, metricValue := range q.Metrics {
		if !metricValue.Registered {
			if metricValue.Gauge != nil {
				prometheus.MustRegister(*metricValue.Gauge)
				metricValue.Registered = true
				counter++
			}
			if metricValue.Counter != nil {
				prometheus.MustRegister(*metricValue.Counter)
				metricValue.Registered = true
				counter++
			}
		}
	}
	return counter, nil
}

//Inc increment named gauge
func (q *QStat) Inc(ShortName string) error {
	metric, err := q.checkShortName(ShortName)
	if err != nil {
		return err
	}
	if metric.Gauge != nil {
		m := *metric.Gauge
		m.Inc()
	}
	if metric.Counter != nil {
		c := *metric.Counter
		c.Inc()
	}
	return nil
}

//Dec increment named gauge
func (q *QStat) Dec(ShortName string) error {
	metric, err := q.checkShortName(ShortName)
	if err != nil {
		return err
	}
	if metric.Gauge != nil {
		m := *metric.Gauge
		m.Dec()
		return nil
	}
	msg := fmt.Sprintf("Metrics %s is not gauge", ShortName)
	return errors.New(msg)
}

//Set set counbter value
func (q *QStat) Set(ShortName string, value float64) error {
	metric, err := q.checkShortName(ShortName)
	if err != nil {
		return err
	}
	if metric.Gauge != nil {
		m := *metric.Gauge
		m.Set(value)
		return nil
	}
	msg := fmt.Sprintf("Metrics %s is not gauge", ShortName)
	return errors.New(msg)
}

//Add set counbter value
func (q *QStat) Add(ShortName string, value float64) error {
	metric, err := q.checkShortName(ShortName)
	if err != nil {
		return err
	}
	if metric.Gauge != nil {
		m := *metric.Gauge
		m.Add(value)
	}
	if metric.Counter != nil {
		if value < 0 {
			msg := fmt.Sprintf("Metrics %s value less than 0 %f", ShortName, value)
			return errors.New(msg)
		}
		c := *metric.Counter
		c.Add(value)
	}
	return nil
}

//Sub set counbter value
func (q *QStat) Sub(ShortName string, value float64) error {
	metric, err := q.checkShortName(ShortName)
	if err != nil {
		return err
	}
	if metric.Gauge != nil {
		m := *metric.Gauge
		m.Sub(value)
		return nil
	}
	msg := fmt.Sprintf("Metrics %s is not gauge", ShortName)
	return errors.New(msg)
}
