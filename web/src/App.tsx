import React, { useEffect, useState, useCallback } from "react";
import { Navbar } from "./components/Navbar";
import { KpiGrid } from "./components/KpiGrid";
import { TenantTable } from "./components/TenantTable";
import { FeatureWaterfall } from "./components/FeatureWaterfall";
import { ModelDistribution } from "./components/ModelDistribution";
import { AlertsPanel } from "./components/AlertsPanel";
import { RecommendationOptimizer } from "./components/RecommendationOptimizer";
import { LiveInspector } from "./components/LiveInspector";
import { TenantModal } from "./components/TenantModal";
import { NewTenantModal } from "./components/NewTenantModal";
import { MarginSimulator } from "./components/MarginSimulator";
import { CacheRoiAnalytics } from "./components/CacheRoiAnalytics";
import {
  OverviewSummary,
  TenantMetric,
  FeatureMetric,
  ModelMetric,
  Alert,
  Recommendation,
  RequestLog,
} from "./types";

export const App: React.FC = () => {
  const [currentTab, setTab] = useState("overview");
  const [overview, setOverview] = useState<OverviewSummary | null>(null);
  const [tenants, setTenants] = useState<TenantMetric[]>([]);
  const [features, setFeatures] = useState<FeatureMetric[]>([]);
  const [models, setModels] = useState<ModelMetric[]>([]);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [recommendations, setRecommendations] = useState<Recommendation[]>([]);
  const [logs, setLogs] = useState<RequestLog[]>([]);

  const [selectedTenantId, setSelectedTenantId] = useState<string | null>(null);
  const [isNewTenantOpen, setIsNewTenantOpen] = useState(false);

  const fetchAll = useCallback(async () => {
    try {
      const [ovRes, tenRes, featRes, modRes, altRes, recRes, logRes] = await Promise.all([
        fetch("/api/overview").then((r) => r.json()),
        fetch("/api/tenants").then((r) => r.json()),
        fetch("/api/features").then((r) => r.json()),
        fetch("/api/models").then((r) => r.json()),
        fetch("/api/alerts?resolved=false").then((r) => r.json()),
        fetch("/api/recommendations").then((r) => r.json()),
        fetch("/api/logs?limit=50").then((r) => r.json()),
      ]);

      setOverview(ovRes);
      setTenants(Array.isArray(tenRes) ? tenRes : []);
      setFeatures(Array.isArray(featRes) ? featRes : []);
      setModels(Array.isArray(modRes) ? modRes : []);
      setAlerts(Array.isArray(altRes) ? altRes : []);
      setRecommendations(Array.isArray(recRes) ? recRes : []);
      setLogs(Array.isArray(logRes) ? logRes : []);
    } catch (err) {
      console.error("Failed to fetch finops data:", err);
    }
  }, []);

  useEffect(() => {
    fetchAll();
    const interval = setInterval(fetchAll, 5000); // 5s live polling
    return () => clearInterval(interval);
  }, [fetchAll]);

  const handleResolveAlert = async (id: string) => {
    await fetch(`/api/alerts/${id}/resolve`, { method: "POST" });
    fetchAll();
  };

  const handleApplyRecommendation = async (id: string) => {
    await fetch(`/api/recommendations/${id}/apply`, { method: "POST" });
    fetchAll();
  };

  const handleGenerateRecommendations = async () => {
    await fetch("/api/recommendations", { method: "POST" });
    fetchAll();
  };

  const handleCreateTenant = async (tenantData: {
    id: string;
    name: string;
    monthly_plan_usd: number;
    monthly_budget_usd?: number;
    webhook_url?: string;
  }) => {
    await fetch("/api/tenants", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(tenantData),
    });
    fetchAll();
  };

  const handleDeleteTenant = async (id: string) => {
    await fetch(`/api/tenants/${id}`, { method: "DELETE" });
    fetchAll();
  };

  const unprofitableCount = tenants.filter((t) => t.is_unprofitable).length;
  const activeAlertsCount = alerts.filter((a) => !a.resolved).length;

  return (
    <div className="min-h-screen bg-[#09090b] text-zinc-100 flex flex-col font-sans">
      <Navbar
        currentTab={currentTab}
        setTab={setTab}
        unprofitableCount={unprofitableCount}
        activeAlertsCount={activeAlertsCount}
        onRefresh={fetchAll}
        onOpenNewTenant={() => setIsNewTenantOpen(true)}
      />

      <main className="flex-1 max-w-7xl w-full mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        {/* KPI Row always present on top */}
        <KpiGrid
          summary={overview}
          onFilterUnprofitable={() => setTab("tenants")}
          onViewAlerts={() => setTab("alerts")}
        />

        {/* Tab Views */}
        {currentTab === "overview" && (
          <div className="space-y-6">
            <TenantTable
              tenants={tenants}
              onSelectTenant={(id) => setSelectedTenantId(id)}
            />
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <FeatureWaterfall features={features} />
              <ModelDistribution models={models} />
            </div>
            <RecommendationOptimizer
              recommendations={recommendations}
              onApply={handleApplyRecommendation}
              onGenerate={handleGenerateRecommendations}
            />
          </div>
        )}

        {currentTab === "tenants" && (
          <TenantTable
            tenants={tenants}
            onSelectTenant={(id) => setSelectedTenantId(id)}
            unprofitableOnly={false}
          />
        )}

        {currentTab === "features" && (
          <div className="space-y-6">
            <FeatureWaterfall features={features} />
            <ModelDistribution models={models} />
          </div>
        )}

        {currentTab === "simulator" && (
          <div className="space-y-6">
            <MarginSimulator />
          </div>
        )}

        {currentTab === "cache" && (
          <div className="space-y-6">
            <CacheRoiAnalytics />
          </div>
        )}

        {currentTab === "recommendations" && (
          <RecommendationOptimizer
            recommendations={recommendations}
            onApply={handleApplyRecommendation}
            onGenerate={handleGenerateRecommendations}
          />
        )}

        {currentTab === "alerts" && (
          <div className="space-y-6">
            <AlertsPanel alerts={alerts} onResolve={handleResolveAlert} />
          </div>
        )}

        {currentTab === "inspector" && (
          <LiveInspector logs={logs} onRefresh={fetchAll} />
        )}
      </main>

      {/* Modals */}
      <TenantModal
        tenantId={selectedTenantId}
        onClose={() => setSelectedTenantId(null)}
        onDeleteTenant={handleDeleteTenant}
      />

      <NewTenantModal
        isOpen={isNewTenantOpen}
        onClose={() => setIsNewTenantOpen(false)}
        onCreate={handleCreateTenant}
      />
    </div>
  );
};
