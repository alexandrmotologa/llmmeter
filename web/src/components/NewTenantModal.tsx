import React, { useState } from "react";
import { X, PlusCircle } from "lucide-react";

interface NewTenantModalProps {
  isOpen: boolean;
  onClose: () => void;
  onCreate: (tenant: {
    id: string;
    name: string;
    monthly_plan_usd: number;
    monthly_budget_usd?: number;
    webhook_url?: string;
  }) => Promise<void>;
}

export const NewTenantModal: React.FC<NewTenantModalProps> = ({ isOpen, onClose, onCreate }) => {
  const [name, setName] = useState("");
  const [id, setId] = useState("");
  const [planUSD, setPlanUSD] = useState("49.00");
  const [budgetUSD, setBudgetUSD] = useState("100.00");
  const [webhookURL, setWebhookURL] = useState("");
  const [loading, setLoading] = useState(false);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) return;

    setLoading(true);
    try {
      const slug = id.trim() || name.toLowerCase().replace(/[^a-z0-9]+/g, "-");
      await onCreate({
        id: slug,
        name: name.trim(),
        monthly_plan_usd: parseFloat(planUSD) || 0,
        monthly_budget_usd: budgetUSD ? parseFloat(budgetUSD) : undefined,
        webhook_url: webhookURL.trim() || undefined,
      });
      onClose();
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-surface-1 border border-surface-border rounded-xl max-w-md w-full shadow-2xl overflow-hidden">
        <div className="p-4 border-b border-surface-border flex items-center justify-between">
          <div>
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <PlusCircle className="w-4 h-4 text-emerald-400" />
              <span>Register New SaaS Tenant</span>
            </h2>
            <p className="text-xs text-zinc-400 mt-0.5">
              Add a customer account for real-time cost attribution &amp; margin tracking
            </p>
          </div>
          <button onClick={onClose} className="p-1 rounded text-zinc-400 hover:text-white">
            <X className="w-4 h-4" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 space-y-3.5 text-xs">
          <div>
            <label className="block text-zinc-300 font-medium mb-1">Company / Customer Name</label>
            <input
              type="text"
              required
              placeholder="e.g. Acme Corp"
              value={name}
              onChange={(e) => {
                setName(e.target.value);
                if (!id) {
                  setId(e.target.value.toLowerCase().replace(/[^a-z0-9]+/g, "-"));
                }
              }}
              className="w-full px-3 py-2 bg-surface-2 border border-surface-border rounded-md text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 font-sans"
            />
          </div>

          <div>
            <label className="block text-zinc-300 font-medium mb-1">
              Tenant ID / Slug <span className="text-zinc-500">(used in X-Tenant-ID header)</span>
            </label>
            <input
              type="text"
              required
              placeholder="e.g. acme-corp"
              value={id}
              onChange={(e) => setId(e.target.value)}
              className="w-full px-3 py-2 bg-surface-2 border border-surface-border rounded-md text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 font-mono"
            />
          </div>

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="block text-zinc-300 font-medium mb-1">Monthly Plan ($/mo)</label>
              <input
                type="number"
                step="0.01"
                min="0"
                value={planUSD}
                onChange={(e) => setPlanUSD(e.target.value)}
                className="w-full px-3 py-2 bg-surface-2 border border-surface-border rounded-md text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 font-mono"
              />
            </div>
            <div>
              <label className="block text-zinc-300 font-medium mb-1">Budget Cap ($/mo)</label>
              <input
                type="number"
                step="0.01"
                min="0"
                placeholder="Optional cap"
                value={budgetUSD}
                onChange={(e) => setBudgetUSD(e.target.value)}
                className="w-full px-3 py-2 bg-surface-2 border border-surface-border rounded-md text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 font-mono"
              />
            </div>
          </div>

          <div>
            <label className="block text-zinc-300 font-medium mb-1">
              Alert Webhook URL <span className="text-zinc-500">(Slack, Zapier, Webhook)</span>
            </label>
            <input
              type="url"
              placeholder="https://hooks.slack.com/services/..."
              value={webhookURL}
              onChange={(e) => setWebhookURL(e.target.value)}
              className="w-full px-3 py-2 bg-surface-2 border border-surface-border rounded-md text-white placeholder-zinc-500 focus:outline-none focus:border-emerald-500 font-mono"
            />
          </div>

          <div className="pt-3 border-t border-surface-border flex items-center justify-end gap-2">
            <button
              type="button"
              onClick={onClose}
              className="px-3 py-1.5 rounded-md bg-surface-2 hover:bg-surface-3 text-zinc-300 font-medium transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={loading}
              className="px-4 py-1.5 rounded-md bg-emerald-600 hover:bg-emerald-500 text-white font-medium transition-colors disabled:opacity-50"
            >
              {loading ? "Creating..." : "Save Tenant"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
