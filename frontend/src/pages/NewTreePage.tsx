import { useState } from "react";
import { useNavigate } from "react-router-dom";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { streamCreateTreeFirstQuestion } from "@/api/trees";
import { Button } from "@/components/ui/button";
import { markdownComponents } from "@/components/workspace/markdownComponents";

const STORAGE_KEY_TREE = "cognitree:currentTreeId";

type FlowPhase = "idle" | "creating" | "streaming" | "error";

interface TreeSession {
  treeId: string;
  rootNodeId: string;
}

export function NewTreePage() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState("");
  const [answer, setAnswer] = useState("");
  const [session, setSession] = useState<TreeSession | null>(null);
  const [loading, setLoading] = useState(false);
  const [phase, setPhase] = useState<FlowPhase>("idle");
  const [error, setError] = useState<string | null>(null);

  const canSubmit = Boolean(
    !loading &&
      ((session === null && question.trim()) ||
        (session !== null && phase === "error")),
  );

  const handleSubmit = async () => {
    const trimmed = question.trim();
    if (!canSubmit || !trimmed) return;

    setLoading(true);
    setError(null);
    setAnswer("");
    setPhase("creating");

    let treeId = session?.treeId ?? "";
    let rootNodeId = session?.rootNodeId ?? "";

    try {
      await streamCreateTreeFirstQuestion(
        session
          ? {
              question: trimmed,
              tree_id: session.treeId,
              root_node_id: session.rootNodeId,
            }
          : { question: trimmed },
        {
          onTreeReady: (tree) => {
            treeId = tree.id;
            setSession((prev) =>
              prev ? { ...prev, treeId: tree.id } : { treeId: tree.id, rootNodeId: "" },
            );
          },
          onRootNodeReady: (rootNode) => {
            treeId = rootNode.tree_id;
            rootNodeId = rootNode.id;
            setSession({ treeId: rootNode.tree_id, rootNodeId: rootNode.id });
            setPhase("streaming");
          },
          onDelta: (delta) => {
            setAnswer((prev) => prev + delta);
          },
        },
      );

      if (!treeId || !rootNodeId) {
        throw new Error("思维树信息缺失");
      }

      localStorage.setItem(STORAGE_KEY_TREE, treeId);
      navigate("/", { replace: true });
    } catch (err) {
      setPhase("error");
      setError(err instanceof Error ? err.message : "创建失败，请稍后重试");
    } finally {
      setLoading(false);
    }
  };

  const statusText =
    phase === "creating"
      ? "正在创建思维树"
      : phase === "streaming"
        ? "正在生成首个回答"
        : phase === "error"
          ? "生成失败，保留当前状态以便重试"
          : session
            ? "思维树已创建"
            : "输入一个问题，系统会先创建树，再生成首答";

  const showAnswer = answer.trim().length > 0 || phase === "streaming" || phase === "error";

  return (
    <main className="relative flex-1 overflow-hidden bg-background">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top,_rgba(97,134,255,0.12),_transparent_38%),radial-gradient(circle_at_bottom_right,_rgba(97,134,255,0.08),_transparent_32%)]" />
      <div className="relative flex min-h-full items-center justify-center px-6 py-12">
        <section className="w-full max-w-4xl space-y-8">
          <div className="space-y-3">
            <p className="text-sm font-medium uppercase tracking-[0.22em] text-muted-foreground">
              Cognitree
            </p>
            <h1 className="text-4xl font-semibold tracking-tight sm:text-5xl">
              先问问题，再长出树
            </h1>
            <p className="max-w-2xl text-base leading-7 text-muted-foreground">
              输入你的第一个问题后，系统会先创建一棵新的思维树，再把首个回答流式返回给你。
            </p>
          </div>

          <div className="grid gap-6 lg:grid-cols-[1.05fr_0.95fr]">
            <div className="rounded-3xl border border-border bg-card/90 p-6 shadow-[0_20px_60px_-25px_rgba(0,0,0,0.18)] backdrop-blur-sm sm:p-8">
              <label className="mb-3 block text-sm font-medium text-foreground">
                你的第一个问题
              </label>
              <textarea
                className="min-h-[200px] w-full resize-none rounded-2xl border border-input bg-background px-4 py-4 text-sm leading-7 outline-none transition-colors placeholder:text-muted-foreground focus:border-primary focus:ring-2 focus:ring-primary/20 disabled:cursor-not-allowed disabled:opacity-70"
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
                disabled={loading || session !== null}
              />

              <div className="mt-4 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
                <p className="text-sm text-muted-foreground">{statusText}</p>
                <div className="flex gap-2">
                  <Button variant="outline" onClick={() => navigate("/")} disabled={loading}>
                    返回工作区
                  </Button>
                  <Button onClick={handleSubmit} disabled={!canSubmit}>
                    {loading
                      ? "处理中..."
                      : session && phase === "error"
                        ? "重试生成"
                        : "发送问题"}
                  </Button>
                </div>
              </div>

              {error && (
                <p className="mt-4 rounded-2xl border border-danger/20 bg-danger/5 px-4 py-3 text-sm text-danger">
                  {error}
                </p>
              )}
            </div>

            <div className="rounded-3xl border border-border bg-card/70 p-6 backdrop-blur-sm sm:p-8">
              <div className="flex items-center justify-between gap-3 border-b border-border pb-4">
                <div>
                  <p className="text-sm font-medium text-foreground">流式预览</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {phase === "streaming"
                      ? "回答正在逐步生成"
                      : session
                        ? "树已就绪，等待下一步"
                        : "这里会实时显示首个回答"}
                  </p>
                </div>
                <span className="rounded-full bg-primary/10 px-3 py-1 text-xs font-medium text-primary">
                  {session ? "已创建" : "待开始"}
                </span>
              </div>

              <div className="pt-4">
                <div className="flex justify-end mb-4">
                  <div className="max-w-[80%] rounded-2xl rounded-tr-sm bg-primary text-primary-foreground px-4 py-3">
                    <p className="text-sm leading-6 whitespace-pre-wrap">{question || " "}</p>
                  </div>
                </div>

                <div className="flex justify-start">
                  <div className="max-w-[90%] flex items-start gap-2">
                    <span className="shrink-0 w-7 h-7 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium mt-0.5">
                      AI
                    </span>
                    <div className="rounded-2xl rounded-tl-sm bg-card border border-border px-4 py-3 min-h-[120px] flex-1">
                      {showAnswer ? (
                        <div className="space-y-3 text-sm leading-7">
                          <ReactMarkdown remarkPlugins={[remarkGfm]} components={markdownComponents}>
                            {answer || "AI 正在整理回答..."}
                          </ReactMarkdown>
                        </div>
                      ) : (
                        <div className="text-sm text-muted-foreground">
                          <span className="inline-flex items-center gap-2">
                            <span className="h-2 w-2 rounded-full bg-primary/60 animate-pulse" />
                            AI 正在等待首问...
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>
  );
}
