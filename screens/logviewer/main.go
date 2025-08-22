package logviewer

import (
	"alogviewer/widgets/clickable"
	"bufio"
	"container/heap"
	"fmt"
	"math"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// Cache structure for parsed data
type ParsedLogData struct {
	IPCount        map[string]int
	UserAgentCount map[string]int
	TotalLines     int
	LastModified   time.Time
	Filename       string
}

var (
	logCache      = make(map[string]*ParsedLogData)
	cacheMutex    sync.RWMutex
	compiledRegex *regexp.Regexp
)

func init() {
	// Pre-compile regex for better performance
	compiledRegex = regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]+) (\S+)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)"$`)
}

func parseLogFile(filename string) (map[string]int, map[string]int, int, error) {
	// Check cache first
	cacheMutex.RLock()
	if cached, exists := logCache[filename]; exists {
		// Check if file was modified
		fileInfo, err := os.Stat(filename)
		if err == nil && !fileInfo.ModTime().After(cached.LastModified) {
			cacheMutex.RUnlock()
			return cached.IPCount, cached.UserAgentCount, cached.TotalLines, nil
		}
	}
	cacheMutex.RUnlock()

	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, 0, err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()

	ipCount := make(map[string]int, 1000)       // Pre-allocate with estimated capacity
	userAgentCount := make(map[string]int, 500) // Pre-allocate with estimated capacity

	scanner := bufio.NewScanner(file)
	// Use larger buffer for better performance with large log files
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue // Skip empty lines
		}

		// Use faster field splitting
		if firstSpace := strings.IndexByte(line, ' '); firstSpace > 0 {
			ip := line[:firstSpace]
			ipCount[ip]++

			// Extract user agent more efficiently
			if lastQuote := strings.LastIndexByte(line, '"'); lastQuote > 0 {
				if secondLastQuote := strings.LastIndexByte(line[:lastQuote], '"'); secondLastQuote >= 0 {
					userAgent := line[secondLastQuote+1 : lastQuote]
					userAgentCount[userAgent]++
				}
			}
		}
		lineCount++
	}

	// Cache the results
	cached := &ParsedLogData{
		IPCount:        ipCount,
		UserAgentCount: userAgentCount,
		TotalLines:     lineCount,
		LastModified:   fileInfo.ModTime(),
		Filename:       filename,
	}

	cacheMutex.Lock()
	logCache[filename] = cached
	cacheMutex.Unlock()

	return ipCount, userAgentCount, lineCount, scanner.Err()
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
}) int {
	ua_char := 0
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
			if len(agent.Key) > ua_char {
				ua_char = len(agent.Key)
			}
			*data = append(*data, []string{fmt.Sprintf("%d", i+1), agent.Key, fmt.Sprintf("%d", agent.Count)})
		}
	}
	return ua_char
}

func parseLogLine(line string) (ip, datetime, method, path, protocol, status, userAgent string) {
	// Regex for combined log format
	re := regexp.MustCompile(`^(\S+) \S+ \S+ \[([^\]]+)\] "(\S+) ([^"]+) (\S+)" (\d{3}) (\d+|-) "([^"]*)" "([^"]*)"$`)

	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return "", "", "", "", "", "", ""
	}

	return matches[1], // IP
		matches[2], // Date/Time
		matches[3], // Method
		matches[4], // Path
		matches[5], // Protocol
		matches[6], // Status
		matches[9] // User Agent
}

func setLogData(data *[][]string, cp int, n int, filename string) (ip_c, time_c, url_c, status_c, userAgent_c int) {
	*data = [][]string{
		{"S.no", "IP Address", "Time", "Method", "Path", "Protocol", "Status", "User Agent"},
	}
	file, err := os.Open(filename)
	if err != nil {
		return 0, 0, 0, 0, 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	start := (cp - 1) * n
	end := start + n

	i := 1

	var ip_char, time_char, url_char, status_char, userAgent_char int

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) > 0 && i >= start+1 && i <= end {
			ip, time, method, path, protocol, status, userAgent := parseLogLine(line)
			*data = append(*data, []string{fmt.Sprintf("%d", i), ip, time, method, path, protocol, status, userAgent})
			if len(ip) > ip_char {
				ip_char = len(ip)
			}
			if len(time) > time_char {
				time_char = len(time)
			}
			if len(path) > url_char {
				url_char = len(path)
			}
			if len(status) > status_char {
				status_char = len(status)
			}
			if len(userAgent) > userAgent_char {
				userAgent_char = len(userAgent)
			}

		} else if i >= end {
			break
		}
		i++

	}
	return ip_char, time_char, url_char, status_char, userAgent_char

}

