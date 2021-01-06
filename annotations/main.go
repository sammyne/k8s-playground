package main

import (
	"encoding/json"
	"fmt"

	flag "github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	ccApi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	masterURL string
	token     string
	user      string
)

const clusterName = "sammyne"

func getKubeConfig() (*ccApi.Config, error) {
	clusters := map[string]*ccApi.Cluster{
		clusterName: {
			Server:                masterURL,
			InsecureSkipTLSVerify: true,
		},
	}

	authInfos := map[string]*ccApi.AuthInfo{
		user: {Token: token},
	}

	currentCtx := clusterName
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
	flag.StringVarP(&masterURL, "master", "m", "",
		"url of master's API server, e.g. https://1.2.3.4:6443")
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

	const DeploymentName = "hello-world"

	annotations := map[string]string{"how-do-you-do": "I'm fine"}

	spec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Name:        DeploymentName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "hello-world"},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "hello-world"},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:    "hello-world",
							Image:   "busybox:1.33.0",
							Command: []string{"tail", "-f", "/dev/null"},
						},
					},
				},
			},
		},
	}

	client := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	{
		fmt.Println("Creating deployment...")
		d, err := client.Create(spec)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Created deployment %q\n", d.GetObjectMeta().GetName())
		fmt.Println("Done creating deployment")
	}

	{
		fmt.Println("Read deployments in namespace:", apiv1.NamespaceDefault)
		response, err := client.Get(DeploymentName, metav1.GetOptions{})
		if err != nil {
			panic(err)
		}

		metaJSON, _ := json.MarshalIndent(response.ObjectMeta.Annotations, "", "  ")
		fmt.Println("meta")
		fmt.Printf("%s\n", metaJSON)

		//statusJSON, _ := json.MarshalIndent(response.Status, "", "  ")
		//fmt.Printf("status: %s\n", statusJSON)

		fmt.Println("Done read deployments")
	}

	{
		fmt.Println("Delete deployment...")
		policy := metav1.DeletePropagationForeground
		opts := metav1.DeleteOptions{PropagationPolicy: &policy}
		if err := client.Delete(DeploymentName, &opts); err != nil {
			panic(err)
		}
		fmt.Println("Done deleting deployment...")
	}
}

func int32Ptr(i int32) *int32 { return &i }
