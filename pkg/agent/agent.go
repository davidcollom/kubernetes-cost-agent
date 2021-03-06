package agent

import (
	"bytes"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"time"

	"k8s.io/client-go/kubernetes"

	k8sNode "managedkube.com/kubernetes-cost-agent/pkg/metrics/k8s/node"
	k8sPersistentVolume "managedkube.com/kubernetes-cost-agent/pkg/metrics/k8s/persistentVolume"
	k8sPod "managedkube.com/kubernetes-cost-agent/pkg/metrics/k8s/pod"
)

var AgentVersion = "1.1"
var exportCycleSeconds time.Duration = 10
var exportURL = ""
var exportToken = ""
var clusterName = ""

var (
	AppVersionPrometheus = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mk_agent_build_info",
		Help: "ManagedKube - Build information",
	},
		[]string{"version"},
	)
)

func SetExportURL(url string) {
	exportURL = url
}

func SetExportToken(token string) {
	exportToken = token
}

func SetClusterName(name string) {
	clusterName = name
}

// Registers the Prometheus metrics
func register() {
	// Metrics have to be registered to be exposed:
	prometheus.MustRegister(AppVersionPrometheus)
}

func setVersion() {

	AppVersionPrometheus.With(prometheus.Labels{"version": AgentVersion}).Set(1)
}

func Run(clientset *kubernetes.Clientset) {

	glog.V(3).Infof("ManagedKube Agent version %s", AgentVersion)

	register()
	setVersion()
	k8sNode.Register()
	k8sPod.Register()
	k8sPersistentVolume.Register()
	//k8sNamespace.Register()

	go k8sNode.Watch(clientset)
	time.Sleep(5 * time.Second)
	go k8sPod.Watch(clientset)
	go k8sPersistentVolume.Watch(clientset)

	if exportURL != "" {
		go export()
	}
}

func export() {
	update()
}

func update() {
	for {
		time.Sleep(exportCycleSeconds * time.Second)
		glog.V(3).Infof("Sending exports")

		sendPods()
		sendNodes()
		sendPersistentDisk()
	}
}

func send(urlPath string, bytesRepresentation []uint8) {

	timeout := time.Duration(30 * time.Second)

	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", exportURL+urlPath, bytes.NewBuffer(bytesRepresentation))
	req.Header.Set("Apikey", exportToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()

	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)

	log.Println(result)
	log.Println(result["data"])

	if resp.StatusCode != 200 {
		glog.V(3).Infof("Error sending export to: %s, StatusCode: %s", exportURL, resp.Status)
	}
}

func sendPods() {
	podList := k8sPod.GetList()

	for _, p := range podList.Pod {

		data := PodExport{
			ApiVersion: "managedkube/v1alpha1",
			Kind:       "PodMetric",
			Metadata: Metadata{
				Name:      clusterName,
				Namespace: p.Namespace_name,
				Labels: Labels{
					ClusterName: clusterName,
				},
			},
			Spec: p,
		}

		bytesRepresentation, err := json.Marshal(data)
		if err != nil {
			log.Fatalln(err)
		}

		go send("/exports/pods", bytesRepresentation)
		//go send("", bytesRepresentation)
	}
}

func sendNodes() {
	nodeList := k8sNode.GetList()

	for _, n := range nodeList.Node {

		data := NodeExport{
			ApiVersion: "managedkube/v1alpha1",
			Kind:       "NodeMetric",
			Metadata: Metadata{
				Name:      clusterName,
				Namespace: "",
				Labels: Labels{
					ClusterName: clusterName,
				},
			},
			Spec: n,
		}

		bytesRepresentation, err := json.Marshal(data)
		if err != nil {
			log.Fatalln(err)
		}

		go send("/exports/nodes", bytesRepresentation)
	}
}

func sendPersistentDisk() {
	pvList := k8sPersistentVolume.GetList()

	for _, n := range pvList.PersistentVolume {

		data := PersistentDiskExport{
			ApiVersion: "managedkube/v1alpha1",
			Kind:       "PersistentVolumeeMetric",
			Metadata: Metadata{
				Name:      clusterName,
				Namespace: n.Claim.Namespace,
				Labels: Labels{
					ClusterName: clusterName,
				},
			},
			Spec: n,
		}

		bytesRepresentation, err := json.Marshal(data)
		if err != nil {
			log.Fatalln(err)
		}

		go send("/exports/persistentvolumes", bytesRepresentation)
	}
}
