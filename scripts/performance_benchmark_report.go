//go:build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type benchmarkRun struct {
	StartedAt   string            `json:"started_at"`
	FinishedAt  string            `json:"finished_at"`
	CommitSHA   string            `json:"commit_sha"`
	RefName     string            `json:"ref_name"`
	WorkflowURL string            `json:"workflow_url"`
	RunnerOS    string            `json:"runner_os"`
	GoVersion   string            `json:"go_version"`
	GOOS        string            `json:"goos"`
	GOARCH      string            `json:"goarch"`
	CPU         string            `json:"cpu"`
	Verdict     string            `json:"verdict"`
	Benchmarks  []benchmarkResult `json:"benchmarks"`
}

type benchmarkResult struct {
	Package     string   `json:"package"`
	Name        string   `json:"name"`
	Iterations  int64    `json:"iterations"`
	NsPerOp     float64  `json:"ns_per_op"`
	BytesPerOp  *float64 `json:"bytes_per_op,omitempty"`
	AllocsPerOp *float64 `json:"allocs_per_op,omitempty"`
}

type reportSummary struct {
	Date           string           `json:"date"`
	Verdict        string           `json:"verdict"`
	BenchmarkCount int              `json:"benchmark_count"`
	PackageCount   int              `json:"package_count"`
	Slowest        *benchmarkResult `json:"slowest,omitempty"`
	HighestAlloc   *benchmarkResult `json:"highest_alloc,omitempty"`
	WorkflowURL    string           `json:"workflow_url"`
	RefName        string           `json:"ref_name"`
	CommitSHA      string           `json:"commit_sha"`
}

type comparison struct {
	PreviousNs float64
	DeltaPct   float64
}

func main() {
	input := flag.String("input", "benchmark.txt", "raw go test benchmark output")
	reportDir := flag.String("report-dir", "docs/reports/performance", "directory for Markdown reports")
	docsSrc := flag.String("docs-site-src", "docs-site/src", "mdBook source directory")
	reportDate := flag.String("date", time.Now().UTC().Format("2006-01-02"), "report date in YYYY-MM-DD format")
	retention := flag.Int("retention", 3, "number of dated reports to retain")
	summaryOut := flag.String("summary-output", "", "optional JSON summary output path")
	flag.Parse()

	data, err := os.ReadFile(*input)
	if err != nil {
		fail("read benchmark input: %v", err)
	}

	run := parseBenchmarkOutput(string(data))
	run.StartedAt = envOr("BENCHMARK_STARTED_AT", "")
	run.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	run.CommitSHA = envOr("GITHUB_SHA", "")
	run.RefName = envOr("GITHUB_REF_NAME", "")
	run.WorkflowURL = workflowURL()
	run.RunnerOS = envOr("RUNNER_OS", "")
	run.GoVersion = envOr("GOVERSION", "")
	run.Verdict = "recorded"
	if len(run.Benchmarks) == 0 {
		fail("no benchmark results parsed from %s", *input)
	}

	if err := os.MkdirAll(*reportDir, 0o755); err != nil {
		fail("create report dir: %v", err)
	}

	previous := readPreviousBenchmarks(filepath.Join(*reportDir, "latest.md"))
	reportName := *reportDate + ".md"
	reportBody := renderReport(*reportDate, run, previous)
	if err := os.WriteFile(filepath.Join(*reportDir, reportName), []byte(reportBody), 0o644); err != nil {
		fail("write dated report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(*reportDir, "latest.md"), []byte(reportBody), 0o644); err != nil {
		fail("write latest report: %v", err)
	}

	reports, err := pruneReports(*reportDir, *retention)
	if err != nil {
		fail("prune reports: %v", err)
	}
	if !contains(reports, reportName) {
		reports = append([]string{reportName}, reports...)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(reports)))

	if err := writeDocsPages(*docsSrc, reports, run); err != nil {
		fail("write docs pages: %v", err)
	}

	if *summaryOut != "" {
		if err := writeSummary(*summaryOut, *reportDate, run); err != nil {
			fail("write summary: %v", err)
		}
	}
}

