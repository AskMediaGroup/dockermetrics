package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
)

type fmtGraphite struct {
	hostname string
}

func NewGraphite() Formatter {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	return &fmtGraphite{hostname: hostname}
}

// Format metrics and write to stdout
func (s *fmtGraphite) Fmt(stats *docker.Stats, container *docker.Container, image string) {
	id := container.ID[:12]
	name := container.Name[1:]
	stime := stats.Read.Unix()
	metricbase := fmt.Sprintf("metrics.%s.%s.%s.%s", s.hostname, id, name, image)

	// network stats
	fmt.Printf("%s.network.rxdropped %d %d\n", metricbase,
		stats.Network.RxDropped, stime)
	fmt.Printf("%s.network.rxbytes %d %d\n", metricbase,
		stats.Network.RxBytes, stime)
	fmt.Printf("%s.network.rxerrors %d %d\n", metricbase,
		stats.Network.RxErrors, stime)
	fmt.Printf("%s.network.rxpackets %d %d\n", metricbase,
		stats.Network.RxPackets, stime)
	fmt.Printf("%s.network.txdropped %d %d\n", metricbase,
		stats.Network.TxDropped, stime)
	fmt.Printf("%s.network.txbytes %d %d\n", metricbase,
		stats.Network.TxBytes, stime)
	fmt.Printf("%s.network.txerrors %d %d\n", metricbase,
		stats.Network.TxErrors, stime)
	fmt.Printf("%s.network.txpackets %d %d\n", metricbase,
		stats.Network.TxPackets, stime)

	// block stats
	for _, v := range stats.BlkioStats.IOServiceBytesRecursive {
		fmt.Printf("%s.block.servicebytes.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.IOServicedRecursive {
		fmt.Printf("%s.block.serviced.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.IOQueueRecursive {
		fmt.Printf("%s.block.queue.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.IOServiceTimeRecursive {
		fmt.Printf("%s.block.servicetime.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.IOWaitTimeRecursive {
		fmt.Printf("%s.block.waittime.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.IOMergedRecursive {
		fmt.Printf("%s.block.merged.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.IOTimeRecursive {
		fmt.Printf("%s.block.time.%d-%d %d %d\n", metricbase,
			v.Major, v.Minor, v.Value, stime)
	}
	for _, v := range stats.BlkioStats.SectorsRecursive {
		fmt.Printf("%s.block.sectors.%d-%d.%s %d %d\n", metricbase,
			v.Major, v.Minor, v.Op, v.Value, stime)
	}

	// memory stats
	fmt.Printf("%s.memory.totalpgmafault %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalPgmafault, stime)
	fmt.Printf("%s.memory.cache %d %d\n", metricbase,
		stats.MemoryStats.Stats.Cache, stime)
	fmt.Printf("%s.memory.mappedfile %d %d\n", metricbase,
		stats.MemoryStats.Stats.MappedFile, stime)
	fmt.Printf("%s.memory.totalinactivefile %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalInactiveFile, stime)
	fmt.Printf("%s.memory.pgpgout %d %d\n", metricbase,
		stats.MemoryStats.Stats.Pgpgout, stime)
	fmt.Printf("%s.memory.rss %d %d\n", metricbase,
		stats.MemoryStats.Stats.Rss, stime)
	fmt.Printf("%s.memory.totalmappedfile %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalMappedFile, stime)
	fmt.Printf("%s.memory.writeback %d %d\n", metricbase,
		stats.MemoryStats.Stats.Writeback, stime)
	fmt.Printf("%s.memory.unexvictable %d %d\n", metricbase,
		stats.MemoryStats.Stats.Unevictable, stime)
	fmt.Printf("%s.memory.pgpgin %d %d\n", metricbase,
		stats.MemoryStats.Stats.Pgpgin, stime)
	fmt.Printf("%s.memory.totalunevictable %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalUnevictable, stime)
	fmt.Printf("%s.memory.pgmajfault %d %d\n", metricbase,
		stats.MemoryStats.Stats.Pgmajfault, stime)
	fmt.Printf("%s.memory.totalrss %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalRss, stime)
	fmt.Printf("%s.memory.totalrsshuge %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalRssHuge, stime)
	fmt.Printf("%s.memory.totalwriteback %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalWriteback, stime)
	fmt.Printf("%s.memory.totalinactiveannon %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalInactiveAnon, stime)
	fmt.Printf("%s.memory.rsshuge %d %d\n", metricbase,
		stats.MemoryStats.Stats.RssHuge, stime)
	fmt.Printf("%s.memory.hierarchicalmemorylimit %d %d\n", metricbase,
		stats.MemoryStats.Stats.HierarchicalMemoryLimit, stime)
	fmt.Printf("%s.memory.totalpgfault %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalPgfault, stime)
	fmt.Printf("%s.memory.totalactivefile %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalActiveFile, stime)
	fmt.Printf("%s.memory.activeanon %d %d\n", metricbase,
		stats.MemoryStats.Stats.ActiveAnon, stime)
	fmt.Printf("%s.memory.totalactiveanon %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalActiveAnon, stime)
	fmt.Printf("%s.memory.totalpgpgout %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalPgpgout, stime)
	fmt.Printf("%s.memory.totalcache %d %d\n", metricbase,
		stats.MemoryStats.Stats.TotalCache, stime)
	fmt.Printf("%s.memory.inactiveanon %d %d\n", metricbase,
		stats.MemoryStats.Stats.InactiveAnon, stime)
	fmt.Printf("%s.memory.activefile %d %d\n", metricbase,
		stats.MemoryStats.Stats.ActiveFile, stime)
	fmt.Printf("%s.memory.pgfault %d %d\n", metricbase,
		stats.MemoryStats.Stats.Pgfault, stime)
	fmt.Printf("%s.memory.inactivefile %d %d\n", metricbase,
		stats.MemoryStats.Stats.InactiveFile, stime)
	fmt.Printf("%s.memory.maxusage %d %d\n", metricbase,
		stats.MemoryStats.MaxUsage, stime)
	fmt.Printf("%s.memory.usage %d %d\n", metricbase,
		stats.MemoryStats.Usage, stime)
	fmt.Printf("%s.memory.failcnt %d %d\n", metricbase,
		stats.MemoryStats.Failcnt, stime)
	fmt.Printf("%s.memory.limit %d %d\n", metricbase,
		stats.MemoryStats.Limit, stime)

	// cpu stats
	fmt.Printf("%s.cpu.systemcpuusage %d %d\n", metricbase,
		stats.CPUStats.SystemCPUUsage, stime)
	fmt.Printf("%s.cpu.cpuusage.usageinusermode %d %d\n", metricbase,
		stats.CPUStats.CPUUsage.UsageInUsermode, stime)
	fmt.Printf("%s.cpu.cpuusage.usageinkernelmode %d %d\n", metricbase,
		stats.CPUStats.CPUUsage.UsageInKernelmode, stime)
	fmt.Printf("%s.cpu.cpuusage.totalusage %d %d\n", metricbase,
		stats.CPUStats.CPUUsage.TotalUsage, stime)
	for i, usage := range stats.CPUStats.CPUUsage.PercpuUsage {
		fmt.Printf("%s.cpu.cpuusage.percpuusage.%d %d %d\n", metricbase, i, usage, stime)
	}
}
