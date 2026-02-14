package inspect

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/guptarohit/asciigraph"
	daoCommon "github.com/jr-k/d4s/internal/dao/common"
	"github.com/jr-k/d4s/internal/ui/common"
	"github.com/jr-k/d4s/internal/ui/styles"
	"github.com/rivo/tview"
)

type StatsInspector struct {
	App           common.AppController
	ContainerID   string
	ContainerName string
	Layout        *tview.Flex

	// Text Mode
	Viewer *TextViewer

	// Graph Mode (Dashboard)
	Grid      *tview.Grid
	GraphCPU  *tview.TextView
	GraphMem  *tview.TextView
	GraphNet  *tview.TextView
	GraphDisk *tview.TextView

	Mode     string // "text" or "graph"
	StopChan chan struct{}

	cpuHistory       []float64
	memHistory       []float64
	netRxHistory     []float64
	netTxHistory     []float64
	diskReadHistory  []float64
	diskWriteHistory []float64

	// Previous values for rate calculation
	prevNetRx     float64
	prevNetTx     float64
	prevDiskRead  float64
	prevDiskWrite float64
	firstSample   bool

	maxPoints int

	// State management
	mu        sync.RWMutex
	lastStats map[string]interface{}

	// Dashboard Cached Values
	curCPU   float64
	curMem   uint64
	curLimit uint64
	curRx    float64
	curTx    float64
	curRead  float64
	curWrite float64
}

// Ensure interface compliance
var _ common.Inspector = (*StatsInspector)(nil)

func NewStatsInspector(containerID, containerName string) *StatsInspector {
	return newStatsInspectorInternal(containerID, containerName, "text")
}

func NewMonitorInspector(containerID, containerName string) *StatsInspector {
	return newStatsInspectorInternal(containerID, containerName, "graph")
}

func newStatsInspectorInternal(containerID, containerName, mode string) *StatsInspector {
	return &StatsInspector{
		ContainerID:   containerID,
		ContainerName: containerName,
		Mode:          mode,
		StopChan:      make(chan struct{}),
		maxPoints:     120,
		firstSample:   true,
	}
}

func (i *StatsInspector) GetID() string { return "inspect" }

func (i *StatsInspector) GetPrimitive() tview.Primitive {
	return i.Layout
}

func (i *StatsInspector) GetTitle() string {
	mode := "graph"
	if i.Mode == "text" {
		mode = "json"
	}
	id := i.ContainerID
	if len(id) > 12 {
		id = id[:12]
	}
	name := strings.TrimPrefix(i.ContainerName, "/")
	subject := fmt.Sprintf("%s@%s", name, id)

	filter, idx, count := "", 0, 0
	if i.Viewer != nil {
		filter, idx, count = i.Viewer.GetSearchInfo()
	}
	return FormatInspectorTitle("Stats", subject, mode, filter, idx, count)
}

func (i *StatsInspector) GetShortcuts() []string {
	shortcuts := []string{
		common.FormatSCHeader("esc", "Close"),
	}
	if i.Mode == "text" {
		shortcuts = append(shortcuts, common.FormatSCHeader("c", "Copy"))
		shortcuts = append(shortcuts, common.FormatSCHeader("n/p", "Next/Prev"))
	}
	return shortcuts
}

func (i *StatsInspector) OnMount(app common.AppController) {
	i.App = app

	// Initialize ViewModel for Text Mode
	i.Viewer = NewTextViewer(app)
	i.Viewer.TitleUpdateFunc = func() {
		// Update parent layout title when search status changes
		i.updateLayout()
	}

	// Initialize Grid (Graph Mode)
	i.Grid = tview.NewGrid().
		SetRows(0, 0).
		SetColumns(0, 0).
		SetBorders(false).
		SetGap(0, 0)

	i.Grid.SetBackgroundColor(styles.ColorBg)

	i.GraphCPU = createGraphView("CPU Usage")
	i.GraphMem = createGraphView("Memory Usage")
	i.GraphNet = createGraphView("Network I/O")
	i.GraphDisk = createGraphView("Disk I/O")

	i.Grid.AddItem(i.GraphCPU, 0, 0, 1, 1, 0, 0, true)
	i.Grid.AddItem(i.GraphMem, 0, 1, 1, 1, 0, 0, true)
	i.Grid.AddItem(i.GraphNet, 1, 0, 1, 1, 0, 0, true)
	i.Grid.AddItem(i.GraphDisk, 1, 1, 1, 1, 0, 0, true)

	i.Layout = tview.NewFlex().SetDirection(tview.FlexRow)
	// Keep outer frame opaque to prevent bleed-through
	i.Layout.SetBorder(true).SetTitleColor(styles.ColorTitle)
	i.Layout.SetBackgroundColor(styles.ColorBg)

	i.updateLayout()
	// Initial draw to ensure no empty boxes
	i.drawDashboard(0, 0, 0, 0, 0, 0, 0, nil, nil, nil, nil, nil, nil)
	i.startRefresher()
}

