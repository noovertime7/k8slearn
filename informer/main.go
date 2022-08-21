package main

import (
	"flag"
	"fmt"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"time"
)

func main() {
	var err error
	var config *rest.Config
	var kubeconfig *string

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// 使用 ServiceAccount 创建集群配置（InCluster模式）
	if config, err = rest.InClusterConfig(); err != nil {
		// 使用 KubeConfig 文件创建集群配置
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
			panic(err.Error())
		}
	}

	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// 初始化informer factory
	factory := informers.NewSharedInformerFactory(clientset, 30*time.Second)
	//监听想要获取的资源对象informer
	deploymentInformer := factory.Apps().V1().Deployments()
	//注册一下informer
	informer := deploymentInformer.Informer()
	//创建lister
	deployLister := deploymentInformer.Lister()
	//注册事件处理程序 （ADD,UPDATE,DELETE）
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			deploy := obj.(*v1.Deployment)
			fmt.Println("add a deployment", deploy.Name)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			olddeploy := oldObj.(*v1.Deployment)
			newdeploy := newObj.(*v1.Deployment)
			fmt.Println("update a deployment", olddeploy.Name, "==>", newdeploy.Name)
		},
		DeleteFunc: func(obj interface{}) {
			deploy := obj.(*v1.Deployment)
			fmt.Println("delete a deployment", deploy.Name)
		},
	})
	//启动informer (list,watch)
	stopChannl := make(chan struct{})
	factory.Start(stopChannl)
	defer close(stopChannl)
	//等待所有的informer同步完成
	factory.WaitForCacheSync(stopChannl)
	//通过lister获取缓存中的deployment数据
	deployments, err := deployLister.Deployments("default").List(labels.Everything())
	if err != nil {
		panic(err)
	}
	for index, deploy := range deployments {
		fmt.Printf("%d -->%s", index, deploy.Name)
	}
	<-stopChannl
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
