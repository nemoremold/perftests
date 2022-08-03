package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/pflag"
	clientv3 "go.etcd.io/etcd/client/v3"
	"k8s.io/klog"
)

func main() {
	var (
		endpoints []string
		testCount int
	)

	pflag.StringArrayVarP(&endpoints, "endpoints", "e", nil, "etcd endpoints")
	pflag.IntVarP(&testCount, "test-count", "c", 100, "number of tests")
	pflag.Parse()

	c, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		klog.Fatal(err)
	}

	startTime := time.Now()
	for i := 0; i < testCount; i++ {
		for {
			_, err := c.Put(context.Background(), fmt.Sprintf("put-%v", i), fmt.Sprintf("%v-%v", i, time.Now().UnixMilli()))
			if err == nil {
				klog.Infof("%.2f%% (%v/%v) done.", float64(i)/float64(testCount), i, testCount)
				break
			} else {
				klog.Errorf("Failed to put put-%v", i)
			}
		}
	}
	duration := time.Since(startTime)
	putsPerSecond := float64(testCount) / duration.Seconds()

	klog.Infof("Test duration:   %v", duration)
	klog.Infof("Puts per second: %v", putsPerSecond)

	for i := 0; i < testCount; i++ {
		for {
			_, err := c.Delete(context.Background(), fmt.Sprintf("put-%v", i))
			if err == nil {
				break
			} else {
				klog.Errorf("Failed to delete put-%v", i)
			}
		}
	}
}