func createGraphView(title string) *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetWrap(false).
		SetTextAlign(tview.AlignLeft)

	tv.SetBorder(true).
		SetTitle(fmt.Sprintf(" %s ", title)).
		SetTitleColor(styles.ColorTitle).
		SetBorderColor(styles.ColorTitle)

	tv.SetBackgroundColor(styles.ColorBg)
	return tv
}

func (i *StatsInspector) updateLayout() {
	i.Layout.Clear()
	i.Layout.SetTitle(i.GetTitle())

	if i.Mode == "text" {
		i.Layout.AddItem(i.Viewer.GetPrimitive(), 0, 1, true)
	} else {
		i.Layout.AddItem(i.Grid, 0, 1, true)
	}
}

func (i *StatsInspector) OnUnmount() {
	close(i.StopChan)
}

func (i *StatsInspector) ApplyFilter(filter string) {
	if i.Mode == "text" {
		i.Viewer.ApplyFilter(filter)
		// No need to redraw, Viewer handles it
	}
}

func (i *StatsInspector) InputHandler(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyEsc {
		i.App.CloseInspector()
		return nil
	}

	if i.Mode == "text" {
		// Delegate input to Viewer
		if i.Viewer.InputHandler(event) {
			return nil
		}
		// Also allow native scrolling
		if handler := i.Viewer.View.InputHandler(); handler != nil {
			handler(event, func(p tview.Primitive) {})
			return nil
		}
	}

	return event
}

func (i *StatsInspector) startRefresher() {
	go i.tick()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				i.tick()
			case <-i.StopChan:
				return
			}
		}
	}()
}

func (i *StatsInspector) tick() {
	statsJSON, err := i.App.GetDocker().GetContainerStats(i.ContainerID)
	if err != nil {
		return
	}

	// Parse
	var v map[string]interface{}
	json.Unmarshal([]byte(statsJSON), &v)
	cpu, mem, limit, netRx, netTx, diskRead, diskWrite := daoCommon.CalculateStatsFromMap(v)

	i.mu.Lock()
	var rxRate, txRate, readRate, writeRate float64

	if i.firstSample {
		i.firstSample = false
	} else {
		rxRate = netRx - i.prevNetRx
		txRate = netTx - i.prevNetTx
		readRate = diskRead - i.prevDiskRead
		writeRate = diskWrite - i.prevDiskWrite

		if rxRate < 0 {
			rxRate = 0
		}
		if txRate < 0 {
			txRate = 0
		}
		if readRate < 0 {
			readRate = 0
		}
		if writeRate < 0 {
			writeRate = 0
		}
	}

	i.prevNetRx = netRx
	i.prevNetTx = netTx
	i.prevDiskRead = diskRead
	i.prevDiskWrite = diskWrite

	i.lastStats = v

	// Store calculated values
	i.curCPU = cpu
	i.curMem = mem
	i.curLimit = limit
	i.curRx = rxRate
	i.curTx = txRate
	i.curRead = readRate
	i.curWrite = writeRate

	// Update History
	i.cpuHistory = pushHistory(i.cpuHistory, cpu, i.maxPoints)

	memPct := 0.0
	if limit > 0 {
		memPct = float64(mem) / float64(limit) * 100.0
	}
	i.memHistory = pushHistory(i.memHistory, memPct, i.maxPoints)

	i.netRxHistory = pushHistory(i.netRxHistory, rxRate, i.maxPoints)
	i.netTxHistory = pushHistory(i.netTxHistory, txRate, i.maxPoints)

	i.diskReadHistory = pushHistory(i.diskReadHistory, readRate, i.maxPoints)
	i.diskWriteHistory = pushHistory(i.diskWriteHistory, writeRate, i.maxPoints)
	i.mu.Unlock()

	i.draw()
}