func parseBenchmarkOutput(output string) benchmarkRun {
	var run benchmarkRun
	var pkg string
	var pending string

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "goos:"):
			run.GOOS = strings.TrimSpace(strings.TrimPrefix(line, "goos:"))
		case strings.HasPrefix(line, "goarch:"):
			run.GOARCH = strings.TrimSpace(strings.TrimPrefix(line, "goarch:"))
		case strings.HasPrefix(line, "cpu:"):
			run.CPU = strings.TrimSpace(strings.TrimPrefix(line, "cpu:"))
		case strings.HasPrefix(line, "pkg:"):
			pkg = strings.TrimSpace(strings.TrimPrefix(line, "pkg:"))
		case strings.HasPrefix(line, "Benchmark"):
			fields := strings.Fields(line)
			if len(fields) >= 2 && isInt(fields[1]) {
				if result, ok := parseBenchmarkFields(pkg, fields); ok {
					run.Benchmarks = append(run.Benchmarks, result)
				}
				pending = ""
				continue
			}
			if len(fields) > 0 {
				pending = fields[0]
			}
		default:
			fields := strings.Fields(line)
			if pending != "" && len(fields) >= 3 && isInt(fields[0]) {
				fields = append([]string{pending}, fields...)
				if result, ok := parseBenchmarkFields(pkg, fields); ok {
					run.Benchmarks = append(run.Benchmarks, result)
				}
				pending = ""
			}
		}
	}

	return run
}

func parseBenchmarkFields(pkg string, fields []string) (benchmarkResult, bool) {
	if len(fields) < 4 {
		return benchmarkResult{}, false
	}
	iterations, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return benchmarkResult{}, false
	}
	result := benchmarkResult{
		Package:    pkg,
		Name:       stripBenchmarkSuffix(fields[0]),
		Iterations: iterations,
	}
	for i := 2; i < len(fields); i++ {
		switch fields[i] {
		case "ns/op":
			if i == 0 {
				continue
			}
			result.NsPerOp = parseFloat(fields[i-1])
		case "B/op":
			value := parseFloat(fields[i-1])
			result.BytesPerOp = &value
		case "allocs/op":
			value := parseFloat(fields[i-1])
			result.AllocsPerOp = &value
		}
	}
	if result.NsPerOp == 0 {
		return benchmarkResult{}, false
	}
	return result, true
}

func renderReport(date string, run benchmarkRun, previous map[string]benchmarkResult) string {
	comparisons := compareBenchmarks(run.Benchmarks, previous)
	var b strings.Builder
	fmt.Fprintf(&b, "# Performance Benchmark Report - %s\n\n", date)
	fmt.Fprintln(&b, "## English")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "**Verdict:** RECORDED")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "This report is informational. GitHub hosted runners can vary, so this workflow records benchmark visibility and trend signals without failing on timing changes.")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### Run Context")
	fmt.Fprintln(&b)
	writeContextTable(&b, run)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### Summary")
	fmt.Fprintln(&b)
	writeSummaryTable(&b, run)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### Slowest Benchmarks")
	fmt.Fprintln(&b)
	writeBenchmarkTable(&b, topByNs(run.Benchmarks, 10), comparisons)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### Highest Allocation Benchmarks")
	fmt.Fprintln(&b)
	writeBenchmarkTable(&b, topByBytes(run.Benchmarks, 10), comparisons)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### All Benchmark Results")
	fmt.Fprintln(&b)
	writeBenchmarkTable(&b, run.Benchmarks, comparisons)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## 中文")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "**结论：** 已记录")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "本报告只提供性能可见性。GitHub 托管 runner 存在波动，因此当前工作流记录趋势信号，但不会因为耗时变化直接失败。")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### 运行上下文")
	fmt.Fprintln(&b)
	writeContextTable(&b, run)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### 摘要")
	fmt.Fprintln(&b)
	writeSummaryTableCN(&b, run)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### 最慢 Benchmark")
	fmt.Fprintln(&b)
	writeBenchmarkTable(&b, topByNs(run.Benchmarks, 10), comparisons)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "### 最高内存分配 Benchmark")
	fmt.Fprintln(&b)
	writeBenchmarkTable(&b, topByBytes(run.Benchmarks, 10), comparisons)
	return b.String()
}

