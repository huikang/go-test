package mytest

import (
	"context"
	"fmt"
	// "net"
	"testing"

	"github.com/armon/go-metrics"
	"github.com/armon/go-metrics/datadog"
	"github.com/hashicorp/consul/lib/retry"
)

func NewDogstatsdSink(count uint) (*datadog.DogStatsdSink, error) {
	hostname := "MacBook-Pro-2.home"
	addr := "badaddress:8125"
	if count == 2 {
		addr = "dogstats:8125"
	}

	dogSink, err := datadog.NewDogStatsdSink(addr, hostname)
	return dogSink, err
}

/*
    for {
        err := connectTodogStats(w.Failures())
        if err != nil {
            // update global metrics
            metrics
        }

        if err is not DNSError {
            return
        }

        if err := w.Wait(ctx); err != nil {
			fmt.Println("error")
			return
		}
    }
*/

func TestConsulRetryLib(t *testing.T) {

	var sinks metrics.FanoutSink
	var dogSink *datadog.DogStatsdSink
	var err error

	w := &retry.Waiter{}
	ctx := context.Background()
	for {
		fmt.Println("failure times:", w.Failures())

		dogSink, err = NewDogstatsdSink(uint(w.Failures()))
		if err == nil {
			fmt.Println("Created a dogSink")
			break
		}

		if err := w.Wait(ctx); err != nil {
			fmt.Println("error")
			return
		}
	}

	sinks = append(sinks, dogSink)
	metrics.NewGlobal(metrics.DefaultConfig("service-name"), sinks)
	sendSample()
}