func (i *StatsInspector) draw() {
	i.mu.RLock()
	v := i.lastStats
	mode := i.Mode
	
	// Dashboard snapshots
	cpu := i.curCPU
	mem := i.curMem
	limit := i.curLimit
	rx := i.curRx
	tx := i.curTx
	dread := i.curRead
	dwrite := i.curWrite

	// Copy histories under lock to prevent race conditions with the tick loop
	cpuHist := make([]float64, len(i.cpuHistory))
	copy(cpuHist, i.cpuHistory)

	memHist := make([]float64, len(i.memHistory))
	copy(memHist, i.memHistory)

	rxHist := make([]float64, len(i.netRxHistory))
	copy(rxHist, i.netRxHistory)

	txHist := make([]float64, len(i.netTxHistory))
	copy(txHist, i.netTxHistory)

	readHist := make([]float64, len(i.diskReadHistory))
	copy(readHist, i.diskReadHistory)

	writeHist := make([]float64, len(i.diskWriteHistory))
	copy(writeHist, i.diskWriteHistory)
	i.mu.RUnlock()

	if mode == "text" {
		// Update Text View
		// Marshal logic is heavy, do it off UI thread (we are in background ticker usually here)
		pretty, _ := json.MarshalIndent(v, "", "  ")

		// Push to UI thread
		i.App.GetTviewApp().QueueUpdateDraw(func() {
			i.Viewer.Update(string(pretty), "json")
		})
	} else {
		// Update Dashboard
		i.App.GetTviewApp().QueueUpdateDraw(func() {
			if i.Mode != "graph" {
				return
			}
			i.drawDashboard(cpu, mem, limit, rx, tx, dread, dwrite, cpuHist, memHist, rxHist, txHist, readHist, writeHist)
		})
	}
}

func pushHistory(hist []float64, val float64, max int) []float64 {
	hist = append(hist, val)
	if len(hist) > max {
		return hist[1:]
	}
	return hist
}

func (i *StatsInspector) drawDashboard(cpu float64, mem uint64, limit uint64, rx, tx, dread, dwrite float64, cpuHist, memHist, rxHist, txHist, readHist, writeHist []float64) {
	// 1. CPU
	{
		label := fmt.Sprintf("Current: %.2f%%", cpu)
		i.renderGraph(i.GraphCPU, cpuHist, label, asciigraph.Green)
	}

	// 2. Memory
	{
		memPct := 0.0
		if limit > 0 {
			memPct = float64(mem) / float64(limit) * 100.0
		}
		label := fmt.Sprintf("Current: %.2f%% (%s / %s)",
			memPct, daoCommon.FormatBytes(int64(mem)), daoCommon.FormatBytes(int64(limit)))
		i.renderGraph(i.GraphMem, memHist, label, asciigraph.Green)
	}

	// 3. Network
	{
		label := fmt.Sprintf("[%s]●[-] Rx: %s/s  [%s]●[-] Tx: %s/s", styles.TagInfo, daoCommon.FormatBytes(int64(rx)), styles.TagCyan, daoCommon.FormatBytes(int64(tx)))
		i.renderGraphMany(i.GraphNet, [][]float64{rxHist, txHist}, label, []asciigraph.AnsiColor{asciigraph.Green, asciigraph.Cyan}, true)
	}

	// 4. Disk
	{
		label := fmt.Sprintf("[%s]●[-] Read: %s/s  [%s]●[-] Write: %s/s", styles.TagInfo, daoCommon.FormatBytes(int64(dread)), styles.TagError, daoCommon.FormatBytes(int64(dwrite)))
		i.renderGraphMany(i.GraphDisk, [][]float64{readHist, writeHist}, label, []asciigraph.AnsiColor{asciigraph.Green, asciigraph.Red}, true)
	}
}