func writeContextTable(b *strings.Builder, run benchmarkRun) {
	rows := [][2]string{
		{"Started", run.StartedAt},
		{"Finished", run.FinishedAt},
		{"Commit", shortSHA(run.CommitSHA)},
		{"Ref", run.RefName},
		{"Workflow", markdownLink(run.WorkflowURL)},
		{"Runner OS", run.RunnerOS},
		{"Go", run.GoVersion},
		{"GOOS/GOARCH", strings.Trim(run.GOOS+"/"+run.GOARCH, "/")},
		{"CPU", run.CPU},
	}
	fmt.Fprintln(b, "| Field | Value |")
	fmt.Fprintln(b, "| --- | --- |")
	for _, row := range rows {
		fmt.Fprintf(b, "| %s | %s |\n", row[0], escapeTable(row[1]))
	}
}

func writeSummaryTable(b *strings.Builder, run benchmarkRun) {
	packages := packageCount(run.Benchmarks)
	slowest := firstBenchmark(topByNs(run.Benchmarks, 1))
	highestAlloc := firstBenchmark(topByBytes(run.Benchmarks, 1))
	fmt.Fprintln(b, "| Metric | Value |")
	fmt.Fprintln(b, "| --- | ---: |")
	fmt.Fprintf(b, "| Benchmarks | %d |\n", len(run.Benchmarks))
	fmt.Fprintf(b, "| Packages | %d |\n", packages)
	fmt.Fprintf(b, "| Slowest | %s |\n", escapeTable(formatBenchmarkRef(slowest)))
	fmt.Fprintf(b, "| Highest B/op | %s |\n", escapeTable(formatBenchmarkRef(highestAlloc)))
}

func writeSummaryTableCN(b *strings.Builder, run benchmarkRun) {
	packages := packageCount(run.Benchmarks)
	slowest := firstBenchmark(topByNs(run.Benchmarks, 1))
	highestAlloc := firstBenchmark(topByBytes(run.Benchmarks, 1))
	fmt.Fprintln(b, "| 指标 | 值 |")
	fmt.Fprintln(b, "| --- | ---: |")
	fmt.Fprintf(b, "| Benchmark 数量 | %d |\n", len(run.Benchmarks))
	fmt.Fprintf(b, "| 包数量 | %d |\n", packages)
	fmt.Fprintf(b, "| 最慢项 | %s |\n", escapeTable(formatBenchmarkRef(slowest)))
	fmt.Fprintf(b, "| 最高 B/op | %s |\n", escapeTable(formatBenchmarkRef(highestAlloc)))
}

func writeBenchmarkTable(b *strings.Builder, benchmarks []benchmarkResult, comparisons map[string]comparison) {
	fmt.Fprintln(b, "| Package | Benchmark | Iterations | ns/op | B/op | allocs/op | Delta ns/op |")
	fmt.Fprintln(b, "| --- | --- | ---: | ---: | ---: | ---: | ---: |")
	for _, bench := range benchmarks {
		delta := "baseline"
		if cmp, ok := comparisons[benchmarkKey(bench)]; ok {
			delta = formatPercent(cmp.DeltaPct)
		}
		fmt.Fprintf(b, "| %s | %s | %d | %s | %s | %s | %s |\n",
			escapeTable(bench.Package),
			escapeTable(bench.Name),
			bench.Iterations,
			formatFloat(bench.NsPerOp),
			formatOptional(bench.BytesPerOp),
			formatOptional(bench.AllocsPerOp),
			delta,
		)
	}
}