func copytxt(data [][]string, row, col int) string {

	if row < 0 || row >= len(data) || col < 0 || col >= len(data[row]) {
		return ""
	}

	if col == 0 {
		return strings.Join(data[row], " ")
	}

	return data[row][col]
}

func LogViewerScreen(path string, app fyne.App) fyne.CanvasObject {
	ipCount, uaCount, total_req, err := parseLogFile(path)
	cPage := 1
	perPage := 10
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
	var n_btn, b_btn *widget.Button
	var pagination_buttons *fyne.Container
	selected := "analyze"
	selected_analyze_type := "ip"

	table := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			clickableObj := clickable.NewClickable(label, nil, nil)
			return clickableObj
		},
		func(id widget.TableCellID, o fyne.CanvasObject) {

			clickableObj := o.(*clickable.Clickable)
			label := clickableObj.Content.(*widget.Label)
			label.SetText(data[id.Row][id.Col])

			clickableObj.OnClick = func() {
				if id.Row != 0 {
					app.Clipboard().SetContent(copytxt(data, id.Row, id.Col))
				}
			}
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
		update_data(&data, selected_analyze_type, topIPs, topUserAgents)
		table.Refresh()
		analyze_btn.Refresh()
		view_btn.Refresh()
		analyze_divisions.Show()
		pagination_buttons.Hide()
		debug.FreeOSMemory()
	})

	updateTable := func() {
		ip_c, time_c, url_c, _, userAgent_c := setLogData(&data, cPage, perPage, path)
		nScaleFactor := float32(8.0)
		table.SetColumnWidth(0, 40)                                // Sno
		table.SetColumnWidth(1, float32(ip_c)*nScaleFactor)        // ip
		table.SetColumnWidth(2, float32(time_c)*nScaleFactor)      // t
		table.SetColumnWidth(3, 75)                                // method
		table.SetColumnWidth(4, float32(url_c)*nScaleFactor)       // url
		table.SetColumnWidth(5, 75)                                // protocol
		table.SetColumnWidth(6, 75)                                // stat
		table.SetColumnWidth(7, float32(userAgent_c)*nScaleFactor) // ua
		table.Refresh()
		view_btn.Importance = widget.HighImportance
		analyze_btn.Importance = widget.MediumImportance
		view_btn.Refresh()
		analyze_btn.Refresh()
		analyze_divisions.Hide()
		b_btn.SetText(fmt.Sprintf("< %d", cPage-1))
		n_btn.SetText(fmt.Sprintf("%d >", cPage+1))
		if cPage == 1 {
			b_btn.Hide()
		} else {
			b_btn.Show()
		}
		debug.FreeOSMemory()
	}

	view_btn = widget.NewButton("View", func() {
		selected = "view"
		updateTable()
		pagination_buttons.Show()
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
		debug.FreeOSMemory()
	})

	topUserAgent_btn = widget.NewButton("Top User Agents", func() {
		selected_analyze_type = "user_agent"
		ua_c := update_data(&data, selected_analyze_type, nil, topUserAgents)
		table.SetColumnWidth(0, 60)              // S.no
		table.SetColumnWidth(1, float32(ua_c)*8) // IP Address
		table.SetColumnWidth(2, 100)             // Requests
		table.Refresh()
		topIP_btn.Importance = widget.MediumImportance
		topUserAgent_btn.Importance = widget.HighImportance
		topIP_btn.Refresh()
		topUserAgent_btn.Refresh()
		debug.FreeOSMemory()
	})

	topIP_btn.Importance = widget.HighImportance
	topUserAgent_btn.Importance = widget.MediumImportance

	n_btn = widget.NewButton("2 >", func() {

		if selected == "view" && cPage+1 <= int(math.Ceil(float64(total_req)/float64(perPage))) {
			cPage++
			updateTable()
			debug.FreeOSMemory()
		}

	})

	b_btn = widget.NewButton("< 1", func() {
		if cPage <= 1 {
			cPage = 1
		} else {
			cPage--
			updateTable()
			debug.FreeOSMemory()
		}
	})

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
	pagination_buttons = container.NewBorder(
		nil,
		nil,
		b_btn,
		n_btn,
		nil,
	)
	pagination_buttons.Hide()

	scrollTable := container.NewScroll(table)

	content := container.NewBorder(
		upperActionBar,
		pagination_buttons, // bottom
		nil,                //left
		nil,                // right
		scrollTable,        //center
	)

	return content
}
