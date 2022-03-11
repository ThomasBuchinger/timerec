package providers

import (
	"fmt"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/thomasbuchinger/timerec/api"
	"gopkg.in/yaml.v2"
)

type FileOrMemoryProvider struct {
	Path string
	Data StateV2
}
type FileDiskFormat map[string]StateV2

func NewMemoryProvider() *FileOrMemoryProvider {
	mem := FileOrMemoryProvider{}
	mem.Data = StateV2{
		Partition: "memory",
	}
	return &mem
}

func NewFileProvider(path string) *FileOrMemoryProvider {
	file := &FileOrMemoryProvider{
		Path: path,
		Data: StateV2{
			Partition: "file",
		},
	}
	file.Refresh("file")
	return file
}

func (store *FileOrMemoryProvider) Refresh(partition string) (StateV2, error) {
	if store.Path == "" {
		return store.Data, nil
	}

	var data FileDiskFormat
	content, err := os.ReadFile(store.Path)
	if err != nil {
		fmt.Println(err)
		return StateV2{}, err
	}

	err = yaml.Unmarshal(content, &data)
	if err != nil {
		fmt.Println(err)
		return StateV2{}, err
	}
	return data[partition], nil

}
func (store *FileOrMemoryProvider) Save(partition string, state StateV2) error {
	store.Data = state
	if store.Path == "" {
		// Nothing more to do for Memory Providers
		return nil
	}

	yamlData := FileDiskFormat{}
	yamlData[partition] = state
	content, err := yaml.Marshal(yamlData)
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = os.WriteFile(store.Path, content, os.ModeType)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (store *FileOrMemoryProvider) SaveRecord(rec api.Record) (api.Record, error) {
	data, _ := store.Refresh("file")
	data.Records = append(data.Records, rec)
	store.Save("file", data)
	return rec, nil
}

func (store *FileOrMemoryProvider) NotifyUser(event cloudevents.Event) error {

	fmt.Printf("Event: %s\n", event.String())
	return nil
}
