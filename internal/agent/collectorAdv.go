package agent

import (
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
)

func (a *agent) collectorAdv() {
	for {
		vmStat, _ := mem.VirtualMemory()
		cpuPercents, _ := cpu.Percent(0, false)

		a.Mutex.Lock()
		a.Metrics["TotalMemory"] = float64(vmStat.Total)
		a.Metrics["FreeMemory"] = float64(vmStat.Free)
		if len(cpuPercents) > 0 {
			a.Metrics["CPUutilization1"] = cpuPercents[0]
		}
		a.Mutex.Unlock()

		time.Sleep(time.Duration(a.PollInterval) * time.Second)
	}
}