func (i *StatsInspector) renderGraph(tv *tview.TextView, data []float64, label string, color asciigraph.AnsiColor) {
	_, _, w, h := tv.GetInnerRect()

	// Asciigraph needs explicit resizing
	// Height must be >= 1. Width must be positive.

	// Accounting for label text lines
	graphHeight := h - 2
	if graphHeight < 1 {
		graphHeight = 1
	}

	graphWidth := w - 8 // Reserve space for axis labels (approx)
	if graphWidth < 10 {
		graphWidth = 10
	}

	if len(data) == 0 {
		return
	}

	plot := asciigraph.Plot(data,
		asciigraph.Height(graphHeight),
		asciigraph.Width(graphWidth),
		asciigraph.SeriesColors(color),
		asciigraph.Caption(label),
	)

	// Reset bg to opaque before drawing
	tv.SetText("")
	// TranslateANSI converts the color codes from asciigraph
	tv.SetText(tview.TranslateANSI(plot))
}

func (i *StatsInspector) renderGraphMany(tv *tview.TextView, data [][]float64, label string, colors []asciigraph.AnsiColor, isBytes bool) {
	_, _, w, h := tv.GetInnerRect()

	maxVal := 0.0
	for _, series := range data {
		for _, v := range series {
			if v > maxVal {
				maxVal = v
			}
		}
	}

	scale := 1.0
	unit := ""
	if isBytes {
		scale, unit = determineGraphUnit(maxVal)
	}

	plotData := data
	if scale != 1.0 {
		plotData = make([][]float64, len(data))
		for idx, series := range data {
			scaled := make([]float64, len(series))
			for j, val := range series {
				scaled[j] = val / scale
			}
			plotData[idx] = scaled
		}
	}

	// Asciigraph needs explicit resizing
	// Height must be >= 1. Width must be positive.

	// Accounting for label text lines
	graphHeight := h - 2
	if graphHeight < 1 {
		graphHeight = 1
	}

	// Reserve space for axis labels
	// Max value 1023.99 + " MB" -> ~10 chars + axis ~2 chars = 12.
	// We add some buffer to prevent wrapping issues.
	axisReservation := 15
	if isBytes {
		axisReservation = 18 // More space for units
	}

	graphWidth := w - axisReservation
	if graphWidth < 10 {
		graphWidth = 10
	}

	if len(plotData) == 0 {
		return
	}

	hasData := false
	for _, series := range plotData {
		if len(series) > 0 {
			hasData = true
			break
		}
	}
	if !hasData {
		return
	}

	opts := []asciigraph.Option{
		asciigraph.Height(graphHeight),
		asciigraph.Width(graphWidth),
		asciigraph.SeriesColors(colors...),
		asciigraph.Caption(label),
	}

	if isBytes {
		prec := uint(2)
		if unit == "B" {
			prec = 0
		}
		opts = append(opts, asciigraph.Precision(prec))
	}

	plot := asciigraph.PlotMany(plotData, opts...)

	if isBytes && unit != "" {
		plot = addUnitToAxis(plot, unit)
	}

	// Reset bg to opaque before drawing
	tv.SetText("")
	// TranslateANSI converts the color codes from asciigraph
	tv.SetText(tview.TranslateANSI(plot))
}

func determineGraphUnit(maxVal float64) (float64, string) {
	if maxVal >= 1024*1024*1024*1024 {
		return 1024 * 1024 * 1024 * 1024, "TB"
	}
	if maxVal >= 1024*1024*1024 {
		return 1024 * 1024 * 1024, "GB"
	}
	if maxVal >= 1024*1024 {
		return 1024 * 1024, "MB"
	}
	if maxVal >= 1024 {
		return 1024, "KB"
	}
	return 1, "B"
}

func addUnitToAxis(plotString, unit string) string {
	lines := strings.Split(plotString, "\n")
	unitLen := len(unit) + 1 // +1 for space
	for i, line := range lines {
		// Find axis separator
		idx := strings.Index(line, "┤")
		if idx == -1 {
			idx = strings.Index(line, "┼")
		}

		if idx != -1 {
			prefix := line[:idx]
			suffix := line[idx:]

			// Check if prefix has numbers (is a label) or just spaces
			// Asciigraph aligns numbers right, so left padding is spaces.
			if strings.TrimSpace(prefix) != "" {
				// It's a label, append unit
				// We need to be careful not to make the line too long, hindering the graph?
				// Actually, we reserved space via axisReservation.
				lines[i] = prefix + " " + unit + suffix
			} else {
				// It's whitespace padding
				// We need to replicate the padding length added by " unit"
				lines[i] = prefix + strings.Repeat(" ", unitLen) + suffix
			}
		}
	}
	return strings.Join(lines, "\n")
}

