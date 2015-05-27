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
		"timestamp",
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
		stats.Read.UnixNano(),
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
