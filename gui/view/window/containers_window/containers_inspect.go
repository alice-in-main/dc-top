package containers_window

import (
	docker "dc-top/docker"
	"dc-top/gui/elements"
	"fmt"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/gdamore/tcell/v2"
)

func generatePrettyInspectInfo(state tableState, window_width int) (map[int]elements.StringStyler, error) {
	var info map[int]elements.StringStyler = make(map[int]elements.StringStyler)
	var info_arr []elements.StringStyler = make([]elements.StringStyler, 0)

	index, err := findIndexOfId(state.containers_data.GetData(), state.focused_id)
	if err != nil {
		return info, fmt.Errorf("didn't find container '%s' for inspecting", state.focused_id)
	}
	stats := state.containers_data.GetData()[index]
	inspect_info := stats.InspectData()

	if inspect_info.ContainerJSONBase == nil {
		return info, fmt.Errorf("inspection ended")
	}

	basic_info_table := elements.TableWithoutSeperator(window_width, []float64{0.2, 0.8}, [][]elements.StringStyler{
		{elements.TextDrawer("Name:", tcell.StyleDefault), elements.TextDrawer(stats.CachedStats().Name, tcell.StyleDefault)},
		{elements.TextDrawer("ID:", tcell.StyleDefault), elements.TextDrawer(stats.ID(), tcell.StyleDefault)},
		{elements.TextDrawer("Image:", tcell.StyleDefault), elements.TextDrawer(stats.Image(), tcell.StyleDefault)},
		{elements.TextDrawer("State:", tcell.StyleDefault), elements.TextDrawer(stats.State(), tcell.StyleDefault)},
		{elements.TextDrawer("Start date:", tcell.StyleDefault), elements.TextDrawer(inspect_info.Created, tcell.StyleDefault)},
		{elements.TextDrawer("Restart count:", tcell.StyleDefault), elements.IntegerDrawer(inspect_info.RestartCount, tcell.StyleDefault)},
	})
	info_arr = append(info_arr, basic_info_table...)
	cpu_usage := stats.CachedStats().Cpu.ContainerUsage.TotalUsage - stats.CachedStats().PreCpu.ContainerUsage.TotalUsage
	cpu_quota := inspect_info.HostConfig.NanoCPUs
	cpu_limit := stats.CachedStats().Cpu.SystemUsage - stats.CachedStats().PreCpu.SystemUsage
	memory_usage := stats.CachedStats().Memory.Usage
	memory_quota := inspect_info.HostConfig.Memory
	memory_limit := stats.CachedStats().Memory.Limit
	max_desc_len := 25
	bar_len := window_width - max_desc_len
	if bar_len < 0 {
		bar_len = 0
	} else if bar_len > 40 {
		bar_len = 40
	}
	info_arr = append(info_arr,
		generateResourceUsageStyler(cpu_usage, cpu_quota, cpu_limit, "CPU:    ", "Cores", bar_len),
		generateResourceUsageStyler(memory_usage, memory_quota, memory_limit, "Memory: ", "GB", bar_len),
		generateInspectSeperator(),
		elements.TextDrawer("Ports:", tcell.StyleDefault),
	)

	port_map := generatePortMap(inspect_info.NetworkSettings.Ports)
	for _, port_binding := range port_map {
		info_arr = append(info_arr, elements.TextDrawer(port_binding, tcell.StyleDefault))
	}
	info_arr = append(info_arr, generateInspectSeperator(),
		elements.TextDrawer("Mounts:", tcell.StyleDefault),
	)

	parsed_mounts := generateMountsMap(inspect_info.Mounts)
	for _, mount := range parsed_mounts {
		info_arr = append(info_arr, elements.TextDrawer(mount, tcell.StyleDefault))
	}

	info_arr = append(info_arr, generateInspectSeperator(),
		elements.TextDrawer("Network Usage:", tcell.StyleDefault),
	)
	network_usage, err := generateNetworkUsage(stats.CachedStats().Network, stats.CachedStats().PreNetwork)
	if err != nil {
		return info, err
	}
	info_arr = append(info_arr, network_usage...)

	info_arr = append(info_arr, generateInspectSeperator())

	var row_offset int
	num_rows := len(info_arr)
	if num_rows > state.inspect_height {
		row_offset += state.top_line_inspect % (1 + num_rows - state.inspect_height)
		if row_offset < 0 {
			row_offset += 1 + num_rows - state.inspect_height
		}
	}
	for i, line := range info_arr {
		info[i-row_offset] = line
	}
	return info, nil
}

