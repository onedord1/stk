package health

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/systask/systask/internal/config"
	"github.com/systask/systask/internal/ssh"
)

// Metrics holds system health data
type Metrics struct {
	CPUUsage    float64
	CPUCores    int
	MemTotal    uint64
	MemUsed     uint64
	MemFree     uint64
	MemPercent  float64
	SwapTotal   uint64
	SwapUsed    uint64
	DiskTotal   uint64
	DiskUsed    uint64
	DiskPercent float64
	LoadAvg     [3]float64
	Uptime      string
	Hostname    string
	OS          string
	Kernel      string
	NetworkRX   uint64
	NetworkTX   uint64
}

// Dashboard displays health metrics
type Dashboard struct {
	view      *tview.Flex
	theme     *config.Theme
	sshClient *ssh.Client
	host      *ssh.HostEntry

	// Widgets
	cpuGauge  *tview.TextView
	memGauge  *tview.TextView
	diskGauge *tview.TextView
	infoBox   *tview.TextView
	loadBox   *tview.TextView
}

// NewDashboard creates a new health dashboard
func NewDashboard(theme *config.Theme, client *ssh.Client) *Dashboard {
	d := &Dashboard{
		theme:     theme,
		sshClient: client,
	}
	d.build()
	return d
}

// build constructs the dashboard layout
func (d *Dashboard) build() {
	// CPU gauge
	d.cpuGauge = tview.NewTextView()
	d.cpuGauge.SetDynamicColors(true)
	d.cpuGauge.SetBorder(true)
	d.cpuGauge.SetTitle(" CPU ")
	d.cpuGauge.SetTitleColor(d.theme.Primary)
	d.cpuGauge.SetBorderColor(d.theme.Border)
	d.cpuGauge.SetBackgroundColor(d.theme.Background)

	// Memory gauge
	d.memGauge = tview.NewTextView()
	d.memGauge.SetDynamicColors(true)
	d.memGauge.SetBorder(true)
	d.memGauge.SetTitle(" Memory ")
	d.memGauge.SetTitleColor(d.theme.Primary)
	d.memGauge.SetBorderColor(d.theme.Border)
	d.memGauge.SetBackgroundColor(d.theme.Background)

	// Disk gauge
	d.diskGauge = tview.NewTextView()
	d.diskGauge.SetDynamicColors(true)
	d.diskGauge.SetBorder(true)
	d.diskGauge.SetTitle(" Disk (/) ")
	d.diskGauge.SetTitleColor(d.theme.Primary)
	d.diskGauge.SetBorderColor(d.theme.Border)
	d.diskGauge.SetBackgroundColor(d.theme.Background)

	// System info
	d.infoBox = tview.NewTextView()
	d.infoBox.SetDynamicColors(true)
	d.infoBox.SetBorder(true)
	d.infoBox.SetTitle(" System Info ")
	d.infoBox.SetTitleColor(d.theme.Primary)
	d.infoBox.SetBorderColor(d.theme.Border)
	d.infoBox.SetBackgroundColor(d.theme.Background)

	// Load average
	d.loadBox = tview.NewTextView()
	d.loadBox.SetDynamicColors(true)
	d.loadBox.SetBorder(true)
	d.loadBox.SetTitle(" Load Average ")
	d.loadBox.SetTitleColor(d.theme.Primary)
	d.loadBox.SetBorderColor(d.theme.Border)
	d.loadBox.SetBackgroundColor(d.theme.Background)

	// Layout: gauges on top row, info on bottom
	gaugeRow := tview.NewFlex()
	gaugeRow.AddItem(d.cpuGauge, 0, 1, false)
	gaugeRow.AddItem(d.memGauge, 0, 1, false)
	gaugeRow.AddItem(d.diskGauge, 0, 1, false)

	infoRow := tview.NewFlex()
	infoRow.AddItem(d.infoBox, 0, 2, false)
	infoRow.AddItem(d.loadBox, 0, 1, false)

	d.view = tview.NewFlex().SetDirection(tview.FlexRow)
	d.view.AddItem(gaugeRow, 0, 1, false)
	d.view.AddItem(infoRow, 0, 1, false)
	d.view.SetBackgroundColor(d.theme.Background)
}

