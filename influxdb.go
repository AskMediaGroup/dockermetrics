package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/influxdb/influxdb/client"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func getEnv(name, def string) string {
	if val := os.Getenv(name); val != "" {
		return val
	}
	return def
}

// XXX: currently summing the values for all block devices attached to each container
func readBlockStat(stats []docker.BlkioStatsEntry) []interface{} {
	z := uint64(0)
	entries := []interface{}{z, z, z, z, z}
	for _, v := range stats {
		switch v.Op {
		case "Read":
			entries[0] = entries[0].(uint64) + v.Value
		case "Write":
			entries[1] = entries[1].(uint64) + v.Value
		case "Sync":
			entries[2] = entries[2].(uint64) + v.Value
		case "Async":
			entries[3] = entries[3].(uint64) + v.Value
		case "Total":
			entries[4] = entries[4].(uint64) + v.Value
		}
	}
	return entries
}

type fmtInfluxdb struct {
	machinename string
	database    string
	table       string
	columns     *[]string
	client      *client.Client
	series      []*client.Series
	sync.Mutex
}

func NewInfluxdb() Formatter {
	// TODO: figure out a better way to include formatter options than env variables
	host := getEnv("INFLUXDB_HOST", "")
	if host == "" {
		log.Fatal("Missing influxdb host")
	}
	port, err := strconv.Atoi(getEnv("INFLUXDB_PORT", "8086"))
	if err != nil {
		log.Fatal(err)
	}
	user := getEnv("INFLUXDB_USER", "root")
	passwd := getEnv("INFLUXDB_PASSWD", "root")
	db := getEnv("INFLUXDB_DATABASE", "stats")
	series := getEnv("INFLUXDB_SERIES", "container_metrics")
	write_interval, err := strconv.Atoi(getEnv("INFLUXDB_WRITE_INTERVAL", "1"))
	if err != nil {
		log.Fatal(err)
	}
	machinename, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	u := fmt.Sprintf("%s:%d", host, port)
	if err != nil {
		log.Fatal(err)
	}
	influxdbclient, err := client.NewClient(&client.ClientConfig{
		Host:     u,
		Username: user,
		Password: passwd,
		Database: db,
	})
	if err != nil {
		log.Fatal(err)
	}
	columns := &[]string{
		"time",
		"machine_name",
		"container_name",
		"container_id",
		"container_image",
		"network_rx_dropped",
		"network_rx_bytes",
		"network_rx_errors",
		"network_rx_packets",
		"network_tx_dropped",
		"network_tx_bytes",
		"network_tx_errors",
		"network_tx_packets",
		"memory_totalpgmafault",
		"memory_cache",
		"memory_mappedfile",
		"memory_totalinactivefile",
		"memory_pgpgout",
		"memory_rss",
		"memory_totalmappedfile",
		"memory_writeback",
		"memory_unexvictable",
		"memory_pgpgin",
		"memory_totalunevictable",
		"memory_pgmajfault",
		"memory_totalrss",
		"memory_totalrsshuge",
		"memory_totalwriteback",
		"memory_totalinactiveannon",
		"memory_rsshuge",
		"memory_hierarchicalmemorylimit",
		"memory_totalpgfault",
		"memory_totalactivefile",
		"memory_activeanon",
		"memory_totalactiveanon",
		"memory_totalpgpgout",
		"memory_totalcache",
		"memory_inactiveanon",
		"memory_activefile",
		"memory_pgfault",
		"memory_inactivefile",
		"memory_maxusage",
		"memory_usage",
		"memory_failcnt",
		"memory_limit",
		"cpu_system_cpu_usage",
		"cpu_usage_in_usermode",
		"cpu_usage_in_kernelmode",
		"cpu_total_usage",
		"block_service_bytes_read",
		"block_service_bytes_write",
		"block_service_bytes_sync",
		"block_service_bytes_async",
		"block_service_bytes_total",
		"block_serviced_read",
		"block_serviced_write",
		"block_serviced_sync",
		"block_serviced_async",
		"block_serviced_total",
		"block_queue_read",
		"block_queue_write",
		"block_queue_sync",
		"block_queue_async",
		"block_queue_total",
		"block_service_time_read",
		"block_service_time_write",
		"block_service_time_sync",
		"block_service_time_async",
		"block_service_time_total",
		"block_wait_time_read",
		"block_wait_time_write",
		"block_wait_time_sync",
		"block_wait_time_async",
		"block_wait_time_total",
		"block_merged_read",
		"block_merged_write",
		"block_merged_sync",
		"block_merged_async",
		"block_merged_total",
		"block_time",
		"block_sectors",
	}

	fmt := &fmtInfluxdb{
		machinename: machinename,
		client:      influxdbclient,
		columns:     columns,
		table:       series,
		series:      make([]*client.Series, 0),
	}
	ticker := time.NewTicker(time.Duration(write_interval) * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Write()
			}
		}
	}()
	return fmt
}

