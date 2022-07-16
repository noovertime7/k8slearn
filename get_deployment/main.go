package main

import (
	"context"
	"flag"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func main() {
	var err error
	var config *rest.Config
	var kubeconfig *string

	//获取kubeconfig
	if home := homedir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "使用 --kubeconfig命令指定kubeconfig")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "")
	}
	flag.Parse()

	// 使用 incluster 模式
	if config, err = rest.InClusterConfig(); err != nil {
		//使用 kubeconfig创建集群配置
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	//使用clientset对象获取资源对象，进行crud
	// 使用 clientsent 获取 Deployments
	deployments, err := clientset.AppsV1().Deployments("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for idx, deploy := range deployments.Items {
		fmt.Printf("%d -> %s\n", idx+1, deploy.Name)
	}
}

func homedir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
