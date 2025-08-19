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

func parseLogFile(filename string) (map[string]int, error) {
	file, err := os.Open(filename)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	ipCount := make(map[string]int)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {

		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) > 0 {
			ip := parts[0]
			ipCount[ip]++
		}
	}

	return ipCount, scanner.Err()
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

func getTopUserAgents() {
	// TODO: Implement user agent analysis
	return
}

func getTopIPs(ipCount map[string]int, n int) []struct {
	IP    string
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
		IP    string
		Count int
	}, h.Len())

	for i := len(result) - 1; i >= 0; i-- {
		itm := heap.Pop(h).(kv)
		result[i].IP = itm.Key
		result[i].Count = itm.Value
	}
	return result
}

func update_data(data *[][]string, view_type string, topIPs []struct {
	IP    string
	Count int
}, topUserAgents []struct {
	Agent string
	Count int
}) {

	switch view_type {

	case "analyze":
		*data = [][]string{
			{"S.no", "IP Address", "Requests"},
		}
		for i, ip := range topIPs {
			*data = append(*data, []string{fmt.Sprintf("%d", i+1), ip.IP, fmt.Sprintf("%d", ip.Count)})
		}
	case "view":
		*data = [][]string{
			{"S.no", "User Agent", "Requests"},
		}
		for i, agent := range topUserAgents {
			*data = append(*data, []string{fmt.Sprintf("%d", i+1), agent.Agent, fmt.Sprintf("%d", agent.Count)})
		}
	}
}

func LogViewerScreen(path string) fyne.CanvasObject {
	ipCount, err := parseLogFile(path)

	if err != nil {
		return widget.NewLabel("Error: " + err.Error())
	}
	topIPs := getTopIPs(ipCount, 10)
	data := [][]string{
		{"S.no", "IP Address", "Requests"},
	}
	for i, ip := range topIPs {
		data = append(data, []string{fmt.Sprintf("%d", i+1), ip.IP, fmt.Sprintf("%d", ip.Count)})
	}

	var analyze_btn, view_btn *widget.Button
	var topIP, topUserAgent *widget.Button
	var analyze_divisions *fyne.Container

	selected := "analyze"

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
	table.SetColumnWidth(1, 180) // IP Address
	table.SetColumnWidth(2, 100) // Requests

	analyze_btn = widget.NewButton("Analyze", func() {
		selected = "analyze"
		analyze_btn.Importance = widget.HighImportance
		view_btn.Importance = widget.MediumImportance
		analyze_btn.Refresh()
		view_btn.Refresh()
		update_data(&data, selected, topIPs, nil)
		fmt.Println(data)
		table.Refresh()
		analyze_divisions.Show()
	})

	view_btn = widget.NewButton("View", func() {
		selected = "view"
		view_btn.Importance = widget.HighImportance
		analyze_btn.Importance = widget.MediumImportance
		view_btn.Refresh()
		analyze_btn.Refresh()
		update_data(&data, selected, topIPs, nil)
		fmt.Println(data)
		table.Refresh()
		analyze_divisions.Hide()
	})

	analyze_btn.Importance = widget.HighImportance
	view_btn.Importance = widget.MediumImportance

	topIP = widget.NewButton("Top IPs", func() {
	})
	topUserAgent = widget.NewButton("Top User Agents", func() {
	})

	analyze_divisions = container.NewHBox(
		topIP,
		topUserAgent,
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
