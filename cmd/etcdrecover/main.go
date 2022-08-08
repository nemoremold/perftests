package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/spf13/pflag"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
)

var (
	etcdConnections int
	etcdWatchers    int

	etcdServers []string

	// The short keepalive timeout and interval have been chosen to aggressively
	// detect a failed etcd server without introducing much overhead.
	keepaliveTime    = 30 * time.Second
	keepaliveTimeout = 10 * time.Second

	// dialTimeout is the timeout for failing to establish a connection.
	// It is set to 20 seconds as times shorter than that will cause TLS connections to fail
	// on heavily loaded arm64 CPUs (issue #64649)
	dialTimeout = 20 * time.Second

	healthcheckTimeout = 2
)

func init() {
	pflag.IntVarP(&etcdConnections, "etcd-connections", "c", 1, "number of etcd connections to be created")
	pflag.IntVarP(&etcdWatchers, "etcd-watchers", "w", 1, "number of etcd watchers to be created")
	pflag.StringArrayVarP(&etcdServers, "etcd-servers", "s", nil, "etcd server endpoints")
	pflag.IntVarP(&healthcheckTimeout, "health-check-timeout", "t", 2, "timeout in seconds for etcd healthcheck")
}

func main() {
	pflag.Parse()

	ctx := NewContextWithShutdownSignalHandler()

	mux := http.NewServeMux()
	healthcheck, err := createEtcdConnectionHealthCheck(ctx)
	if err != nil {
		klog.Fatalf(err.Error())
	}
	mux.HandleFunc("/livez", livez(ctx, healthcheck))
	svr := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	go func() {
		<-ctx.Done()
		klog.Info("shutting down health check server")
		_ = svr.Shutdown(context.Background())
	}()

	go runWorkers(ctx)

	klog.Info("starting health check server")
	if err := svr.ListenAndServe(); err != nil {
		klog.Fatal(err)
	}
}

func newEtcdClient() (*clientv3.Client, error) {
	return clientv3.New(clientv3.Config{
		DialTimeout:          dialTimeout,
		DialKeepAliveTime:    keepaliveTime,
		DialKeepAliveTimeout: keepaliveTimeout,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(), // block until the underlying connection is up
			// use chained interceptors so that the default (retry and backoff) interceptors are added.
			// otherwise they will be overwritten by the metric interceptor.
			//
			// these optional interceptors will be placed after the default ones.
			// which seems to be what we want as the metrics will be collected on each attempt (retry)
			grpc.WithChainUnaryInterceptor(grpcprom.UnaryClientInterceptor),
			grpc.WithChainStreamInterceptor(grpcprom.StreamClientInterceptor),
		},
		Endpoints: etcdServers,
	})
}

func livez(ctx context.Context, healthcheck func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := healthcheck()
		if err != nil {
			klog.Infof("etcd connection health check failed: %v", err.Error())
			http.Error(w, fmt.Sprintf("etcd connection health check failed: %v", err.Error()), http.StatusInternalServerError)
			return
		}

		klog.Info("etcd connection health check passed")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if _, found := r.URL.Query()["verbose"]; !found {
			fmt.Fprint(w, "ok")
			return
		}
		fmt.Fprint(w, "etcd connection health check passed.\n")
	}
}

func createEtcdConnectionHealthCheck(ctx context.Context) (func() error, error) {
	// constructing the etcd v3 client blocks and times out if etcd is not available.
	// retry in a loop in the background until we successfully create the client, storing the client or error encountered

	lock := sync.Mutex{}
	var c *clientv3.Client
	clientErr := fmt.Errorf("etcd client connection not yet established")

	go wait.PollUntil(time.Second, func() (bool, error) {
		newClient, err := newEtcdClient()

		lock.Lock()
		defer lock.Unlock()

		// Ensure that server is already not shutting down.
		select {
		case <-ctx.Done():
			if err == nil {
				newClient.Close()
			}
			return true, nil
		default:
		}

		if err != nil {
			clientErr = err
			return false, nil
		}
		c = newClient
		clientErr = nil
		return true, nil
	}, ctx.Done())

	// Close the client on shutdown.
	go func() {
		<-ctx.Done()

		lock.Lock()
		defer lock.Unlock()
		if c != nil {
			c.Close()
			clientErr = fmt.Errorf("server is shutting down")
		}
	}()

	return func() error {
		// Given that client is closed on shutdown we hold the lock for
		// the entire period of healthcheck call to ensure that client will
		// not be closed during healthcheck.
		// Given that healthchecks has a 2s timeout, worst case of blocking
		// shutdown for additional 2s seems acceptable.
		lock.Lock()
		defer lock.Unlock()
		if clientErr != nil {
			return clientErr
		}

		healthcheckTimeout := time.Duration(healthcheckTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), healthcheckTimeout)
		defer cancel()
		// See https://github.com/etcd-io/etcd/blob/c57f8b3af865d1b531b979889c602ba14377420e/etcdctl/ctlv3/command/ep_command.go#L118
		_, err := c.Get(ctx, "health")
		if err == nil {
			return nil
		}
		return fmt.Errorf("error getting data from etcd: %v", err)
	}, nil
}

func runWorkers(ctx context.Context) {
	for i := 0; i < etcdConnections; i++ {
		go operateEtcdConnection(ctx, fmt.Sprint(i))
	}

	for i := 0; i < etcdWatchers; i++ {
		go operateEtcdWatcher(ctx, fmt.Sprint(i))
	}
}

func operateEtcdConnection(ctx context.Context, connectionID string) {
	c, err := newEtcdClient()
	if err != nil {
		klog.Errorf(err.Error())
	}
	defer c.Close()

	klog.Infof("starting connection %v to generate workload", connectionID)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				klog.Infof("conn: %v, connection state: %v, connection target: %v", connectionID, c.ActiveConnection().GetState().String(), c.ActiveConnection().Target())
				time.Sleep(time.Second)
			}
		}
	}()
	index := 0
	timer := time.NewTimer(time.Second)
	for {
		select {
		case <-ctx.Done():
			klog.Infof("shutting down connection %v", connectionID)
			timer.Stop()
			return
		case <-timer.C:
			_, err := c.Put(ctx, fmt.Sprintf("conn-%v", connectionID), fmt.Sprint(index))
			if err != nil {
				klog.Infof("conn: %v, put %v failed", connectionID, index)
			} else {
				klog.Infof("conn: %v, put %v succeeded", connectionID, index)
			}
			index++
			_ = timer.Reset(time.Second)
		}
	}
}

func operateEtcdWatcher(ctx context.Context, connectionID string) {
	c, err := newEtcdClient()
	if err != nil {
		klog.Errorf(err.Error())
	}
	defer c.Close()
	w := clientv3.NewWatcher(c)
	defer w.Close()
	klog.Infof("starting watcher %v to watch connection %v", connectionID, connectionID)
	wChan := w.Watch(ctx, fmt.Sprintf("conn-%v", connectionID))
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				klog.Infof("watcher: %v, connection state: %v, connection target: %v", connectionID, c.ActiveConnection().GetState().String(), c.ActiveConnection().Target())
				time.Sleep(time.Second)
			}
		}
	}()
	for {
		select {
		case <-ctx.Done():
			klog.Infof("shutting down watcher %v", connectionID)
			return
		case change := <-wChan:
			for _, event := range change.Events {
				klog.Infof("watcher: %v, watched %v changed to %v", connectionID, string(event.Kv.Key), string(event.Kv.Value))
			}
		}
	}
}
