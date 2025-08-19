package logviewer

import (
	"bufio"
	"container/heap"
	"fmt"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func parseLogFile(filename string) (map[string]int, map[string]int, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	ipCount := make(map[string]int)
	userAgentCount := make(map[string]int)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) > 0 {
			ip := parts[0]
			userAgent := parts[len(parts)-1][1 : len(parts[len(parts)-1])-1] // Remove quotes around user agent
			userAgentCount[userAgent]++
			ipCount[ip]++
		}
	}

	return ipCount, userAgentCount, scanner.Err()
}

type kv struct {
	Key   string
	Value int
}

type minHeap []kv

func (h minHeap) Len() int {
	return len(h)
}

func (h minHeap) Less(i, j int) bool {
	return h[i].Value < h[j].Value
}

func (h minHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *minHeap) Push(x interface{}) {
	*h = append(*h, x.(kv))
}

func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func getTopH(ipCount map[string]int, n int) []struct {
	Key   string
	Count int
} {
	h := &minHeap{}
	heap.Init(h)

	for ip, count := range ipCount {
		heap.Push(h, kv{Key: ip, Value: count})
		if h.Len() > n {
			heap.Pop(h)
		}
	}
	result := make([]struct {
		Key   string
		Count int
	}, h.Len())

	for i := len(result) - 1; i >= 0; i-- {
		itm := heap.Pop(h).(kv)
		result[i].Key = itm.Key
		result[i].Count = itm.Value
	}
	return result
}

func update_data(data *[][]string, view_type string, topIPs []struct {
	Key   string
	Count int
}, topUserAgents []struct {
	Key   string
	Count int
}) {

	switch view_type {

	case "ip":
		*data = [][]string{
			{"S.no", "IP Address", "Requests"},
		}
		for i, ip := range topIPs {
			*data = append(*data, []string{fmt.Sprintf("%d", i+1), ip.Key, fmt.Sprintf("%d", ip.Count)})
		}
	case "user_agent":
		*data = [][]string{
			{"S.no", "User Agent", "Requests"},
		}
		for i, agent := range topUserAgents {
			*data = append(*data, []string{fmt.Sprintf("%d", i+1), agent.Key, fmt.Sprintf("%d", agent.Count)})
		}
	}
}

func LogViewerScreen(path string) fyne.CanvasObject {
	ipCount, uaCount, err := parseLogFile(path)

	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}
	topIPs := getTopH(ipCount, 10)
	topUserAgents := getTopH(uaCount, 10)
	data := [][]string{
		{"S.no", "IP Address", "Requests"},
	}
	for i, ip := range topIPs {
		data = append(data, []string{fmt.Sprintf("%d", i+1), ip.Key, fmt.Sprintf("%d", ip.Count)})
	}

	var analyze_btn, view_btn *widget.Button
	var topIP_btn, topUserAgent_btn *widget.Button
	var analyze_divisions *fyne.Container

	selected := "analyze"
	selected_analyze_type := "ip"

	table := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},

		func() fyne.CanvasObject {
			return widget.NewLabel("placeholder")
		},

		func(id widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[id.Row][id.Col])
		},
	)

	// Set column widths for readability
	table.SetColumnWidth(0, 60)  // S.no
	table.SetColumnWidth(1, 400) // IP Address
	table.SetColumnWidth(2, 100) // Requests

	analyze_btn = widget.NewButton("Analyze", func() {
		selected = "analyze"
		analyze_btn.Importance = widget.HighImportance
		view_btn.Importance = widget.MediumImportance
		analyze_btn.Refresh()
		view_btn.Refresh()
		fmt.Println("++++ ", selected, " ++++")
		analyze_divisions.Show()
	})

	view_btn = widget.NewButton("View", func() {
		selected = "view"
		view_btn.Importance = widget.HighImportance
		analyze_btn.Importance = widget.MediumImportance
		view_btn.Refresh()
		analyze_btn.Refresh()
		analyze_divisions.Hide()
	})

	analyze_btn.Importance = widget.HighImportance
	view_btn.Importance = widget.MediumImportance

	topIP_btn = widget.NewButton("Top IPs", func() {
		selected_analyze_type = "ip"
		update_data(&data, selected_analyze_type, topIPs, nil)
		table.Refresh()
		topIP_btn.Importance = widget.HighImportance
		topUserAgent_btn.Importance = widget.MediumImportance
		topIP_btn.Refresh()
		topUserAgent_btn.Refresh()
	})
	topUserAgent_btn = widget.NewButton("Top User Agents", func() {
		selected_analyze_type = "user_agent"
		update_data(&data, selected_analyze_type, nil, topUserAgents)
		table.Refresh()
		topIP_btn.Importance = widget.MediumImportance
		topUserAgent_btn.Importance = widget.HighImportance
		topIP_btn.Refresh()
		topUserAgent_btn.Refresh()
	})

	topIP_btn.Importance = widget.HighImportance
	topUserAgent_btn.Importance = widget.MediumImportance

	analyze_divisions = container.NewHBox(
		topIP_btn,
		topUserAgent_btn,
	)

	topBar := container.NewHBox(
		analyze_btn,
		view_btn,
	)

	upperActionBar := container.NewVBox(
		topBar,
		analyze_divisions,
	)

	scrollTable := container.NewScroll(table)

	content := container.NewBorder(
		upperActionBar,
		nil,         // bottom
		nil,         //left
		nil,         // right
		scrollTable, //center
	)

	return content
}