// SetHost sets the current host
func (d *Dashboard) SetHost(host *ssh.HostEntry) {
	d.host = host
}

// Refresh updates metrics from the remote host
func (d *Dashboard) Refresh() error {
	if d.host == nil {
		return fmt.Errorf("no host set")
	}

	metrics, err := d.fetchMetrics()
	if err != nil {
		return err
	}

	d.updateDisplay(metrics)
	return nil
}

// fetchMetrics retrieves metrics from remote host
func (d *Dashboard) fetchMetrics() (*Metrics, error) {
	m := &Metrics{}

	// Get CPU usage
	cpuOutput, _ := d.sshClient.RunCommand(*d.host,
		"top -bn1 | grep 'Cpu(s)' | awk '{print $2}' | cut -d'%' -f1")
	if cpu, err := strconv.ParseFloat(strings.TrimSpace(cpuOutput), 64); err == nil {
		m.CPUUsage = cpu
	}

	// Get CPU cores
	coresOutput, _ := d.sshClient.RunCommand(*d.host, "nproc")
	if cores, err := strconv.Atoi(strings.TrimSpace(coresOutput)); err == nil {
		m.CPUCores = cores
	}

	// Get memory info
	memOutput, _ := d.sshClient.RunCommand(*d.host, "free -b | grep Mem")
	d.parseMemory(memOutput, m)

	// Get swap info
	swapOutput, _ := d.sshClient.RunCommand(*d.host, "free -b | grep Swap")
	d.parseSwap(swapOutput, m)

	// Get disk usage for /
	diskOutput, _ := d.sshClient.RunCommand(*d.host, "df -B1 / | tail -1")
	d.parseDisk(diskOutput, m)

	// Get load average
	loadOutput, _ := d.sshClient.RunCommand(*d.host, "cat /proc/loadavg")
	d.parseLoad(loadOutput, m)

	// Get system info
	hostnameOutput, _ := d.sshClient.RunCommand(*d.host, "hostname")
	m.Hostname = strings.TrimSpace(hostnameOutput)

	osOutput, _ := d.sshClient.RunCommand(*d.host,
		"cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"'")
	m.OS = strings.TrimSpace(osOutput)

	kernelOutput, _ := d.sshClient.RunCommand(*d.host, "uname -r")
	m.Kernel = strings.TrimSpace(kernelOutput)

	uptimeOutput, _ := d.sshClient.RunCommand(*d.host, "uptime -p 2>/dev/null || uptime")
	m.Uptime = strings.TrimSpace(uptimeOutput)

	return m, nil
}

// parseMemory extracts memory info from free output
func (d *Dashboard) parseMemory(output string, m *Metrics) {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(output, -1)
	if len(matches) >= 3 {
		m.MemTotal, _ = strconv.ParseUint(matches[0], 10, 64)
		m.MemUsed, _ = strconv.ParseUint(matches[1], 10, 64)
		m.MemFree, _ = strconv.ParseUint(matches[2], 10, 64)
		if m.MemTotal > 0 {
			m.MemPercent = float64(m.MemUsed) / float64(m.MemTotal) * 100
		}
	}
}

// parseSwap extracts swap info
func (d *Dashboard) parseSwap(output string, m *Metrics) {
	re := regexp.MustCompile(`\d+`)
	matches := re.FindAllString(output, -1)
	if len(matches) >= 2 {
		m.SwapTotal, _ = strconv.ParseUint(matches[0], 10, 64)
		m.SwapUsed, _ = strconv.ParseUint(matches[1], 10, 64)
	}
}

// parseDisk extracts disk info
func (d *Dashboard) parseDisk(output string, m *Metrics) {
	fields := strings.Fields(output)
	if len(fields) >= 5 {
		m.DiskTotal, _ = strconv.ParseUint(fields[1], 10, 64)
		m.DiskUsed, _ = strconv.ParseUint(fields[2], 10, 64)
		percentStr := strings.TrimSuffix(fields[4], "%")
		m.DiskPercent, _ = strconv.ParseFloat(percentStr, 64)
	}
}

// parseLoad extracts load average
func (d *Dashboard) parseLoad(output string, m *Metrics) {
	fields := strings.Fields(output)
	if len(fields) >= 3 {
		m.LoadAvg[0], _ = strconv.ParseFloat(fields[0], 64)
		m.LoadAvg[1], _ = strconv.ParseFloat(fields[1], 64)
		m.LoadAvg[2], _ = strconv.ParseFloat(fields[2], 64)
	}
}

// updateDisplay updates UI with metrics
func (d *Dashboard) updateDisplay(m *Metrics) {
	primary := colorToTag(d.theme.Primary)
	success := colorToTag(d.theme.Success)
	warning := colorToTag(d.theme.Warning)
	err := colorToTag(d.theme.Error)

	// CPU gauge
	cpuColor := success
	if m.CPUUsage > 80 {
		cpuColor = err
	} else if m.CPUUsage > 50 {
		cpuColor = warning
	}
	d.cpuGauge.SetText(fmt.Sprintf(
		"\n  [%s]%.1f%%[white]\n\n  %s\n  %d cores",
		cpuColor, m.CPUUsage,
		d.progressBar(m.CPUUsage, cpuColor),
		m.CPUCores,
	))

	// Memory gauge
	memColor := success
	if m.MemPercent > 80 {
		memColor = err
	} else if m.MemPercent > 50 {
		memColor = warning
	}
	d.memGauge.SetText(fmt.Sprintf(
		"\n  [%s]%.1f%%[white]\n\n  %s\n  %s / %s",
		memColor, m.MemPercent,
		d.progressBar(m.MemPercent, memColor),
		formatBytes(m.MemUsed), formatBytes(m.MemTotal),
	))

	// Disk gauge
	diskColor := success
	if m.DiskPercent > 90 {
		diskColor = err
	} else if m.DiskPercent > 70 {
		diskColor = warning
	}
	d.diskGauge.SetText(fmt.Sprintf(
		"\n  [%s]%.1f%%[white]\n\n  %s\n  %s / %s",
		diskColor, m.DiskPercent,
		d.progressBar(m.DiskPercent, diskColor),
		formatBytes(m.DiskUsed), formatBytes(m.DiskTotal),
	))

	// System info
	d.infoBox.SetText(fmt.Sprintf(
		"\n  [%s]Hostname:[white] %s\n"+
			"  [%s]OS:[white] %s\n"+
			"  [%s]Kernel:[white] %s\n"+
			"  [%s]Uptime:[white] %s",
		primary, m.Hostname,
		primary, m.OS,
		primary, m.Kernel,
		primary, m.Uptime,
	))

	// Load average
	loadColor := success
	if m.LoadAvg[0] > float64(m.CPUCores) {
		loadColor = err
	} else if m.LoadAvg[0] > float64(m.CPUCores)*0.7 {
		loadColor = warning
	}
	d.loadBox.SetText(fmt.Sprintf(
		"\n  [%s]1 min:[white]  %.2f\n"+
			"  [%s]5 min:[white]  %.2f\n"+
			"  [%s]15 min:[white] %.2f",
		loadColor, m.LoadAvg[0],
		loadColor, m.LoadAvg[1],
		loadColor, m.LoadAvg[2],
	))
}

// progressBar creates a text progress bar
func (d *Dashboard) progressBar(percent float64, color string) string {
	width := 20
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("[%s]%s[white]", color, bar)
}

// View returns the dashboard view
func (d *Dashboard) View() *tview.Flex {
	return d.view
}

// Helper functions
func colorToTag(c tcell.Color) string {
	r, g, b := c.RGB()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
