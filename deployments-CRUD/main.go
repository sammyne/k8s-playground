package main

import (
	"context"
	"fmt"

	"github.com/kubernetes/client-go/util/retry"
	flag "github.com/spf13/pflag"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
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

	const DeploymentName = "hello-world"

	spec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: DeploymentName},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
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
							Name:  "hello-world",
							Image: "nginx:1.19.3-alpine",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	client := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	{
		fmt.Println("Creating deployment...")
		d, err := client.Create(context.TODO(), spec, metav1.CreateOptions{})
		if err != nil {
			panic(err)
		}

		fmt.Printf("Created deployment %q\n", d.GetObjectMeta().GetName())
		fmt.Println("Done creating deployment")
	}

	{
		fmt.Println("Read deployments in namespace:", apiv1.NamespaceDefault)
		response, err := client.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		for _, v := range response.Items {
			fmt.Printf(" * %s (%d replicas)\n", v.Name, *v.Spec.Replicas)
		}

		fmt.Println("Done read deployments")
	}

	{
		fmt.Println("Update deployment...")

		//    You have two options to Update() this Deployment:
		//
		//    1. Modify the "deployment" variable and call: Update(deployment).
		//       This works like the "kubectl replace" command and it overwrites/loses changes
		//       made by other clients between you Create() and Update() the object.
		//    2. Modify the "result" returned by Get() and retry Update(result) until
		//       you no longer get a conflict error. This way, you can preserve changes made
		//       by other clients between Create() and Update(). This is implemented below
		//			 using the retry utility package included with client-go. (RECOMMENDED)
		//
		// More Info:
		// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			// Retrieve the latest version of Deployment before attempting update
			// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
			d, err := client.Get(context.TODO(), DeploymentName, metav1.GetOptions{})
			if err != nil {
				panic(fmt.Errorf("Failed to get latest version of Deployment: %v", err))
			}

			d.Spec.Replicas = int32Ptr(1)                                    // reduce replica count
			d.Spec.Template.Spec.Containers[0].Image = "nginx:1.19.2-alpine" // change nginx version
			_, err = client.Update(context.TODO(), d, metav1.UpdateOptions{})
			return err
		})
		if retryErr != nil {
			panic(fmt.Errorf("Update failed: %v", retryErr))
		}

		fmt.Println("Done updating deployment")
	}

	{
		fmt.Println("Read again deployments in namespace:", apiv1.NamespaceDefault)
		response, err := client.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			panic(err)
		}

		for _, v := range response.Items {
			fmt.Printf(" * %s (%d replicas)\n", v.Name, *v.Spec.Replicas)
		}

		fmt.Println("Done reading again deployments")
	}

	{
		fmt.Println("Delete deployment...")
		policy := metav1.DeletePropagationForeground
		opts := metav1.DeleteOptions{PropagationPolicy: &policy}
		if err := client.Delete(context.TODO(), DeploymentName, opts); err != nil {
			panic(err)
		}
		fmt.Println("Done deleting deployment...")
	}
}

func int32Ptr(i int32) *int32 { return &i }