func readPreviousBenchmarks(path string) map[string]benchmarkResult {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	previous := make(map[string]benchmarkResult)
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "| ") || strings.Contains(line, "---") || strings.Contains(line, "Package | Benchmark") {
			continue
		}
		cells := splitMarkdownRow(line)
		if len(cells) < 7 || !strings.HasPrefix(cells[1], "Benchmark") {
			continue
		}
		bench := benchmarkResult{
			Package:    unescapeTable(cells[0]),
			Name:       unescapeTable(cells[1]),
			Iterations: int64(parseFloat(cells[2])),
			NsPerOp:    parseFloat(cells[3]),
		}
		if cells[4] != "" {
			value := parseFloat(cells[4])
			bench.BytesPerOp = &value
		}
		if cells[5] != "" {
			value := parseFloat(cells[5])
			bench.AllocsPerOp = &value
		}
		previous[benchmarkKey(bench)] = bench
	}
	return previous
}

func splitMarkdownRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	cells := make([]string, 0, len(parts))
	for _, part := range parts {
		cells = append(cells, strings.TrimSpace(part))
	}
	return cells
}

func compareBenchmarks(current []benchmarkResult, previous map[string]benchmarkResult) map[string]comparison {
	comparisons := make(map[string]comparison)
	for _, bench := range current {
		prev, ok := previous[benchmarkKey(bench)]
		if !ok || prev.NsPerOp == 0 {
			continue
		}
		comparisons[benchmarkKey(bench)] = comparison{
			PreviousNs: prev.NsPerOp,
			DeltaPct:   ((bench.NsPerOp - prev.NsPerOp) / prev.NsPerOp) * 100,
		}
	}
	return comparisons
}

