package cron

import (
	"github.com/open-falcon/agent/funcs"
	"github.com/open-falcon/agent/g"
	"log"
	"time"
)

func InitDataHistory() {
	for {
		funcs.UpdateCpuStat()
		funcs.UpdateDiskStats()
		time.Sleep(g.COLLECT_INTERVAL)
	}
}

func Collect() {

	if !g.Config().Transfer.Enabled {
		return
	}

	if g.Config().Transfer.Addr == "" {
		return
	}

	for _, v := range funcs.Mappers {
		go collect(int64(v.Interval), v.Fs)
	}
}

func collect(sec int64, fns []func() []*g.MetricValue) {

	for {
	REST:
		time.Sleep(time.Duration(sec) * time.Second)

		hostname, err := g.Hostname()
		if err != nil {
			log.Println("g.Hostname() fail:", err)
			goto REST
		}

		mvs := []*g.MetricValue{}
		ignoreMetrics := g.Config().IgnoreMetrics

		for _, fn := range fns {
			items := fn()
			if items == nil {
				continue
			}

			if len(items) == 0 {
				continue
			}

			for _, mv := range items {
				if _, ok := ignoreMetrics[mv.Metric]; ok {
					continue
				} else {
					mvs = append(mvs, mv)
				}
			}
		}

		now := time.Now().Unix()
		for j := 0; j < len(mvs); j++ {
			mvs[j].Step = sec
			mvs[j].Endpoint = hostname
			mvs[j].Timestamp = now
		}

		g.SendToTransfer(mvs)

	}
}
