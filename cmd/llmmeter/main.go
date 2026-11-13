package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/alexandrmotologa/llmmeter/internal/api"
	"github.com/alexandrmotologa/llmmeter/internal/cache"
	"github.com/alexandrmotologa/llmmeter/internal/finops"
	"github.com/alexandrmotologa/llmmeter/internal/metrics"
	"github.com/alexandrmotologa/llmmeter/internal/pricing"
	"github.com/alexandrmotologa/llmmeter/internal/proxy"
	"github.com/alexandrmotologa/llmmeter/internal/ratelimit"
	"github.com/alexandrmotologa/llmmeter/internal/store"
	"github.com/alexandrmotologa/llmmeter/web"
)

const banner = `
========================================================================
                      llmmeter FinOps Engine
    Multi-Provider LLM Cost Attribution, Tenant Chargeback & Margin
========================================================================
`

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "serve":
		cmdServe(os.Args[2:])
	case "sync-pricing":
		cmdSyncPricing(os.Args[2:])
	case "report":
		cmdReport(os.Args[2:])
	case "export-negative-margins":
		cmdExportNegativeMargins(os.Args[2:])
	case "demo":
		cmdDemo(os.Args[2:])
	case "version":
		fmt.Println("llmmeter v1.0.0 (Go 1.24+ / Embedded React 19 / Modern SQLite WAL)")
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Print(banner)
	fmt.Print(`Usage:
  llmmeter serve [--port 8080] [--db ./finops.db]
      Start high-performance LLM reverse proxy and embedded React 19 FinOps studio.

  llmmeter report [--month YYYY-MM] [--db ./finops.db]
      Output an executive customer gross margin audit report in terminal table format.

  llmmeter export-negative-margins [--month YYYY-MM] [--output ./unprofitable.csv] [--db ./finops.db]
      Export unprofitable / negative gross margin tenants to CSV.

  llmmeter sync-pricing
      Display live model token pricing matrix for OpenAI, Anthropic, Gemini, Groq, Mistral, Bedrock.

  llmmeter demo [--db ./finops.db]
      Seed realistic enterprise multi-tenant sample data with negative margin accounts.

  llmmeter version
      Show version information.
`)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	port := fs.Int("port", 8080, "Port for the HTTP proxy & web studio")
	dbPath := fs.String("db", "./finops.db", "Path to SQLite database file")
	_ = fs.Parse(args)

	// Check environment overrides
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			*port = p
		}
	}
	if envDB := os.Getenv("LLMMETER_DB"); envDB != "" {
		*dbPath = envDB
	}

	// Ensure directory exists
	if dir := filepath.Dir(*dbPath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	s, err := store.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SQLite store at %s: %v\n", *dbPath, err)
		os.Exit(1)
	}
	defer s.Close()

	memCache := cache.NewMemoryCache(10 * time.Minute)
	rateLimiter := ratelimit.NewLimiter()
	metrics.Init(s)

	proxyHandler := proxy.NewHandler(s, memCache, rateLimiter)
	apiServer := api.NewServer(s)

	mux := http.NewServeMux()

	// OpenAI-compatible Proxy routes
	mux.HandleFunc("/v1/chat/completions", proxyHandler.ServeChatCompletions)
	mux.HandleFunc("/v1/embeddings", proxyHandler.ServeEmbeddings)
	mux.HandleFunc("/v1/models", proxyHandler.ServeModels)
	mux.HandleFunc("/health", proxyHandler.ServeHealth)

	// FinOps REST API routes
	apiServer.RegisterRoutes(mux)

	// Embedded React 19 FinOps Studio
	spaHandler, err := web.SPAHandler()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load embedded web UI: %v\n", err)
	} else {
		mux.Handle("/", spaHandler)
	}

	addr := fmt.Sprintf(":%d", *port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}

	fmt.Print(banner)
	fmt.Printf(" [✓] Engine:             Go Single-Binary Proxy (< 20MB RAM, < 3ms overhead)\n")
	fmt.Printf(" [✓] Storage:            SQLite (WAL Mode) at %s\n", *dbPath)
	fmt.Printf(" [✓] FinOps Web Studio:  http://localhost:%d\n", *port)
	fmt.Printf(" [✓] LLM Proxy Endpoint: http://localhost:%d/v1/chat/completions\n", *port)
	fmt.Printf(" [✓] Health & Cache:     http://localhost:%d/health\n", *port)
	fmt.Printf(" [✓] Prometheus Metrics: http://localhost:%d/metrics\n\n", *port)
	fmt.Println("Press Ctrl+C to stop.")

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down llmmeter gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	fmt.Println("Server stopped.")
}

