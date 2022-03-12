package containers_window

import (
	docker "dc-top/docker"
	"dc-top/errors"
)

func findIndexOfId(data []docker.ContainerDatum, id string) (int, error) {
	for i, datum := range data {
		if datum.ID() == id {
			return i, nil
		}
	}
	return -1, errors.NewNotFoundError("index of id")
}
