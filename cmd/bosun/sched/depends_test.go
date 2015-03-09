package sched

import (
	"bosun.org/opentsdb"
	"testing"
)

// Crit returns {a=b},{a=c}, but {a=b} is ignored by dependency expression.
// Result should be {a=c} only.
func TestDependency_Simple(t *testing.T) {
	testSched(t, &schedTest{
		conf: `alert a {
			crit = avg(q("avg:c{a=*}", "5m", "")) > 0
			depends = avg(q("avg:d{a=*}", "5m", "")) > 0
		}`,
		queries: map[string]opentsdb.ResponseSet{
			`q("avg:c{a=*}", ` + window5Min + `)`: {
				{
					Metric: "c",
					Tags:   opentsdb.TagSet{"a": "b"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
				{
					Metric: "c",
					Tags:   opentsdb.TagSet{"a": "c"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
			},
			`q("avg:d{a=*}", ` + window5Min + `)`: {
				{
					Metric: "d",
					Tags:   opentsdb.TagSet{"a": "b"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
				{
					Metric: "d",
					Tags:   opentsdb.TagSet{"a": "c"},
					DPS:    map[string]opentsdb.Point{"0": 0},
				},
			},
		},
		state: map[schedState]bool{
			schedState{"a{a=c}", "critical"}: true,
		},
	})
}

// Crit and depends don't have same tag sets.
func TestDependency_Overlap(t *testing.T) {
	testSched(t, &schedTest{
		conf: `alert a {
			crit = avg(q("avg:c{a=*,b=*}", "5m", "")) > 0
			depends = avg(q("avg:d{a=*,d=*}", "5m", "")) > 0
		}`,
		queries: map[string]opentsdb.ResponseSet{
			`q("avg:c{a=*,b=*}", ` + window5Min + `)`: {
				{
					Metric: "c",
					Tags:   opentsdb.TagSet{"a": "b", "b": "r"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
				{
					Metric: "c",
					Tags:   opentsdb.TagSet{"a": "b", "b": "z"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
				{
					Metric: "c",
					Tags:   opentsdb.TagSet{"a": "c", "b": "q"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
			},
			`q("avg:d{a=*,d=*}", ` + window5Min + `)`: {
				{
					Metric: "d",
					Tags:   opentsdb.TagSet{"a": "b", "d": "q"}, //this matches first and second datapoints from crit.
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
			},
		},
		state: map[schedState]bool{
			schedState{"a{a=c,b=q}", "critical"}: true,
		},
	})
}

func TestDependency_OtherAlert(t *testing.T) {
	testSched(t, &schedTest{
		conf: `alert a {
			crit = avg(q("avg:a{host=*,cpu=*}", "5m", "")) > 0
		}
		alert b{
			depends = alert("a","crit")
			crit = avg(q("avg:b{host=*}", "5m", "")) > 0
		}
		alert c{
			crit = avg(q("avg:b{host=*}", "5m", "")) > 0
		}
		alert d{
			#b will be unevaluated because of a.
			depends = alert("b","crit")
			crit = avg(q("avg:b{host=*}", "5m", "")) > 0
		}
		`,
		queries: map[string]opentsdb.ResponseSet{
			`q("avg:a{cpu=*,host=*}", ` + window5Min + `)`: {
				{
					Metric: "a",
					Tags:   opentsdb.TagSet{"host": "ny01", "cpu": "0"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
			},
			`q("avg:b{host=*}", ` + window5Min + `)`: {
				{
					Metric: "b",
					Tags:   opentsdb.TagSet{"host": "ny01"},
					DPS:    map[string]opentsdb.Point{"0": 1},
				},
			},
		},
		state: map[schedState]bool{
			schedState{"a{cpu=0,host=ny01}", "critical"}: true,
			schedState{"c{host=ny01}", "critical"}:       true,
		},
	})
}