func generateResourceUsageStyler(usage, quota, limit int64, resource, unit string, bar_len int) elements.StringStyler {
	var quota_desc string
	if quota == 0 {
		quota_desc = " Quota isn't set"
		quota = limit
	} else {
		quota_desc = fmt.Sprintf(" Quota: %.2f%s", float64(quota)/float64(1<<30), unit)
	}
	usage_human_readable := resourceFormatter(usage, quota, unit)
	return elements.ValuesBarDrawer(
		resource,
		0.0,
		float64(quota),
		float64(usage),
		bar_len,
		[]rune(" "+usage_human_readable+quota_desc))
}

func generatePortMap(ports nat.PortMap) []string {
	var port_map []string = make([]string, len(ports))
	index := 0
	for port, port_bindings := range ports {
		if len(port_bindings) > 0 {
			for _, binding := range port_bindings {
				port_map[index] = fmt.Sprintf("  %s : %s", port, binding.HostPort)
			}
		} else {
			port_map[index] = fmt.Sprintf("  %s not mapped", port)
		}
		index++
	}
	sort.Strings(sort.StringSlice(port_map))
	return port_map
}

func generateMountsMap(mounts []types.MountPoint) []string {
	sort.SliceStable(mounts, func(i, j int) bool { return mounts[i].Destination < mounts[j].Destination })
	var parsed_mounts []string = make([]string, 3*len(mounts))
	for i := 0; i < 2*len(mounts); i += 3 {
		mount_num := i / 3
		parsed_mounts[i] = fmt.Sprintf("  %s> %s", mounts[mount_num].Type, mounts[mount_num].Name)
		parsed_mounts[i+1] = fmt.Sprintf("    %s:%s", mounts[mount_num].Source, mounts[mount_num].Destination)
		parsed_mounts[i+2] = fmt.Sprintf("    Mode: %s, Driver: %s, RW: %t", mounts[mount_num].Mode, mounts[mount_num].Driver, mounts[mount_num].RW)
	}
	return parsed_mounts
}

func generateNetworkUsage(curr_stats map[string]docker.NetworkUsage, prev_stats map[string]docker.NetworkUsage) ([]elements.StringStyler, error) {
	ret := make([]elements.StringStyler, 0)
	for network_interface, usage := range curr_stats {
		ret = append(ret, elements.TextDrawer("  "+network_interface, tcell.StyleDefault))
		mapped_usage, err := docker.NetworkUsageToMapOfInt(usage)
		if err != nil {
			return nil, err
		}
		prev_mapped_usage, err := docker.NetworkUsageToMapOfInt(prev_stats[network_interface])
		if err != nil {
			return nil, err
		}
		time_diff := curr_stats[network_interface].LastUpdateTime.Sub(prev_stats[network_interface].LastUpdateTime).Seconds()

		sorted_usage_keys := make([]string, 0, len(mapped_usage))
		for k := range mapped_usage {
			sorted_usage_keys = append(sorted_usage_keys, k)
		}
		sort.Strings(sorted_usage_keys)

		for _, key := range sorted_usage_keys {
			max_line_len := 30
			curr_usage := mapped_usage[key]
			prev_usage := prev_mapped_usage[key]
			line := fmt.Sprintf("    %s:%d", key, curr_usage)
			padding_len := max_line_len - len(line)
			line += strings.Repeat(" ", padding_len)
			if strings.Contains(key, "byte") {
				speed := float64(curr_usage-prev_usage) / time_diff
				var unit string
				switch {
				case speed > (1 << 30):
					speed /= (1 << 30)
					unit = "GB/s"
				case speed > (1 << 20):
					speed /= (1 << 20)
					unit = "MB/s"
				case speed > (1 << 10):
					speed /= (1 << 10)
					unit = "KB/s"
				default:
					unit = "Bytes/s"
				}
				line += fmt.Sprintf("%.3f%s", speed, unit)
			} else {
				line += fmt.Sprintf("%.3f/s", float64(curr_usage-prev_usage)/time_diff)
			}
			ret = append(ret, elements.TextDrawer(line, tcell.StyleDefault))
		}

	}
	return ret, nil
}

func generateInspectSeperator() elements.StringStyler {
	const underline_rune = '\u2500'
	return elements.RuneRepeater(underline_rune, tcell.StyleDefault.Foreground(tcell.ColorPaleGreen))
}
