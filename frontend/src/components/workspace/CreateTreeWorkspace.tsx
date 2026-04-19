import { useState } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Tree } from "@/types/tree";
import { streamCreateTreeFirstQuestion } from "@/api/trees";
import { Button } from "@/components/ui/button";
import { markdownComponents } from "./markdownComponents";

type FlowPhase = "idle" | "creating" | "streaming" | "error";

interface TreeSession {
  treeId: string;
  rootNodeId: string;
  tree: Tree;
}

interface CreateTreeWorkspaceProps {
  onCreated: (session: TreeSession) => void;
}

export function CreateTreeWorkspace({
  onCreated,
}: CreateTreeWorkspaceProps) {
  const [question, setQuestion] = useState("");
  const [submittedQuestion, setSubmittedQuestion] = useState("");
  const [answer, setAnswer] = useState("");
  const [session, setSession] = useState<TreeSession | null>(null);
  const [loading, setLoading] = useState(false);
  const [phase, setPhase] = useState<FlowPhase>("idle");
  const [error, setError] = useState<string | null>(null);

  const canSubmit = Boolean(
    !loading &&
      ((session === null && question.trim()) ||
        (session !== null && phase === "error" && submittedQuestion.trim())),
  );

  const handleSubmit = async () => {
    const trimmed = question.trim() || submittedQuestion.trim();
    if (!canSubmit || !trimmed) return;

    setLoading(true);
    setError(null);
    setAnswer("");
    setSubmittedQuestion(trimmed);
    setPhase("creating");

    let treeId = session?.treeId ?? "";
    let rootNodeId = session?.rootNodeId ?? "";
    let tree: Tree | null = session?.tree ?? null;

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
          onTreeReady: (readyTree) => {
            tree = readyTree;
            treeId = readyTree.id;
            setSession((prev) =>
              prev
                ? { ...prev, tree: readyTree, treeId: readyTree.id }
                : {
                    tree: readyTree,
                    treeId: readyTree.id,
                    rootNodeId: "",
                  },
            );
          },
          onRootNodeReady: (rootNode) => {
            treeId = rootNode.tree_id;
            rootNodeId = rootNode.id;
            setSession((prev) =>
              prev
                ? { ...prev, treeId: rootNode.tree_id, rootNodeId: rootNode.id }
                : {
                    tree: tree ?? {
                      id: rootNode.tree_id,
                      title: "",
                      description: "",
                      created_at: "",
                      updated_at: "",
                    },
                    treeId: rootNode.tree_id,
                    rootNodeId: rootNode.id,
                  },
            );
            setPhase("streaming");
          },
          onDelta: (delta) => {
            setAnswer((prev) => prev + delta);
          },
        },
      );

      if (!treeId || !rootNodeId || !tree) {
        throw new Error("思维树信息缺失");
      }

      setQuestion("");
      onCreated({ treeId, rootNodeId, tree });
    } catch (err) {
      setPhase("error");
      setError(err instanceof Error ? err.message : "创建失败，请稍后重试");
    } finally {
      setLoading(false);
    }
  };

  const showConversation =
    submittedQuestion.trim().length > 0 ||
    answer.trim().length > 0 ||
    phase === "streaming" ||
    phase === "error";

  return (
    <div className="flex h-full flex-col bg-background">
      <div className="flex-1 overflow-auto">
        <div className="mx-auto max-w-3xl px-8 py-6">
          {showConversation ? (
            <div className="space-y-6">
              <div className="flex justify-end">
                <div className="flex max-w-[75%] items-start gap-2">
                  <div className="rounded-2xl rounded-tr-sm bg-primary px-4 py-2.5 text-primary-foreground">
                    <p className="text-sm">{submittedQuestion}</p>
                  </div>
                  <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary/10 text-xs font-medium text-primary">
                    U
                  </span>
                </div>
              </div>

              <div className="flex justify-start">
                <div className="flex max-w-[85%] items-start gap-2">
                  <span className="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-muted text-xs font-medium text-muted-foreground">
                    AI
                  </span>
                  <div className="rounded-2xl rounded-tl-sm border border-border bg-card px-4 py-3">
                    <div className="space-y-3 text-sm leading-7">
                      {answer.trim().length > 0 ? (
                        <ReactMarkdown
                          remarkPlugins={[remarkGfm]}
                          components={markdownComponents}
                        >
                          {answer}
                        </ReactMarkdown>
                      ) : phase === "error" ? (
                        <div className="text-danger">创建失败，请稍后重试。</div>
                      ) : (
                        <div className="text-muted-foreground">
                          <span className="inline-flex items-center gap-2">
                            <span className="h-2 w-2 rounded-full bg-primary/60 animate-pulse" />
                            AI 正在生成首个回答...
                          </span>
                        </div>
                      )}
                      {loading && answer.trim().length > 0 && (
                        <div className="text-xs text-muted-foreground">
                          <span className="inline-flex items-center gap-2">
                            <span className="h-2 w-2 rounded-full bg-primary/60 animate-pulse" />
                            AI 正在生成...
                          </span>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>

              {error && (
                <div className="rounded-2xl border border-danger/20 bg-danger/5 px-4 py-3 text-sm text-danger">
                  {error}
                </div>
              )}
            </div>
          ) : (
            <div className="py-12 text-center text-muted-foreground">
              <p>还没有问答记录</p>
              <p className="mt-1 text-sm">在下方输入框中提问开始探索</p>
            </div>
          )}
        </div>
      </div>

      <div className="border-t border-border">
        <div className="mx-auto max-w-3xl px-8 py-4">
          <div className="flex items-end gap-3">
            <textarea
              className="min-h-[80px] max-h-[160px] flex-1 resize-none rounded-xl border border-input bg-background px-4 py-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
              placeholder="输入一个问题，系统会先创建树，再生成首答"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && !e.shiftKey) {
                  e.preventDefault();
                  void handleSubmit();
                }
              }}
              disabled={loading}
            />
            <Button
              className="h-10"
              onClick={() => void handleSubmit()}
              disabled={!canSubmit}
            >
              {loading ? "思考中..." : "发送"}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
