import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { createTree } from "@/api/trees";
import { chatOnNode } from "@/api/chat";
import { Button } from "@/components/ui/button";

const STORAGE_KEY_TREE = "cognitree:currentTreeId";

export function NewTreePage() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [stage, setStage] = useState<"idle" | "creating" | "answering">("idle");
  const [createdTree, setCreatedTree] = useState<{
    treeId: string;
    rootNodeId: string;
  } | null>(null);

  const handleSubmit = async () => {
    const trimmed = question.trim();
    if (!trimmed || loading) return;

    setLoading(true);
    setError(null);
    setStage("creating");

    try {
      let treeId: string;
      let rootNodeId: string;

      if (createdTree) {
        treeId = createdTree.treeId;
        rootNodeId = createdTree.rootNodeId;
      } else {
        const treeRes = await createTree(trimmed);
        if (!treeRes.ok) {
          throw new Error(treeRes.error ?? "创建思维树失败");
        }

        treeId = treeRes.data.tree.id;
        rootNodeId = treeRes.data.root_node.id;
        setCreatedTree({ treeId, rootNodeId });
      }

      setStage("answering");
      const chatRes = await chatOnNode(rootNodeId, trimmed);
      if (!chatRes.ok) {
        throw new Error(chatRes.error ?? "生成首个回答失败");
      }

      localStorage.setItem(STORAGE_KEY_TREE, treeId);
      navigate("/", { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : "创建失败，请稍后重试");
      setStage("idle");
    } finally {
      setLoading(false);
    }
  };

  return (
    <main className="relative flex-1 overflow-hidden bg-background">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(97,134,255,0.12),_transparent_38%),radial-gradient(circle_at_bottom_right,_rgba(97,134,255,0.08),_transparent_32%)]" />
      <div className="relative flex min-h-full items-center justify-center px-6 py-12">
        <section className="w-full max-w-3xl">
          <div className="mb-8 space-y-3">
            <p className="text-sm font-medium uppercase tracking-[0.22em] text-muted-foreground">
              Cognitree
            </p>
            <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">
              先问问题，再长出树
            </h1>
            <p className="max-w-2xl text-base leading-7 text-muted-foreground">
              输入你的第一个问题，系统会先创建一棵新的思维树，再为它生成首个回答。
            </p>
          </div>

          <div className="rounded-3xl border border-border bg-card/90 p-6 shadow-[0_20px_60px_-25px_rgba(0,0,0,0.18)] backdrop-blur-sm sm:p-8">
            <label className="mb-3 block text-sm font-medium text-foreground">
              你的第一个问题
            </label>
            <textarea
              className="min-h-[180px] w-full resize-none rounded-2xl border border-input bg-background px-4 py-4 text-sm leading-7 outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20"
              placeholder="例如：为什么分布式系统这么复杂？"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  void handleSubmit();
                }
              }}
              autoFocus
              disabled={loading || createdTree !== null}
            />

            <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <p className="text-sm text-muted-foreground">
                Enter 发送，Shift+Enter 换行
                {stage === "creating" ? " · 正在创建思维树" : ""}
                {stage === "answering" ? " · 正在生成首个回答" : ""}
                {createdTree && stage === "idle" ? " · 思维树已创建，点击重试继续生成回答" : ""}
              </p>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={() => navigate("/")}
                  disabled={loading}
                >
                  返回工作区
                </Button>
                <Button onClick={handleSubmit} disabled={!question.trim() || loading}>
                  {loading ? "处理中..." : createdTree ? "重试生成" : "发送问题"}
                </Button>
              </div>
            </div>

            {error && (
              <p className="mt-4 rounded-2xl border border-danger/20 bg-danger/5 px-4 py-3 text-sm text-danger">
                {error}
              </p>
            )}
          </div>
        </section>
      </div>
    </main>
  );
}
