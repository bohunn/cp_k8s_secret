package main

import (
	"context"
	"cp-k8s-secret/cluster"
	"fmt"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1" // Add this import
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"log"
	"os"

	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	NS_FILE = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
)

var (
	namespace       string
	targetNamespace string
	configFile      = cluster.ParseArgs()
	conf            = cluster.ReadAppConfig(*configFile)
)

func init() {
	nsBytes, err := os.ReadFile(NS_FILE)
	if err != nil {
		panic(fmt.Sprintf("Unable to read namespace file at %s", NS_FILE))
	}

	namespace = string(nsBytes)
	targetNamespace = conf["namespace"]
}

func main() {
	// Get things set up for watching - we need a valid k8s client
	clientCfg, err := rest.InClusterConfig()
	if err != nil {
		panic("Unable to get our client configuration")
	}

	// Create a new Kubernetes clientset using the configuration
	clientset, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		panic("Unable to create our clientset")
	}

	var secrets []string = strings.Split(conf["secret_name"], ",")
	fmt.Printf("Got secretlist: %s\n", secrets)

	var wg sync.WaitGroup
	wg.Add(len(secrets))

	for _, s := range secrets {
		go func(s string) {
			defer wg.Done()
			fmt.Printf("Starting Watcher for %s\n", s)
			err := startWatcher(clientset, namespace, s)
			if err != nil {
				log.Fatal(err)
				return
			}
		}(s)
	}
	wg.Wait()
}

func startWatcher(clientset kubernetes.Interface, namespace, name string) error {
	// Create a new ConfigMap watcher for the specified namespace and name
	watcher, err := clientset.CoreV1().Secrets(targetNamespace).Watch(context.Background(), metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", name)})
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Start watching for events
	for {
		select {
		case event, ok := <-watcher.ResultChan():
			if !ok {
				fmt.Printf("Error watching secret %s in namespace %s\n", name, targetNamespace)
				return err
			}
			switch event.Type {
			case watch.Added:
				fmt.Printf("Secret added: %s\n", event.Object.(*corev1.Secret).Name)
				err := createSecret(clientset, namespace, name, event.Object.(*corev1.Secret).Data)
				if err != nil {
					log.Fatal(err)
					return err
				}
			case watch.Modified:
				fmt.Printf("Secret modified: %s\n", event.Object.(*corev1.Secret).Name)
				err := updateSecret(clientset, namespace, name, event.Object.(*corev1.Secret).Data)
				if err != nil {
					log.Fatal(err)
					return err
				}
			case watch.Deleted:
				fmt.Printf("Secret deleted: %s\n", event.Object.(*corev1.Secret).Name)
				err := deleteSecret(clientset, namespace, name)
				if err != nil {
					log.Fatal(err)
					return err
				}
			case watch.Error:
				err := event.Object.(error)
				fmt.Printf("Error watching Secret %s in namespace %s: %s\n", name, targetNamespace, err.Error())
				return err
			}
		}
	}
}

func deleteSecret(clientset kubernetes.Interface, namespace, name string) error {
	err := clientset.CoreV1().Secrets(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil {
		fmt.Errorf("Failed to delete secret %s, in namespace: %s", name, namespace)
	}

	fmt.Println("Deleted secret %s in namespace %s\n", name, namespace)
	return nil
}

func createSecret(clientset kubernetes.Interface, namespace, name string, data map[string][]byte) error {
	_, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("%v", err)
	}
	// when there is already existing secret in target namespace update it
	if err == nil {
		err = updateSecret(clientset, namespace, name, data)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Data: data,
	}

	_, err = clientset.CoreV1().Secrets(namespace).Create(context.Background(), secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create secret: %v", err)
	}
	fmt.Printf("Created secret %s in namespace %s\n", name, namespace)
	return nil
}

func updateSecret(clientset kubernetes.Interface, namespace, name string, data map[string][]byte) error {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get secret: %v", err)
	}

	secret.Data = data

	_, err = clientset.CoreV1().Secrets(namespace).Update(context.Background(), secret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret: %v", err)
	}

	fmt.Printf("Updated secret %s in namespace %s\n", name, namespace)
	return nil
}
