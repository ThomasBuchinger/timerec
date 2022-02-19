package providers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/thomasbuchinger/timerec/api"
)

type KubernetesProvider struct {
	client          *kubernetes.Clientset
	Namespace       string
	ConfigMapPrefix string
	ConfigMapNames  []string
	logger          *zap.SugaredLogger
}

type KubernetesProviderConfigMap struct {
	Name      string
	Inactive  bool
	Activity  api.Activity
	Settings  api.Settings
	Templates []api.RecordTemplate
	Jobs      []api.Job
	Records   []api.Record
}

func NewKubernetesProvider(logger zap.SugaredLogger) (*KubernetesProvider, error) {
	new := KubernetesProvider{
		logger:          logger.Named("KubernetesProvider"),
		ConfigMapPrefix: "timerec-state-",
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	new.client = c
	new.RefreshNamespace()
	new.RefreshConfigMaps()
	return &new, nil
}

func (kube *KubernetesProvider) RefreshConfigMaps() {
	kube.ConfigMapNames = []string{"me", "buc"}
}

func (kube *KubernetesProvider) RefreshNamespace() {
	ns, ok := os.LookupEnv("POD_NAMESPACE")
	if ok {
		kube.Namespace = ns
		return
	}

	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns = strings.TrimSpace(string(data))
	if err == nil && len(ns) > 0 {
		kube.Namespace = ns
		return
	}
	kube.Namespace = ""
}

func (kube *KubernetesProvider) configMapIsKnown(name string) bool {
	for _, cm := range kube.ConfigMapNames {
		if name == cm {
			return true
		}
	}
	return false
}

func (kube *KubernetesProvider) getConfigMapForUser(configMapName string) error {
	if !kube.configMapIsKnown(configMapName) {
		kube.RefreshConfigMaps()
		if !kube.configMapIsKnown(configMapName) {
			return fmt.Errorf("database found")
		}
	}

	cm, err := kube.client.CoreV1().ConfigMaps("").Get(context.TODO(), kube.ConfigMapPrefix+configMapName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		fmt.Printf("Pod example-xxxxx not found in default namespace\n")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		fmt.Printf("Error getting pod %v\n", statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	}

	fmt.Println(
		cm.Data["activity"],
		cm.Data["jobs"],
		cm.Data["templates"],
	)

	return nil
}

func (kube *KubernetesProvider) ListUsers() ([]api.User, error) {
	return []api.User{}, nil
}

func (kube *KubernetesProvider) GetUser(u api.User) (api.User, error) {
	return api.User{}, nil
}

func (kube *KubernetesProvider) CreateUser(new api.User) (api.User, error) {
	return api.User{}, nil
}

func (kube *KubernetesProvider) UpdateUser(new api.User) (api.User, error) {
	return api.User{}, nil
}

func (kube *KubernetesProvider) GetTemplates() ([]api.RecordTemplate, error) {
	return []api.RecordTemplate{}, nil
}

func (kube *KubernetesProvider) HasTemplate(name string) (bool, error) {
	return false, nil
}

func (kube *KubernetesProvider) GetTemplate(name string) (api.RecordTemplate, error) {
	return api.RecordTemplate{}, nil
}

func (kube *KubernetesProvider) CreateJob(t api.Job) (api.Job, error) {
	return api.Job{}, nil
}

func (kube *KubernetesProvider) ListJobs() ([]api.Job, error) {
	return []api.Job{}, nil
}

func (kube *KubernetesProvider) GetJob(t api.Job) (api.Job, error) {
	return api.Job{}, nil
}

func (kube *KubernetesProvider) UpdateJob(t api.Job) (api.Job, error) {
	return api.Job{}, nil
}

func (kube *KubernetesProvider) DeleteJob(t api.Job) (api.Job, error) {
	return api.Job{}, nil
}
