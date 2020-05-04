package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type PortForwardAPodRequest struct {
	// RestConfig is the kubernetes config
	RestConfig *rest.Config
	// Pod is the selected pod for this port forwarding
	Pod v1.Pod
	// LocalPort is the local port that will be selected to expose the PodPort
	LocalPort int
	// PodPort is the target port for the pod
	PodPort int
	// Steams configures where to write or read input from
	Streams genericclioptions.IOStreams
	// StopCh is the channel used to manage the port forward lifecycle
	StopCh <-chan struct{}
	// ReadyCh communicates when the tunnel is ready to receive traffic
	ReadyCh chan struct{}
}

func portForwarding(forward Forward) error {

	var wg sync.WaitGroup

	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	for _, s := range forward.Services {
		wg.Add(1)
		go portForwardService(s.Namespace, s.Name, s.Port, s.TargetPort, &wg, config, clientset)
	}

	log.Println("Ready to get traffic!")
	log.Println("Press [Ctrl-C] to stop forwarding.")

	wg.Wait()
	return nil
}

func portForwardService(namespace string, service string, localPort int, podPort int, wg *sync.WaitGroup, config *rest.Config, clientset *kubernetes.Clientset) {

	// stopCh control the port forwarding lifecycle. When it gets closed the
	// port forward will terminate
	stopCh := make(chan struct{}, 1)
	// readyCh communicate when the port forward is ready to get traffic
	readyCh := make(chan struct{})
	// stream is used to tell the port forwarder where to place its output or
	// where to expect input if needed. For the port forwarding we just need
	// the output eventually
	stream := genericclioptions.IOStreams{
		In:     os.Stdin,
		Out:    os.Stdout,
		ErrOut: os.Stderr,
	}

	// managing termination signal from the terminal. As you can see the stopCh
	// gets closed to gracefully handle its termination.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Printf("Stop forwarding %s:%d", service, localPort)
		close(stopCh)
		wg.Done()
	}()

	wait := 10 * time.Second
	for {
		forwardReady := false
		//get first pod name to set up forwarding
		podName, err := getServerPod(clientset, namespace, service)

		if err != nil {
			log.Println(err.Error())
		} else {

			go func() {
				log.Printf("Start forwarding %s", service)
				err := PortForwardAPod(PortForwardAPodRequest{
					RestConfig: config,
					Pod: v1.Pod{
						ObjectMeta: metav1.ObjectMeta{
							Name:      podName,
							Namespace: namespace,
						},
					},
					LocalPort: localPort,
					PodPort:   podPort,
					Streams:   stream,
					StopCh:    stopCh,
					ReadyCh:   readyCh,
				})
				if err != nil {
					log.Println(err.Error())
				}
			}()

			select {
			case <-readyCh:
				forwardReady = true
				break
			}

			if forwardReady {
				break
			}
		}

		log.Printf("Forward %s retry in %d seconds", service, wait/time.Second)
		time.Sleep(wait)

	}

}

func PortForwardAPod(req PortForwardAPodRequest) error {
	path := fmt.Sprintf("/api/v1/namespaces/%s/pods/%s/portforward",
		req.Pod.Namespace, req.Pod.Name)
	hostIP := strings.TrimLeft(req.RestConfig.Host, "htps:/")

	transport, upgrader, err := spdy.RoundTripperFor(req.RestConfig)
	if err != nil {
		return err
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, &url.URL{Scheme: "https", Path: path, Host: hostIP})
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", req.LocalPort, req.PodPort)}, req.StopCh, req.ReadyCh, req.Streams.Out, req.Streams.ErrOut)
	if err != nil {
		return err
	}
	return fw.ForwardPorts()
}

func getServerPod(clientset *kubernetes.Clientset, namespace string, serviceName string) (string, error) {

	serviceClient := clientset.CoreV1().Services(namespace)

	svc, err := serviceClient.Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	set := labels.Set(svc.Spec.Selector)
	selector := set.AsSelector().String()

	if selector == "" {
		return "", errors.New(fmt.Sprintf("WARNING: No Pod selector for service %s in %s on cluster %s.\n", svc.Name, svc.Namespace, svc.ClusterName))
	}

	listOpts := metav1.ListOptions{LabelSelector: selector}

	pods, err := clientset.CoreV1().Pods(svc.Namespace).List(listOpts)

	if err != nil {
		return "", err
	}

	if len(pods.Items) < 1 {
		return "", fmt.Errorf("WARNING: No Running Pods returned for service %s in %s on cluster %s.\n", svc.Name, svc.Namespace, svc.ClusterName)
	}

	return pods.Items[0].Name, err
}

func homeDir() string {
	switch osName := runtime.GOOS; osName {
	case "darwin":
		return os.Getenv("HOME")
	case "linux":
		return os.Getenv("HOME")
	default:
		// freebsd, openbsd,
		// plan9, windows...
		return os.Getenv("USERPROFILE")
	}
}
