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
	CPUUsage     float64
	CPUCores     int
	MemTotal     uint64
	MemUsed      uint64
	MemFree      uint64
	MemPercent   float64
	SwapTotal    uint64
	SwapUsed     uint64
	SwapPercent  float64
	DiskTotal    uint64
	DiskUsed     uint64
	DiskPercent  float64
	LoadAvg      [3]float64
	Uptime       string
	Hostname     string
	OS           string
	Kernel       string
	ProcessCount int
	UserCount    int
	NetworkRX    string
	NetworkTX    string
}

// Dashboard displays health metrics
type Dashboard struct {
	view      *tview.Flex
	theme     *config.Theme
	sshClient *ssh.Client
	host      *ssh.HostEntry

	// Widgets
	cpuPanel     *tview.TextView
	memPanel     *tview.TextView
	diskPanel    *tview.TextView
	systemPanel  *tview.TextView
	networkPanel *tview.TextView
	quickStats   *tview.TextView
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
	// CPU Panel with big gauge
	d.cpuPanel = tview.NewTextView()
	d.cpuPanel.SetDynamicColors(true)
	d.cpuPanel.SetBorder(true)
	d.cpuPanel.SetTitle(" 🔥 CPU ")
	d.cpuPanel.SetTitleColor(d.theme.Primary)
	d.cpuPanel.SetBorderColor(d.theme.Border)
	d.cpuPanel.SetBackgroundColor(d.theme.Background)

	// Memory Panel
	d.memPanel = tview.NewTextView()
	d.memPanel.SetDynamicColors(true)
	d.memPanel.SetBorder(true)
	d.memPanel.SetTitle(" 💾 Memory & Swap ")
	d.memPanel.SetTitleColor(d.theme.Primary)
	d.memPanel.SetBorderColor(d.theme.Border)
	d.memPanel.SetBackgroundColor(d.theme.Background)

	// Disk Panel
	d.diskPanel = tview.NewTextView()
	d.diskPanel.SetDynamicColors(true)
	d.diskPanel.SetBorder(true)
	d.diskPanel.SetTitle(" 💿 Disk Usage ")
	d.diskPanel.SetTitleColor(d.theme.Primary)
	d.diskPanel.SetBorderColor(d.theme.Border)
	d.diskPanel.SetBackgroundColor(d.theme.Background)

	// System Info Panel
	d.systemPanel = tview.NewTextView()
	d.systemPanel.SetDynamicColors(true)
	d.systemPanel.SetBorder(true)
	d.systemPanel.SetTitle(" 🖥️  System Info ")
	d.systemPanel.SetTitleColor(d.theme.Primary)
	d.systemPanel.SetBorderColor(d.theme.Border)
	d.systemPanel.SetBackgroundColor(d.theme.Background)

	// Network Panel
	d.networkPanel = tview.NewTextView()
	d.networkPanel.SetDynamicColors(true)
	d.networkPanel.SetBorder(true)
	d.networkPanel.SetTitle(" 🌐 Network & Load ")
	d.networkPanel.SetTitleColor(d.theme.Primary)
	d.networkPanel.SetBorderColor(d.theme.Border)
	d.networkPanel.SetBackgroundColor(d.theme.Background)

	// Quick Stats (bottom bar)
	d.quickStats = tview.NewTextView()
	d.quickStats.SetDynamicColors(true)
	d.quickStats.SetBackgroundColor(d.theme.Muted)
	d.quickStats.SetTextAlign(tview.AlignCenter)

	// Top Row: CPU | Memory | Disk
	topRow := tview.NewFlex()
	topRow.AddItem(d.cpuPanel, 0, 1, false)
	topRow.AddItem(d.memPanel, 0, 1, false)
	topRow.AddItem(d.diskPanel, 0, 1, false)

	// Bottom Row: System Info | Network
	bottomRow := tview.NewFlex()
	bottomRow.AddItem(d.systemPanel, 0, 2, false)
	bottomRow.AddItem(d.networkPanel, 0, 1, false)

	// Main Layout
	d.view = tview.NewFlex().SetDirection(tview.FlexRow)
	d.view.AddItem(topRow, 0, 1, false)
	d.view.AddItem(bottomRow, 0, 1, false)
	d.view.AddItem(d.quickStats, 2, 0, false)
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
		`cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d= -f2 | tr -d '"'`)
	m.OS = strings.TrimSpace(osOutput)

	kernelOutput, _ := d.sshClient.RunCommand(*d.host, "uname -r")
	m.Kernel = strings.TrimSpace(kernelOutput)

	uptimeOutput, _ := d.sshClient.RunCommand(*d.host, "uptime -p 2>/dev/null || uptime")
	m.Uptime = strings.TrimSpace(uptimeOutput)

	// Get process count
	procOutput, _ := d.sshClient.RunCommand(*d.host, "ps aux | wc -l")
	if proc, err := strconv.Atoi(strings.TrimSpace(procOutput)); err == nil {
		m.ProcessCount = proc - 1 // subtract header
	}

	// Get user count
	userOutput, _ := d.sshClient.RunCommand(*d.host, "who | wc -l")
	if users, err := strconv.Atoi(strings.TrimSpace(userOutput)); err == nil {
		m.UserCount = users
	}

	// Get network stats
	rxOutput, _ := d.sshClient.RunCommand(*d.host,
		`cat /sys/class/net/*/statistics/rx_bytes 2>/dev/null | awk '{sum+=$1} END {print sum}'`)
	if rx, err := strconv.ParseUint(strings.TrimSpace(rxOutput), 10, 64); err == nil {
		m.NetworkRX = formatBytes(rx)
	}
	txOutput, _ := d.sshClient.RunCommand(*d.host,
		`cat /sys/class/net/*/statistics/tx_bytes 2>/dev/null | awk '{sum+=$1} END {print sum}'`)
	if tx, err := strconv.ParseUint(strings.TrimSpace(txOutput), 10, 64); err == nil {
		m.NetworkTX = formatBytes(tx)
	}

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
		if m.SwapTotal > 0 {
			m.SwapPercent = float64(m.SwapUsed) / float64(m.SwapTotal) * 100
		}
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
	errColor := colorToTag(d.theme.Error)
	muted := colorToTag(d.theme.Muted)
	highlight := colorToTag(d.theme.Highlight)

	// CPU Panel - Big centered percentage with visual bar
	cpuColor := success
	if m.CPUUsage > 80 {
		cpuColor = errColor
	} else if m.CPUUsage > 50 {
		cpuColor = warning
	}
	cpuIcon := "▼"
	if m.CPUUsage > 50 {
		cpuIcon = "▲"
	}
	d.cpuPanel.SetText(fmt.Sprintf(`
  [%s]%s CPU USAGE[white]
  
  [%s]  ╔═══════════════════╗[white]
  [%s]  ║[white] [%s]%6.1f%%[white]          [%s]║[white]
  [%s]  ╚═══════════════════╝[white]

  %s

  [%s]Cores:[white] [%s]%d[white]  │  [%s]Threads/Core:[white] [%s]2[white]`,
		cpuColor, cpuIcon,
		primary,
		primary, cpuColor, m.CPUUsage, primary,
		primary,
		d.fancyProgressBar(m.CPUUsage, cpuColor, 25),
		muted, highlight, m.CPUCores, muted, highlight))

	// Memory Panel - RAM + Swap with visual bars
	memColor := success
	if m.MemPercent > 80 {
		memColor = errColor
	} else if m.MemPercent > 50 {
		memColor = warning
	}

	swapColor := success
	if m.SwapPercent > 50 {
		swapColor = warning
	}
	if m.SwapPercent > 80 {
		swapColor = errColor
	}

	swapLine := ""
	if m.SwapTotal > 0 {
		swapLine = fmt.Sprintf(`
  [%s]SWAP[white]  %s
         [%s]%.1f%%[white] (%s / %s)`,
			muted, d.fancyProgressBar(m.SwapPercent, swapColor, 25),
			swapColor, m.SwapPercent, formatBytes(m.SwapUsed), formatBytes(m.SwapTotal))
	} else {
		swapLine = fmt.Sprintf("\n  [%s]SWAP[white]  [%s]Not configured[white]", muted, muted)
	}

	d.memPanel.SetText(fmt.Sprintf(`
  [%s]RAM [white]  %s
         [%s]%.1f%%[white] (%s / %s)
%s

  [%s]Free:[white] [%s]%s[white]`,
		muted, d.fancyProgressBar(m.MemPercent, memColor, 25),
		memColor, m.MemPercent, formatBytes(m.MemUsed), formatBytes(m.MemTotal),
		swapLine,
		muted, success, formatBytes(m.MemFree)))

	// Disk Panel
	diskColor := success
	if m.DiskPercent > 90 {
		diskColor = errColor
	} else if m.DiskPercent > 70 {
		diskColor = warning
	}

	diskIcon := "●"
	if m.DiskPercent > 80 {
		diskIcon = "⚠"
	}

	d.diskPanel.SetText(fmt.Sprintf(`
  [%s]%s ROOT (/)
  
  %s

  [%s]Used:[white]  [%s]%s[white]
  [%s]Total:[white] [%s]%s[white]
  [%s]Free:[white]  [%s]%s[white]`,
		diskColor, diskIcon,
		d.bigProgressBar(m.DiskPercent, diskColor, 25),
		muted, diskColor, formatBytes(m.DiskUsed),
		muted, highlight, formatBytes(m.DiskTotal),
		muted, success, formatBytes(m.DiskTotal-m.DiskUsed)))

	// System Panel - More info
	d.systemPanel.SetText(fmt.Sprintf(`
  [%s]┌─ SYSTEM ─────────────────────────────────────┐[white]
  [%s]│[white]
  [%s]│[white]  [%s]Hostname:[white]   [%s]%s[white]
  [%s]│[white]  [%s]OS:[white]         [%s]%s[white]
  [%s]│[white]  [%s]Kernel:[white]     [%s]%s[white]
  [%s]│[white]  [%s]Uptime:[white]     [%s]%s[white]
  [%s]│[white]
  [%s]│[white]  [%s]Processes:[white]  [%s]%d[white] running
  [%s]│[white]  [%s]Users:[white]      [%s]%d[white] logged in
  [%s]│[white]
  [%s]└────────────────────────────────────────────────┘[white]`,
		primary,
		primary,
		primary, muted, highlight, m.Hostname,
		primary, muted, highlight, truncateStr(m.OS, 35),
		primary, muted, highlight, m.Kernel,
		primary, muted, success, m.Uptime,
		primary,
		primary, muted, highlight, m.ProcessCount,
		primary, muted, highlight, m.UserCount,
		primary,
		primary))

	// Network & Load Panel
	loadColor := success
	if m.LoadAvg[0] > float64(m.CPUCores) {
		loadColor = errColor
	} else if m.LoadAvg[0] > float64(m.CPUCores)*0.7 {
		loadColor = warning
	}

	d.networkPanel.SetText(fmt.Sprintf(`
  [%s]⬇ RECEIVED[white]
    [%s]%s[white]

  [%s]⬆ TRANSMITTED[white]
    [%s]%s[white]

  [%s]━━ LOAD AVERAGE ━━[white]

  [%s]1m:[white]  [%s]%.2f[white]
  [%s]5m:[white]  [%s]%.2f[white]
  [%s]15m:[white] [%s]%.2f[white]`,
		success,
		highlight, m.NetworkRX,
		warning,
		highlight, m.NetworkTX,
		muted,
		muted, loadColor, m.LoadAvg[0],
		muted, loadColor, m.LoadAvg[1],
		muted, loadColor, m.LoadAvg[2]))

	// Quick Stats Bar
	d.quickStats.SetText(fmt.Sprintf(
		"  [%s]CPU:[white] [%s]%.0f%%[white]  │  [%s]RAM:[white] [%s]%.0f%%[white]  │  [%s]DISK:[white] [%s]%.0f%%[white]  │  [%s]LOAD:[white] [%s]%.2f[white]  │  [%s]PROCS:[white] [%s]%d[white]  │  [%s]Press[white] [%s]r[white] [%s]to refresh[white]",
		muted, cpuColor, m.CPUUsage,
		muted, memColor, m.MemPercent,
		muted, diskColor, m.DiskPercent,
		muted, loadColor, m.LoadAvg[0],
		muted, highlight, m.ProcessCount,
		muted, highlight, muted))
}

// fancyProgressBar creates a colorful progress bar with gradient effect
func (d *Dashboard) fancyProgressBar(percent float64, color string, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return fmt.Sprintf("[%s]%s[white]", color, bar)
}

// bigProgressBar creates a larger visual progress bar
func (d *Dashboard) bigProgressBar(percent float64, color string, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	bar := fmt.Sprintf("  [%s]", color)
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "▓"
		} else {
			bar += "░"
		}
	}
	bar += fmt.Sprintf("[white] [%s]%.1f%%[white]", color, percent)
	return bar
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

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
