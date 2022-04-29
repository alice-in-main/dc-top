package compose

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/docker/docker/api/types/filters"
)

var (
	_curr_filters filters.Args
	_filters_lock sync.Mutex
)

func GetContainerFilters(ctx context.Context) filters.Args {
	_filters_lock.Lock()
	defer _filters_lock.Unlock()

	return _curr_filters.Clone()
}

func UpdateContainerFilters(ctx context.Context) error {
	_filters_lock.Lock()
	defer _filters_lock.Unlock()

	var contaiener_filters filters.Args = filters.NewArgs()
	if DcModeEnabled() {
		dc_filters, err := GetDcProcesses(ctx)
		if err != nil {
			log.Printf("Failed to update compose filters: '%s'", err.Error())
			return err
		}

		for _, filter := range dc_filters {
			contaiener_filters.Add("name", fmt.Sprintf("^/%s$", filter.Name))
		}
		_curr_filters = contaiener_filters
	}

	return nil
}
