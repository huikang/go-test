package mytest

import (
	"fmt"
	"net"
	"testing"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
)

func TestGlobalMetrics(t *testing.T) {
	var sinks metrics.FanoutSink

	addr := "dogstatss:8125"
	hostname := "MacBook-Pro-2.home"
	fmt.Println("dogstatdSink addr:", addr)
	dogSink, err := datadog.NewDogStatsdSink(addr, hostname)
	if err != nil {
		fmt.Println("error NewDogStatsdSink", err)

		if netErr, ok := err.(net.Error); ok {
			fmt.Println("net err", netErr)
		}
		// return
	}

	if dogSink != nil {
		sinks = append(sinks, dogSink)
		metrics.NewGlobal(metrics.DefaultConfig("service-name"), sinks)
		sendSample()
	}

	fmt.Println("second dogsink")
	addr = "dogstats:8125"
	dogSink, err = datadog.NewDogStatsdSink(addr, hostname)
	if err != nil {
		fmt.Println("error NewDogStatsdSink", err)

		if netErr, ok := err.(net.Error); ok {
			fmt.Println("net err", netErr)
		}
		return
	}
	sinks = append(sinks, dogSink)
	metrics.NewGlobal(metrics.DefaultConfig("service-name"), sinks)
	sendSample()
}

func sendSample() {
	for i := 0; i < 50; i++ {
		// sink.IncrCounter([]string{"testkey", "name"}, float32(4))
		metrics.AddSampleWithLabels([]string{"testkey", "name"}, float32(4), []metrics.Label{{"tagkey", "tagvalue"}})
	}
}

func TestDogMetrics(t *testing.T) {
	addr := "dogstats:8125"
	hostname := "MacBook-Pro-2.home"
	fmt.Println("dogstatdSink addr:", addr)
	sink, err := datadog.NewDogStatsdSink(addr, hostname)
	if err != nil {
		fmt.Println("error NewDogStatsdSink", err)

		if netErr, ok := err.(net.Error); ok {
			fmt.Println("net err", netErr)
		}
		return
	}

	/*
		metric, err := metrics.NewGlobal(metrics.DefaultConfig("service-name"), sink)
		if err != nil {
			return
		}
		metric.IncrCounter([]string{"testkey", "name"}, float32(4))
	*/

	for i := 0; i < 50; i++ {
		// sink.IncrCounter([]string{"testkey", "name"}, float32(4))
		sink.AddSampleWithLabels([]string{"testkey", "name"}, float32(4), []metrics.Label{{"tagkey", "tagvalue"}})
	}
	// sink.Shutdown()
}