// Write buffered metrics to influxdb
func (f *fmtInfluxdb) Write() {
	f.Lock()
	defer f.Unlock()

	if len(f.series) < 1 {
		return
	}

	err := f.client.WriteSeriesWithTimePrecision(f.series, client.Microsecond)
	if err != nil {
		log.Print(err)
	}
	f.series = make([]*client.Series, 0)
}

// Format stats and add to write buffer
func (f *fmtInfluxdb) Fmt(stats *docker.Stats, container *docker.Container, image string) {
	id := container.ID[:12]
	name := container.Name[1:]
	point := []interface{}{
		// general metadata
		int64(time.Duration(stats.Read.UnixNano()) / time.Microsecond),
		f.machinename,
		name,
		id,
		image,
		// network stats
		stats.Network.RxDropped,
		stats.Network.RxBytes,
		stats.Network.RxErrors,
		stats.Network.RxPackets,
		stats.Network.TxDropped,
		stats.Network.TxBytes,
		stats.Network.TxErrors,
		stats.Network.TxPackets,
		// memory stats
		stats.MemoryStats.Stats.TotalPgmafault,
		stats.MemoryStats.Stats.Cache,
		stats.MemoryStats.Stats.MappedFile,
		stats.MemoryStats.Stats.TotalInactiveFile,
		stats.MemoryStats.Stats.Pgpgout,
		stats.MemoryStats.Stats.Rss,
		stats.MemoryStats.Stats.TotalMappedFile,
		stats.MemoryStats.Stats.Writeback,
		stats.MemoryStats.Stats.Unevictable,
		stats.MemoryStats.Stats.Pgpgin,
		stats.MemoryStats.Stats.TotalUnevictable,
		stats.MemoryStats.Stats.Pgmajfault,
		stats.MemoryStats.Stats.TotalRss,
		stats.MemoryStats.Stats.TotalRssHuge,
		stats.MemoryStats.Stats.TotalWriteback,
		stats.MemoryStats.Stats.TotalInactiveAnon,
		stats.MemoryStats.Stats.RssHuge,
		stats.MemoryStats.Stats.HierarchicalMemoryLimit,
		stats.MemoryStats.Stats.TotalPgfault,
		stats.MemoryStats.Stats.TotalActiveFile,
		stats.MemoryStats.Stats.ActiveAnon,
		stats.MemoryStats.Stats.TotalActiveAnon,
		stats.MemoryStats.Stats.TotalPgpgout,
		stats.MemoryStats.Stats.TotalCache,
		stats.MemoryStats.Stats.InactiveAnon,
		stats.MemoryStats.Stats.ActiveFile,
		stats.MemoryStats.Stats.Pgfault,
		stats.MemoryStats.Stats.InactiveFile,
		stats.MemoryStats.MaxUsage,
		stats.MemoryStats.Usage,
		stats.MemoryStats.Failcnt,
		stats.MemoryStats.Limit,
		// cpu stats
		stats.CPUStats.SystemCPUUsage,
		stats.CPUStats.CPUUsage.UsageInUsermode,
		stats.CPUStats.CPUUsage.UsageInKernelmode,
		stats.CPUStats.CPUUsage.TotalUsage,
	}

	// block stats
	point = append(point, readBlockStat(stats.BlkioStats.IOServiceBytesRecursive)...)
	point = append(point, readBlockStat(stats.BlkioStats.IOServicedRecursive)...)
	point = append(point, readBlockStat(stats.BlkioStats.IOQueueRecursive)...)
	point = append(point, readBlockStat(stats.BlkioStats.IOServiceTimeRecursive)...)
	point = append(point, readBlockStat(stats.BlkioStats.IOWaitTimeRecursive)...)
	point = append(point, readBlockStat(stats.BlkioStats.IOMergedRecursive)...)

	var io_time uint64
	for _, v := range stats.BlkioStats.IOTimeRecursive {
		io_time += v.Value
	}
	point = append(point, io_time)

	var sectors uint64
	for _, v := range stats.BlkioStats.SectorsRecursive {
		sectors += v.Value
	}
	point = append(point, sectors)

	series := &client.Series{
		Name:    f.table,
		Columns: *f.columns,
		Points:  make([][]interface{}, 1),
	}
	series.Points[0] = point
	f.Lock()
	defer f.Unlock()
	f.series = append(f.series, series)
}
