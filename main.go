//process cpu  and memory usage
package main

import (
	// "fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"time"

	"github.com/struCoder/pidusage"

	ps "github.com/mitchellh/go-ps"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type ProcessList struct {
	name     string
	pid      int
	cpu_load float64
	mem_load float64
}

var (
	proclists []ProcessList
	location  = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cpu_usage",
			Help: "process load cpu-mem",
		},
		[]string{"process_name"},
	)
	interval     = 180 //second
	num_p        = 10  // first 10 processes
	current_name string
)

func main() {
	split_args := strings.Split(os.Args[0], "/")
	current_name = split_args[len(split_args)-1]
	prometheus.MustRegister(location)
	http.Handle("/metrics", promhttp.Handler())
	go runSrv()
	log.Fatal(http.ListenAndServe(":9301", nil)) // listen port 9301

}
func runSrv() {
	for {
		processList, err := ps.Processes()
		if err != nil {
			log.Println("ps.Processes() Failed")
			return
		}
		for x := range processList {
			var process ps.Process
			process = processList[x]
			sysInfo, err := pidusage.GetStat(process.Pid())
			if err != nil {
				continue
			}
			if process.Executable() == current_name {
				continue
			}
			if sysInfo.CPU < 1 {
				continue
			}
			myprocessInfo := ProcessList{name: process.Executable(), cpu_load: sysInfo.CPU, mem_load: sysInfo.Memory}
			proclists = append(proclists, myprocessInfo)
		}
		sort.Slice(proclists[:], func(i, j int) bool {
			return proclists[i].cpu_load > proclists[j].cpu_load
		})
		location.Reset()
		for i, a := range proclists {
			location.With(prometheus.Labels{"process_name": a.name}).Set(float64(a.cpu_load))
			if i == num_p {
				proclists = nil
				break
			}
		}
		proclists = nil
		time.Sleep(time.Second * time.Duration(interval))
		//
	}
}