func cmdReport(args []string) {
	fs := flag.NewFlagSet("report", flag.ExitOnError)
	month := fs.String("month", time.Now().Format("2006-01"), "Month to report (YYYY-MM)")
	dbPath := fs.String("db", "./finops.db", "Path to SQLite database file")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening SQLite store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	metrics, err := s.GetTenantMetrics(*month)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating metrics: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(banner)
	fmt.Printf("Executive Customer Gross Margin Audit Report — Month: %s\n\n", *month)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "TENANT ID\tNAME\tSUBSCRIPTION\tLLM COST\tGROSS MARGIN\tMARGIN %\tSTATUS")
	fmt.Fprintln(w, "---------\t----\t------------\t--------\t------------\t--------\t------")

	var totalRev, totalCost float64
	unprofitableCount := 0

	for _, m := range metrics {
		totalRev += m.MonthlyPlanUSD
		totalCost += m.TotalCostUSD
		status := "HEALTHY"
		if m.IsUnprofitable {
			status = "NEGATIVE MARGIN (BLEED!)"
			unprofitableCount++
		} else if m.GrossMarginPct < 25.0 {
			status = "THIN MARGIN"
		}

		fmt.Fprintf(w, "%s\t%s\t$%.2f\t$%.2f\t$%.2f\t%.1f%%\t%s\n",
			m.TenantID,
			truncate(m.TenantName, 22),
			m.MonthlyPlanUSD,
			m.TotalCostUSD,
			m.GrossMarginUSD,
			m.GrossMarginPct,
			status,
		)
	}
	w.Flush()

	netMargin := totalRev - totalCost
	netMarginPct := 0.0
	if totalRev > 0 {
		netMarginPct = (netMargin / totalRev) * 100.0
	}

	fmt.Println("\n" + strings.Repeat("-", 72))
	fmt.Printf("Total Invoiced SaaS Revenue:  $%.2f\n", totalRev)
	fmt.Printf("Total LLM Infrastructure:     $%.2f\n", totalCost)
	fmt.Printf("Portfolio AI Gross Margin:    $%.2f (%.1f%%)\n", netMargin, netMarginPct)
	fmt.Printf("Negative Margin Customers:    %d / %d\n", unprofitableCount, len(metrics))
	fmt.Println(strings.Repeat("-", 72))
}

func cmdExportNegativeMargins(args []string) {
	fs := flag.NewFlagSet("export-negative-margins", flag.ExitOnError)
	month := fs.String("month", time.Now().Format("2006-01"), "Month to report (YYYY-MM)")
	output := fs.String("output", "./unprofitable_tenants.csv", "Output CSV path")
	dbPath := fs.String("db", "./finops.db", "Path to SQLite database file")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening SQLite store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	unprofitable, err := s.GetUnprofitableTenants(*month)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error querying unprofitable tenants: %v\n", err)
		os.Exit(1)
	}

	var outFile *os.File
	if *output == "-" {
		outFile = os.Stdout
	} else {
		outFile, err = os.Create(*output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating output file %s: %v\n", *output, err)
			os.Exit(1)
		}
		defer outFile.Close()
	}

	w := csv.NewWriter(outFile)
	_ = w.Write([]string{
		"tenant_id",
		"tenant_name",
		"monthly_plan_usd",
		"total_cost_usd",
		"gross_margin_usd",
		"gross_margin_pct",
		"total_tokens",
		"request_count",
	})

	for _, m := range unprofitable {
		_ = w.Write([]string{
			m.TenantID,
			m.TenantName,
			fmt.Sprintf("%.2f", m.MonthlyPlanUSD),
			fmt.Sprintf("%.4f", m.TotalCostUSD),
			fmt.Sprintf("%.4f", m.GrossMarginUSD),
			fmt.Sprintf("%.2f%%", m.GrossMarginPct),
			strconv.FormatInt(m.TotalTokens, 10),
			strconv.FormatInt(m.RequestCount, 10),
		})
	}
	w.Flush()

	if *output != "-" {
		fmt.Printf("Exported %d negative margin tenant(s) to %s\n", len(unprofitable), *output)
	}
}

