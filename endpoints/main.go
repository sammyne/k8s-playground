package main

import (
	"context"
	"encoding/json"
	"fmt"

	flag "github.com/spf13/pflag"
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
	endpointName string
	masterURL    string
	token        string
	user         string
)

func getKubeConfig() (*ccApi.Config, error) {
	const (
		cluster    = "hello"
		currentCtx = "world"
	)

	clusters := map[string]*ccApi.Cluster{
		cluster: {
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
			Cluster:  cluster,
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
	flag.StringVarP(&endpointName, "endpoint", "e", "", "name of the endpint")
	flag.StringVarP(&masterURL, "master", "m", "",
		"url of master's API server, e.g. https://1.2.3.4:6443, the 'https' but not 'http'")
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

	const namespace = "default"

	endpoints, err := clientset.CoreV1().Endpoints(namespace).Get(
		context.TODO(), endpointName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	//endpointsJSON, _ := json.MarshalIndent(endpoints, "", "  ")
	out, _ := json.MarshalIndent(endpoints.Subsets, "", "  ")
	fmt.Printf("endpoints: %s\n", out)
}
