package main

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

//todo handler magic variables
var (
	kubeconfig = "/home/user/.kube/config"
	pod        = "nginx-deployment-6ddf4c5bf7-hpw78"
	namespace  = "default"
	container  = "nginx"
	command    = "ls"
)

func main() {
	fmt.Println("Hello from cluster!")

	config := createInnerClusterConfig()
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	for {
		pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

		time.Sleep(10 * time.Second)
	}
}

// todo create config provider

func createInnerClusterConfig() *rest.Config {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic("Unable to get configuration" + err.Error())
	}
	return config
}

func createOuterClusterConfig() *rest.Config {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic("Unable to get configuration" + err.Error())
	}
	return config
}
