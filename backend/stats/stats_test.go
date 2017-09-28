package qstats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	// test database connection error
	q := New()
	assert.Nil(t, q.AddGauge("cpu_temp_gauge", "cpu_temperature_celsius", "meassuring cpu temperature in C"))
	assert.Error(t, q.AddCounter("cpu_temp_gauge", "cpu_temperature_celsius", "meassuring cpu temperature in C"))
	assert.Error(t, q.AddGauge("cpu_temp_gauge", "cpu_temperature_celsius", "meassuring cpu temperature in C"))
	assert.Nil(t, q.AddGauge("hd_temp_gauge", "hd_temperature_celsius", "meassuring hd temperature in C"))
	assert.Nil(t, q.AddCounter("mb_temp_counter", "mb_temperature_celsius", "meassuring motherboard temperature in C"))
	c, err := q.Register()
	assert.Nil(t, err)
	assert.Equal(t, 3, c)
	assert.Nil(t, q.Inc("hd_temp_gauge"))
	assert.Nil(t, q.Inc("mb_temp_counter"))
	assert.Nil(t, q.Dec("hd_temp_gauge"))
	assert.Error(t, q.Dec("mb_temp_counter"))
	assert.Nil(t, q.Set("hd_temp_gauge", 21.5))
	assert.Error(t, q.Set("mb_temp_counter", 21.5))
	assert.Nil(t, q.Add("hd_temp_gauge", 1.0))
	assert.Nil(t, q.Add("mb_temp_counter", 1.0))
	assert.Nil(t, q.Sub("hd_temp_gauge", 21.5))
	assert.Error(t, q.Sub("mb_temp_counter", 21.5))
	assert.Error(t, q.Sub("dummy", 1.0))
	assert.Error(t, q.Add("dummy", 1.0))
	assert.Error(t, q.Set("dummy", 1.0))
	assert.Error(t, q.Inc("dummy"))
	assert.Error(t, q.Dec("dummy"))
	assert.Error(t, q.Add("mb_temp_counter", -1.0))
}
