package main

import (
	"context"
	"fmt"
	"time"

	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ccApi "k8s.io/client-go/tools/clientcmd/api"
	//
	// Uncomment to load all auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth"
	//
	// Or uncomment to load specific auth plugins
	// _ "k8s.io/client-go/plugin/pkg/client/auth/azure"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	// _ "k8s.io/client-go/plugin/pkg/client/auth/openstack"
)

var (
	clusterName string
	currentCtx  string
	masterURL   string
	token       string
	user        string
)

func getKubeConfig() (*ccApi.Config, error) {
	clusters := map[string]*ccApi.Cluster{
		clusterName: {
			Server:                masterURL,
			InsecureSkipTLSVerify: true,
		},
	}

	authInfos := map[string]*ccApi.AuthInfo{
		//user: {Token: "abaf6bc8-f80f-4767-86e0-cbf57b9cb1c7"},
		user: {Token: token},
	}

	contexts := map[string]*ccApi.Context{
		currentCtx: {
			Cluster:  clusterName,
			AuthInfo: user,
		},
	}

	out := &ccApi.Config{
		Clusters:       clusters,
		AuthInfos:      authInfos,
		Contexts:       contexts,
		CurrentContext: currentCtx,
	}

	return out, nil
}

func init() {
	flag.StringVar(&clusterName, "cluster", "sgx1", "name of cluster")
	flag.StringVar(&currentCtx, "ctx", "sgx1", "context to use")
	flag.StringVarP(&masterURL, "master", "m", "",
		"url of master's API server, e.g. http://1.2.3.4:6443")
	flag.StringVarP(&token, "token", "t", "", "NON-EMPTY token for authenticatoin")
	flag.StringVarP(&user, "user", "u", "xml", "user name")
}

func main() {
	flag.Parse()
	if token == "" || masterURL == "" {
		fmt.Println("token/masterURL mustn't be empty")
		flag.PrintDefaults()
		return
	}

	config, err := clientcmd.BuildConfigFromKubeconfigGetter(masterURL, getKubeConfig)
	if err != nil {
		panic(err)
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))

	// Examples for error handling:
	// - Use helper functions like e.g. errors.IsNotFound()
	// - And/or cast to StatusError and use its properties like e.g. ErrStatus.Message
	namespace := "default"
	pod := "example-xxxxx"
	_, err = clientset.CoreV1().Pods(namespace).Get(context.TODO(), pod, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod %s in namespace %s not found\n", pod, namespace)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %s in namespace %s: %v\n",
			pod, namespace, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		fmt.Printf("Found pod %s in namespace %s\n", pod, namespace)
	}

	time.Sleep(10 * time.Second)
}
