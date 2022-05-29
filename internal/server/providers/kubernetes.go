package providers

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/thomasbuchinger/timerec/api"
)

const (
	KubernetesLabelScope        string = "timerec.buc.sh/scope"
	KubernetesLabelPause        string = "timerec.buc.sh/pause"
	KubernetesLabelType         string = "timerec.buc.sh/type"
	KubernetesLabelAppName      string = "app.kubernetes.io/name"
	KubernetesLabelAppManagedBy string = "app.kubernetes.io/managed-by"
	KubernetesAnnotationSchema  string = "timerec.buc.sh/schema"
	KubernetesDataTypeDatastore string = "datastore"
	KubernetesDataAppName       string = "timerec"
	ConfigMapNamePrefix         string = "timerec-"
)

var KubernetesDataPauseValues []string = []string{"true", "yes", "t", "y"}

type KubernetesProvider struct {
	client    *kubernetes.Clientset
	Namespace string
	logger    *zap.SugaredLogger
}

func NewKubernetesProvider(logger zap.SugaredLogger, kubeconfig string) (*KubernetesProvider, error) {
	new := KubernetesProvider{
		logger: logger.Named("KubernetesProvider"),
	}

	var config *rest.Config
	var err error
	// Configure Kubernetes client
	if _, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST"); ok {
		config, err = rest.InClusterConfig()
		logger.Info("Using kubeconfig : InCluster")
	} else if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		logger.Infof("Using kubeconfig : %s/%s\n", kubeconfig, config.Host)
	} else {
		return nil, fmt.Errorf("cannot discover kube config")
	}
	if err != nil {
		return nil, err
	}

	c, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	new.client = c

	// Get the Namespace
	err = new.RefreshNamespace()
	if err != nil {
		return nil, err
	}

	return &new, nil
}

func KubernetesConfigMapFromState(state StateV2) corev1.ConfigMap {
	settingsBytes, _ := yaml.Marshal(state.Users[0].Settings)
	activityBytes, _ := yaml.Marshal(state.Users[0].Activity)
	templatesBytes, _ := yaml.Marshal(state.Templates)
	jobsBytes, _ := yaml.Marshal(state.Jobs)
	recordsBytes, _ := yaml.Marshal(state.Records)

	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: ConfigMapNamePrefix + PartitionToName(state.Users[0].Name),
			Labels: map[string]string{
				KubernetesLabelScope:        state.Users[0].Name,
				KubernetesLabelType:         KubernetesDataTypeDatastore,
				KubernetesLabelAppName:      KubernetesDataAppName,
				KubernetesLabelAppManagedBy: KubernetesDataAppName,
				KubernetesLabelPause:        fmt.Sprint(state.Users[0].Inactive),
			},
			Annotations: map[string]string{
				KubernetesAnnotationSchema: "v1",
			},
			// OwnerReferences: , // At some point a owner reference would probably be a good idea? Maybe?
		},
		Data: map[string]string{
			"Name":      state.Users[0].Name,
			"Settings":  string(settingsBytes),
			"Activity":  string(activityBytes),
			"Templates": string(templatesBytes),
			"Jobs":      string(jobsBytes),
			"Records":   string(recordsBytes),
		},
	}

}
func KubernetesConfigMapToState(state *StateV2, cm corev1.ConfigMap) error {
	var settings api.Settings
	yaml.Unmarshal([]byte(cm.Data["Settings"]), &settings)

	var activity api.Activity
	yaml.Unmarshal([]byte(cm.Data["Activity"]), &activity)

	user := api.User{
		Name:     cm.Labels[KubernetesLabelScope],
		Inactive: false, // Inactive Users are filtered my LabelSelectors
		Activity: activity,
		Settings: settings,
	}
	state.Users = append(state.Users, user)

	var templates []api.RecordTemplate
	yaml.Unmarshal([]byte(cm.Data["Templates"]), &templates)
	state.Templates = append(state.Templates, templates...)

	var jobs []api.Job
	yaml.Unmarshal([]byte(cm.Data["Jobs"]), &jobs)
	state.Jobs = append(state.Jobs, jobs...)

	var records []api.Record
	yaml.Unmarshal([]byte(cm.Data["Records"]), &records)
	state.Records = append(state.Records, records...)

	return nil
}