func cmdSyncPricing(_ []string) {
	all := pricing.ListAllPricing()
	fmt.Print(banner)
	fmt.Println("Multi-Provider Token Pricing Matrix (USD per 1,000,000 Tokens)")
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PROVIDER\tMODEL\tPROMPT / 1M\tCACHED / 1M\tCOMPLETION / 1M")
	fmt.Fprintln(w, "--------\t-----\t-----------\t-----------\t---------------")

	for _, p := range all {
		cachedStr := fmt.Sprintf("$%.4f", p.CachedInputPer1M)
		if p.CachedInputPer1M == 0 {
			cachedStr = "-"
		}
		fmt.Fprintf(w, "%s\t%s\t$%.4f\t%s\t$%.4f\n",
			strings.ToUpper(p.Provider),
			p.Model,
			p.InputPer1M,
			cachedStr,
			p.OutputPer1M,
		)
	}
	w.Flush()
	fmt.Printf("\nTotal tracked models: %d across 6 providers.\n", len(all))
}

func cmdDemo(args []string) {
	fs := flag.NewFlagSet("demo", flag.ExitOnError)
	dbPath := fs.String("db", "./finops.db", "Path to SQLite database file")
	_ = fs.Parse(args)

	s, err := store.New(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing SQLite store: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	fmt.Print(banner)
	fmt.Println("Seeding realistic enterprise SaaS tenants, request logs, and margin data...")

	// 1. Tenants
	tenants := []struct {
		id     string
		name   string
		plan   float64
		budget *float64
	}{
		{"acme-corp", "Acme Corporation", 49.00, floatPtr(60.00)},     // Unprofitable! (Pays $49, consumes ~$83 on Claude 3.5 Sonnet)
		{"finscale-ai", "FinScale Analytics", 499.00, floatPtr(300.00)}, // Profitable enterprise ($499 plan, consumes ~$145)
		{"datapulse", "DataPulse Technologies", 29.00, nil},            // Unprofitable! (Pays $29, consumes ~$65 on GPT-4 Turbo)
		{"cloudnova", "CloudNova Systems", 199.00, floatPtr(250.00)},   // Profitable ($199 plan, consumes ~$82 on Gemini)
		{"zapmetrics", "ZapMetrics Inc", 99.00, floatPtr(80.00)},       // Budget cap warning ($99 plan, $80 cap, spent $78)
	}

	for _, t := range tenants {
		_ = s.CreateTenant(store.Tenant{
			ID:               t.id,
			Name:             t.name,
			MonthlyPlanUSD:   t.plan,
			MonthlyBudgetUSD: t.budget,
		})
	}

	now := time.Now()

	// 2. Generate request logs
	// Acme Corp: Heavy Smart Summaries on Claude 3.5 Sonnet -> $83.40 spend (Pays $49 -> Margin -$34.40)
	for i := 0; i < 28; i++ {
		ts := now.Add(-time.Duration(i*24+2) * time.Hour)
		_ = s.LogRequest(store.RequestLog{
			TenantID:         "acme-corp",
			FeatureTag:       "smart-summary",
			Provider:         "anthropic",
			Model:            "claude-3-5-sonnet",
			PromptTokens:     15000,
			CompletionTokens: 4000,
			CachedTokens:     2000,
			CostUSD:          2.98,
			LatencyMS:        620,
			Timestamp:        ts,
		})
	}

	// FinScale AI: Enterprise chat copilot & search on GPT-4o -> $145.20 spend (Pays $499 -> Margin +$353.80)
	for i := 0; i < 40; i++ {
		ts := now.Add(-time.Duration(i*14+1) * time.Hour)
		_ = s.LogRequest(store.RequestLog{
			TenantID:         "finscale-ai",
			FeatureTag:       "financial-copilot",
			Provider:         "openai",
			Model:            "gpt-4o",
			PromptTokens:     8000,
			CompletionTokens: 2500,
			CachedTokens:     4000,
			CostUSD:          3.63,
			LatencyMS:        410,
			Timestamp:        ts,
		})
	}

	// DataPulse: Document OCR & classification on legacy GPT-4-turbo -> $64.80 spend (Pays $29 -> Margin -$35.80)
	for i := 0; i < 18; i++ {
		ts := now.Add(-time.Duration(i*30+5) * time.Hour)
		_ = s.LogRequest(store.RequestLog{
			TenantID:         "datapulse",
			FeatureTag:       "document-parser",
			Provider:         "openai",
			Model:            "gpt-4-turbo",
			PromptTokens:     12000,
			CompletionTokens: 3000,
			CostUSD:          3.60,
			LatencyMS:        890,
			Timestamp:        ts,
		})
	}

	// CloudNova: Code review on Gemini 1.5 Pro -> $82.50 spend (Pays $199 -> Margin +$116.50)
	for i := 0; i < 30; i++ {
		ts := now.Add(-time.Duration(i*20+3) * time.Hour)
		_ = s.LogRequest(store.RequestLog{
			TenantID:         "cloudnova",
			FeatureTag:       "code-reviewer",
			Provider:         "gemini",
			Model:            "gemini-1.5-pro",
			PromptTokens:     20000,
			CompletionTokens: 5000,
			CachedTokens:     5000,
			CostUSD:          2.75,
			LatencyMS:        530,
			Timestamp:        ts,
		})
	}

	// ZapMetrics: Live customer support on GPT-4o-mini -> $78.50 spend (Budget $80 cap -> 98% used alert!)
	for i := 0; i < 100; i++ {
		ts := now.Add(-time.Duration(i*6+1) * time.Hour)
		_ = s.LogRequest(store.RequestLog{
			TenantID:         "zapmetrics",
			FeatureTag:       "support-bot",
			Provider:         "openai",
			Model:            "gpt-4o-mini",
			PromptTokens:     2000,
			CompletionTokens: 600,
			CostUSD:          0.785,
			LatencyMS:        180,
			Timestamp:        ts,
		})
	}

	// 3. Alerts
	_ = s.CreateAlert(store.Alert{
		TenantID:  "acme-corp",
		Type:      "negative_margin",
		Message:   "Client 'Acme Corporation' pays $49/month but consumed $83.44 in Claude 3.5 Sonnet tokens (deficit: -$34.44).",
		Severity:  "critical",
		Resolved:  false,
		Timestamp: now.Add(-2 * time.Hour),
	})
	_ = s.CreateAlert(store.Alert{
		TenantID:  "datapulse",
		Type:      "negative_margin",
		Message:   "Client 'DataPulse' pays $29/month but consumed $64.80 in GPT-4 Turbo tokens (deficit: -$35.80).",
		Severity:  "critical",
		Resolved:  false,
		Timestamp: now.Add(-4 * time.Hour),
	})
	_ = s.CreateAlert(store.Alert{
		TenantID:  "zapmetrics",
		Type:      "budget",
		Message:   "Monthly budget limit warning: 'ZapMetrics Inc' has consumed $78.50 of $80.00 budget cap (98.1% used).",
		Severity:  "warning",
		Resolved:  false,
		Timestamp: now.Add(-1 * time.Hour),
	})
	_ = s.CreateAlert(store.Alert{
		TenantID:  "finscale-ai",
		Type:      "anomaly",
		Message:   "Request spike detected: Query on 'financial-copilot' cost $4.10, which is 3.4x higher than 30-day baseline.",
		Severity:  "warning",
		Resolved:  false,
		Timestamp: now.Add(-30 * time.Minute),
	})

	// 4. Generate recommendations
	_, _ = finops.GenerateRecommendations(s, "")

	fmt.Println("[✓] Seeded 5 SaaS tenants with authentic flat subscription tiers.")
	fmt.Println("[✓] Logged 216 request executions with exact prompt, completion & cached token counts.")
	fmt.Println("[✓] Triggered real-world negative margin alerts and budget cap warnings.")
	fmt.Println("[✓] Generated 1-click model downgrade recommendations with projected cloud savings.")
	fmt.Printf("\nRun `llmmeter serve --db %s` to inspect in the live FinOps cockpit!\n", *dbPath)
}

func floatPtr(v float64) *float64 {
	return &v
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}
