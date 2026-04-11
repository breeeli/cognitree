import { useEffect, useState } from "react";
import { getHealth, type HealthStatus } from "@/api/health";

export function HomePage() {
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getHealth().then((res) => {
      if (res.ok) {
        setHealth(res.data);
      } else {
        setError(res.error ?? "Unknown error");
      }
      setLoading(false);
    });
  }, []);

  return (
    <div className="max-w-2xl">
      <h1 className="text-3xl font-bold mb-2">Cognitree</h1>
      <p className="text-muted-foreground mb-8">
        Thinking Tree IDE — 将提问过程沉淀为知识结构
      </p>

      <div className="rounded-xl border border-border bg-card p-6">
        <h2 className="text-sm font-medium text-muted-foreground mb-4 uppercase tracking-wider">
          System Status
        </h2>

        {loading && (
          <div className="flex items-center gap-2 text-muted-foreground">
            <span className="animate-pulse">●</span>
            Connecting...
          </div>
        )}

        {error && (
          <div className="flex items-center gap-3">
            <span className="w-3 h-3 rounded-full bg-danger" />
            <div>
              <p className="text-sm font-medium text-danger">Backend Offline</p>
              <p className="text-xs text-muted-foreground mt-1">{error}</p>
            </div>
          </div>
        )}

        {health && (
          <div className="space-y-3">
            <StatusRow
              label="Backend"
              status={health.status === "ok" ? "ok" : "error"}
            />
            <StatusRow
              label="Database"
              status={health.database === "ok" ? "ok" : "error"}
            />
            <p className="text-xs text-muted-foreground pt-2 border-t border-border">
              Last checked: {new Date(health.timestamp).toLocaleString()}
            </p>
          </div>
        )}
      </div>
    </div>
  );
}

function StatusRow({
  label,
  status,
}: {
  label: string;
  status: "ok" | "error";
}) {
  return (
    <div className="flex items-center justify-between">
      <span className="text-sm">{label}</span>
      <div className="flex items-center gap-2">
        <span
          className={`w-2.5 h-2.5 rounded-full ${
            status === "ok" ? "bg-success" : "bg-danger"
          }`}
        />
        <span
          className={`text-xs font-medium ${
            status === "ok" ? "text-success" : "text-danger"
          }`}
        >
          {status === "ok" ? "Connected" : "Unavailable"}
        </span>
      </div>
    </div>
  );
}
