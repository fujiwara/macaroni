package macaroni

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type ECSContainerMetadata struct {
	DockerID      string            `json:"DockerId"`
	Name          string            `json:"Name"`
	DockerName    string            `json:"DockerName"`
	Image         string            `json:"Image"`
	ImageID       string            `json:"ImageID"`
	Labels        map[string]string `json:"Labels"`
	DesiredStatus string            `json:"DesiredStatus"`
	KnownStatus   string            `json:"KnownStatus"`
	Limits        struct {
		CPU    int `json:"CPU"`
		Memory int `json:"Memory"`
	} `json:"Limits"`
	CreatedAt time.Time `json:"CreatedAt"`
	StartedAt time.Time `json:"StartedAt"`
	Type      string    `json:"Type"`
	Networks  []struct {
		NetworkMode   string   `json:"NetworkMode"`
		IPv4Addresses []string `json:"IPv4Addresses"`
	} `json:"Networks"`
}

type ECSTaskMetadata struct {
	Cluster       string `json:"Cluster"`
	TaskARN       string `json:"TaskARN"`
	Family        string `json:"Family"`
	Revision      string `json:"Revision"`
	DesiredStatus string `json:"DesiredStatus"`
	KnownStatus   string `json:"KnownStatus"`
	Containers    []struct {
		DockerID      string            `json:"DockerId"`
		Name          string            `json:"Name"`
		DockerName    string            `json:"DockerName"`
		Image         string            `json:"Image"`
		ImageID       string            `json:"ImageID"`
		Labels        map[string]string `json:"Labels"`
		DesiredStatus string            `json:"DesiredStatus"`
		KnownStatus   string            `json:"KnownStatus"`
		Limits        struct {
			CPU    int `json:"CPU"`
			Memory int `json:"Memory"`
		} `json:"Limits"`
		CreatedAt time.Time `json:"CreatedAt"`
		StartedAt time.Time `json:"StartedAt"`
		Type      string    `json:"Type"`
		Networks  []struct {
			NetworkMode   string   `json:"NetworkMode"`
			IPv4Addresses []string `json:"IPv4Addresses"`
		} `json:"Networks"`
	} `json:"Containers"`
	PullStartedAt time.Time `json:"PullStartedAt"`
	PullStoppedAt time.Time `json:"PullStoppedAt"`
}

type ECSMetadata struct {
	Cluster       string
	TaskARN       string
	ContainerName string
}

func getECSMetadata() (*ECSMetadata, error) {
	// not running in ECS task
	u := getenv("ECS_CONTAINER_METADATA_URI")
	if u == "" {
		return nil, nil
	}

	resp, err := http.Get(u)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ECS container metadata")
	}
	defer resp.Body.Close()
	var meta ECSContainerMetadata
	if err := json.NewDecoder(resp.Body).Decode(&meta); err != nil {
		return nil, errors.Wrap(err, "failed to parse ECS container metadata")
	}

	resp, err = http.Get(u + "/task")
	if err != nil {
		return nil, errors.Wrap(err, "failed to get ECS task metadata")
	}
	defer resp.Body.Close()
	var taskMeta ECSTaskMetadata
	if err := json.NewDecoder(resp.Body).Decode(&taskMeta); err != nil {
		return nil, errors.Wrap(err, "failed to parse ECS task metadata")
	}

	for _, c := range taskMeta.Containers {
		if c.DockerID == meta.DockerID {
			return &ECSMetadata{
				Cluster:       taskMeta.Cluster,
				TaskARN:       taskMeta.TaskARN,
				ContainerName: c.Name,
			}, nil
		}
	}
	return nil, nil
}
