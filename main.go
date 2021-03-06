package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"managedkube.com/kubernetes-cost-agent/pkg/agent"
	"net/http"
	"os"

	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// optional - local kubeconfig for testing
var kubeconfig = flag.String("kubeconfig", "", "Path to a kubeconfig file")

func main() {

	exportURL, ok := os.LookupEnv("EXPORT_URL")
	if !ok {
		glog.V(3).Infof("The EXPORT_URL environment variable is not set, not sending export data")
	} else {
		glog.V(3).Infof("The EXPORT_URL environment variable is set, sending exports to: %s", exportURL)
		agent.SetExportURL(exportURL)
	}

	exportToken, ok := os.LookupEnv("EXPORT_TOKEN")
	if !ok {
		glog.V(3).Infof("The EXPORT_TOKEN environment variable is not set")
	} else {
		glog.V(3).Infof("The EXPORT_TOKEN environment variable is set, using it as the export token")
		agent.SetExportToken(exportToken)
	}

	clusterName, ok := os.LookupEnv("CLUSTER_NAME")
	if !ok {
		glog.V(3).Infof("The CLUSTER_NAME environment variable is not set")
	} else {
		glog.V(3).Infof("The CLUSTER_NAME environment variable is set to: %s", clusterName)
		agent.SetClusterName(clusterName)
	}

	// send logs to stderr so we can use 'kubectl logs'
	flag.Set("logtostderr", "true")
	flag.Set("v", "3")
	flag.Parse()

	config, err := getConfig(*kubeconfig)
	if err != nil {
		glog.Errorf("Failed to load client config: %v", err)
		return
	}

	// build the Kubernetes clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Failed to create kubernetes client: %v", err)
		return
	}

	go agent.Run(clientset)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":9101", nil)
}

func getConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	return rest.InClusterConfig()
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}