func pruneReports(dir string, retention int) ([]string, error) {
	if retention < 1 {
		retention = 1
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	dateFile := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}\.md$`)
	var reports []string
	for _, entry := range entries {
		if entry.IsDir() || !dateFile.MatchString(entry.Name()) {
			continue
		}
		reports = append(reports, entry.Name())
	}
	sort.Sort(sort.Reverse(sort.StringSlice(reports)))
	if len(reports) > retention {
		for _, old := range reports[retention:] {
			if err := os.Remove(filepath.Join(dir, old)); err != nil {
				return nil, err
			}
		}
		reports = reports[:retention]
	}
	return reports, nil
}

func writeDocsPages(src string, reports []string, run benchmarkRun) error {
	en := filepath.Join(src, "en", "performance-benchmark-reports.md")
	zh := filepath.Join(src, "zh", "performance-benchmark-reports.md")
	if err := os.MkdirAll(filepath.Dir(en), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(zh), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(en, []byte(renderDocsEN(reports, run)), 0o644); err != nil {
		return err
	}
	return os.WriteFile(zh, []byte(renderDocsZH(reports, run)), 0o644)
}

func renderDocsEN(reports []string, run benchmarkRun) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Performance Benchmark Reports")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "Weekly benchmark runs record ToughRADIUS performance signals from existing Go `Benchmark*` functions. Reports are informational and do not fail on timing drift from GitHub hosted runners.")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "**Latest verdict:** %s\n\n", strings.ToUpper(run.Verdict))
	fmt.Fprintln(&b, "## Latest Summary")
	fmt.Fprintln(&b)
	writeSummaryTable(&b, run)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Retained Reports")
	fmt.Fprintln(&b)
	writeReportLinks(&b, reports)
	return b.String()
}

func renderDocsZH(reports []string, run benchmarkRun) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# 性能基准测试报告")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "每周 benchmark 任务基于现有 Go `Benchmark*` 函数记录 ToughRADIUS 性能信号。报告仅用于观察趋势，不会因为 GitHub 托管 runner 的耗时波动直接失败。")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "**最近结论：** %s\n\n", chineseVerdict(run.Verdict))
	fmt.Fprintln(&b, "## 最近摘要")
	fmt.Fprintln(&b)
	writeSummaryTableCN(&b, run)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## 保留报告")
	fmt.Fprintln(&b)
	writeReportLinks(&b, reports)
	return b.String()
}

func writeReportLinks(b *strings.Builder, reports []string) {
	if len(reports) == 0 {
		fmt.Fprintln(b, "No scheduled reports have been generated yet.")
		return
	}
	for _, report := range reports {
		date := strings.TrimSuffix(report, ".md")
		fmt.Fprintf(b, "- [%s](https://github.com/talkincode/toughradius/blob/main/docs/reports/performance/%s)\n", date, report)
	}
}

func writeSummary(path, date string, run benchmarkRun) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	summary := reportSummary{
		Date:           date,
		Verdict:        run.Verdict,
		BenchmarkCount: len(run.Benchmarks),
		PackageCount:   packageCount(run.Benchmarks),
		Slowest:        firstBenchmark(topByNs(run.Benchmarks, 1)),
		HighestAlloc:   firstBenchmark(topByBytes(run.Benchmarks, 1)),
		WorkflowURL:    run.WorkflowURL,
		RefName:        run.RefName,
		CommitSHA:      run.CommitSHA,
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func topByNs(benchmarks []benchmarkResult, limit int) []benchmarkResult {
	return topBy(benchmarks, limit, func(a, b benchmarkResult) bool {
		return a.NsPerOp > b.NsPerOp
	})
}

func topByBytes(benchmarks []benchmarkResult, limit int) []benchmarkResult {
	return topBy(benchmarks, limit, func(a, b benchmarkResult) bool {
		return optionalValue(a.BytesPerOp) > optionalValue(b.BytesPerOp)
	})
}

func topBy(benchmarks []benchmarkResult, limit int, less func(a, b benchmarkResult) bool) []benchmarkResult {
	items := append([]benchmarkResult(nil), benchmarks...)
	sort.Slice(items, func(i, j int) bool {
		return less(items[i], items[j])
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func firstBenchmark(benchmarks []benchmarkResult) *benchmarkResult {
	if len(benchmarks) == 0 {
		return nil
	}
	bench := benchmarks[0]
	return &bench
}

func packageCount(benchmarks []benchmarkResult) int {
	seen := make(map[string]struct{})
	for _, bench := range benchmarks {
		seen[bench.Package] = struct{}{}
	}
	return len(seen)
}

func formatBenchmarkRef(bench *benchmarkResult) string {
	if bench == nil {
		return ""
	}
	return bench.Package + " / " + bench.Name
}

func benchmarkKey(bench benchmarkResult) string {
	return bench.Package + "\x00" + bench.Name
}

func stripBenchmarkSuffix(name string) string {
	return regexp.MustCompile(`-\d+$`).ReplaceAllString(name, "")
}

func workflowURL() string {
	server := envOr("GITHUB_SERVER_URL", "")
	repo := envOr("GITHUB_REPOSITORY", "")
	runID := envOr("GITHUB_RUN_ID", "")
	if server == "" || repo == "" || runID == "" {
		return ""
	}
	return server + "/" + repo + "/actions/runs/" + runID
}

func chineseVerdict(verdict string) string {
	switch verdict {
	case "recorded":
		return "已记录"
	default:
		return verdict
	}
}

func markdownLink(url string) string {
	if url == "" {
		return ""
	}
	return "[workflow run](" + url + ")"
}

func shortSHA(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}

func escapeTable(value string) string {
	value = strings.ReplaceAll(value, "\n", "<br>")
	return strings.ReplaceAll(value, "|", "\\|")
}

func unescapeTable(value string) string {
	return strings.ReplaceAll(value, "\\|", "|")
}

func formatFloat(value float64) string {
	if math.Abs(value) >= 1000 {
		return strconv.FormatFloat(value, 'f', 0, 64)
	}
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func formatOptional(value *float64) string {
	if value == nil {
		return ""
	}
	return formatFloat(*value)
}

func formatPercent(value float64) string {
	sign := "+"
	if value < 0 {
		sign = ""
	}
	return sign + strconv.FormatFloat(value, 'f', 1, 64) + "%"
}

func optionalValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func parseFloat(value string) float64 {
	value = strings.ReplaceAll(value, ",", "")
	parsed, _ := strconv.ParseFloat(value, 64)
	return parsed
}

func isInt(value string) bool {
	_, err := strconv.ParseInt(value, 10, 64)
	return err == nil
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