func (kube *KubernetesProvider) RefreshNamespace() error {
	ns, ok := os.LookupEnv("WATCH_NAMESPACE")
	if ok {
		kube.Namespace = ns
		kube.logger.Debugf("Using WATCH_NAMESPACE: %s\n", ns)
		return nil
	}

	data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	ns = strings.TrimSpace(string(data))
	if err == nil && len(ns) > 0 {
		kube.Namespace = ns
		kube.logger.Debugf("Using ServiceAccount Namespace: %s\n")
		return nil
	}

	kube.logger.Info("No Namespace configured. Using all namespaces")
	kube.Namespace = ""
	return nil
}

func PartitionToName(partition string) string {
	// Replaces all dots with dashes
	return strings.NewReplacer(".", "-").Replace(partition)
}

func (kube *KubernetesProvider) getConfigMap(sel labels.Selector, ns string) ([]corev1.ConfigMap, error) {
	cmList, err := kube.client.CoreV1().ConfigMaps(kube.Namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: sel.String()})

	if err != nil {
		kube.logger.Error(err)
		return []corev1.ConfigMap{}, err
	}

	return cmList.Items, nil
}

func (kube *KubernetesProvider) createOrUpdateConfigMap(cm corev1.ConfigMap, exists bool) error {
	var err error = nil
	ns := kube.Namespace
	if kube.Namespace == "" {
		ns = "default"
	}

	if exists {
		_, err = kube.client.CoreV1().ConfigMaps(ns).Update(context.TODO(), &cm, metav1.UpdateOptions{})
	} else {
		_, err = kube.client.CoreV1().ConfigMaps(ns).Create(context.TODO(), &cm, metav1.CreateOptions{})
	}

	if statusError, isStatus := err.(*errors.StatusError); isStatus {
		kube.logger.Warnf("Error create/update ConfigMap %v\n", statusError.ErrStatus.Message)
		return err
	} else if err != nil {
		kube.logger.Warn(err)
		return err
	}

	return nil
}

func PartitionToSelector(partition string) labels.Selector {
	selector := labels.NewSelector()
	var scopeLabel *labels.Requirement
	if partition == ScopeGlobal {
		scopeLabel, _ = labels.NewRequirement(KubernetesLabelScope, selection.Exists, []string{})
	} else {
		scopeLabel, _ = labels.NewRequirement(KubernetesLabelScope, selection.Equals, []string{partition})
	}
	typeLabel, _ := labels.NewRequirement(KubernetesLabelType, selection.Equals, []string{KubernetesDataTypeDatastore})
	pauseLabel, _ := labels.NewRequirement(KubernetesLabelPause, selection.NotIn, KubernetesDataPauseValues)

	selector = selector.Add(*scopeLabel, *typeLabel, *pauseLabel)
	return selector
}

func (kube *KubernetesProvider) SaveRecord(rec api.Record) (api.Record, error) {
	state, err := kube.Refresh(rec.UserName)
	if err != nil {
		kube.logger.Errorf("Error refreshing Record: %v", err)
		return api.Record{}, err
	}

	state.Records = append(state.Records, rec)
	err = kube.Save(rec.UserName, state)
	if err != nil {
		kube.logger.Errorf("Error saving Record: %v", err)
		return api.Record{}, err
	}
	return rec, nil
}

func (kube *KubernetesProvider) Refresh(partition string) (StateV2, error) {
	selector := PartitionToSelector(partition)
	cms, _ := kube.getConfigMap(selector, kube.Namespace)
	defaultState := StateV2{
		Partition: partition,
		Users:     []api.User{},
		Jobs:      []api.Job{},
		Templates: []api.RecordTemplate{},
		Records:   []api.Record{},
	}

	if len(cms) == 0 && partition != ScopeGlobal {
		defaultState.Users = append(defaultState.Users, api.NewDefaultUser(partition))
		cm := KubernetesConfigMapFromState(defaultState)
		err := kube.createOrUpdateConfigMap(cm, false)
		if err != nil {
			return defaultState, err
		}

		cms = append(cms, cm)
	}

	for _, cm := range cms {
		KubernetesConfigMapToState(&defaultState, cm)
	}

	return defaultState, nil
}

func (kube *KubernetesProvider) Save(partition string, data StateV2) error {
	cm := KubernetesConfigMapFromState(data)

	return kube.createOrUpdateConfigMap(cm, true)
}
