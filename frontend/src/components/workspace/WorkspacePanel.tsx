import { useState, useEffect, useCallback, useRef } from "react";
import type { Block, Node, QAPair } from "@/types/tree";
import { AnswerView } from "./AnswerView";
import { QuestionInput } from "./QuestionInput";
import { QuotePopover } from "./QuotePopover";
import { useTextSelection } from "@/hooks/useTextSelection";
import { streamChatOnNode } from "@/api/chat";
import { getNode, createAnchor } from "@/api/nodes";

interface PendingAutoSubmit {
  nodeId: string;
  question: string;
}

interface WorkspacePanelProps {
  node: Node | null;
  onNodeUpdated: (targetNodeId?: string) => Promise<void> | void;
  onNavigateToNode?: (nodeId: string, nextNode?: Node) => void;
  onRequestAutoSubmit?: (payload: PendingAutoSubmit) => void;
  pendingAutoSubmit?: PendingAutoSubmit | null;
  onPendingAutoSubmitConsumed?: () => void;
}

type SubmitStatus = "pending" | "complete" | "error";

function createPendingPair(question: string): QAPair & { status: SubmitStatus } {
  const blockId = crypto.randomUUID();
  return {
    id: crypto.randomUUID(),
    question,
    blocks: [
      {
        id: blockId,
        type: "paragraph" as Block["type"],
        content: "",
      },
    ],
    created_at: new Date().toISOString(),
    status: "pending",
  };
}

interface OptimisticPair extends QAPair {
  status?: SubmitStatus;
  nodeId: string;
  pendingId: string;
}

function getPairSignature(pair: Pick<QAPair, "question" | "blocks">) {
  return JSON.stringify({
    question: pair.question.trim(),
    content: pair.blocks.map((block) => block.content).join("\n"),
  });
}

function mergeQAPairs(existing: QAPair[], incoming: QAPair[]) {
  const merged = [...existing];
  const indexById = new Map(existing.map((pair, index) => [pair.id, index]));

  for (const pair of incoming) {
    const index = indexById.get(pair.id);
    if (index === undefined) {
      indexById.set(pair.id, merged.length);
      merged.push(pair);
    } else {
      merged[index] = pair;
    }
  }

  return merged;
}

export function WorkspacePanel({
  node,
  onNodeUpdated,
  onNavigateToNode,
  onRequestAutoSubmit,
  pendingAutoSubmit,
  onPendingAutoSubmitConsumed,
}: WorkspacePanelProps) {
  const nodeId = node?.id ?? null;
  const [qaPairs, setQaPairs] = useState<QAPair[]>([]);
  const [optimisticPairs, setOptimisticPairs] = useState<OptimisticPair[]>([]);
  const [loading, setLoading] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const shouldAutoScrollRef = useRef(false);
  const isProgrammaticScrollRef = useRef(false);
  const { selection, clearSelection, lockSelection } = useTextSelection(
    "[data-answer-content]",
  );
  const selectionRef = useRef(selection);
  const currentNodeIdRef = useRef<string | null>(null);

  selectionRef.current = selection;
  currentNodeIdRef.current = node?.id ?? null;

  useEffect(() => {
    if (selection) {
      lockSelection();
    }
  }, [selection, lockSelection]);

  const scrollToBottom = useCallback(() => {
    const el = scrollContainerRef.current;
    if (!el) return;
    isProgrammaticScrollRef.current = true;
    el.scrollTop = el.scrollHeight;
    requestAnimationFrame(() => {
      isProgrammaticScrollRef.current = false;
    });
  }, []);

  const handleScroll = useCallback(() => {
    const el = scrollContainerRef.current;
    if (!el || isProgrammaticScrollRef.current) return;

    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
    shouldAutoScrollRef.current = distanceFromBottom < 80;
  }, []);

  const loadNodeDetail = useCallback(async (nodeId: string) => {
    const res = await getNode(nodeId);
    if (res.ok) {
      if (currentNodeIdRef.current !== nodeId) return;

      const incoming = (res.data.qa_pairs ?? []).map((pair) => ({
        ...pair,
        status: pair.status ?? "complete",
      }));

      setQaPairs((prev) => mergeQAPairs(prev, incoming));
      setOptimisticPairs((prev) =>
        prev.filter((pair) => {
          if (pair.nodeId !== nodeId) return true;
          if (pair.status === "error") return true;

          const pairSignature = getPairSignature(pair);
          return !incoming.some(
            (incomingPair) =>
              incomingPair.id === pair.id ||
              getPairSignature(incomingPair) === pairSignature,
          );
        }),
      );
    }
  }, []);

  useEffect(() => {
    let active = true;

    const syncNode = async () => {
      if (!nodeId) {
        setQaPairs([]);
        setOptimisticPairs([]);
        return;
      }

      setQaPairs([]);
      setOptimisticPairs([]);

      await loadNodeDetail(nodeId);
      if (!active) return;
    };

    void syncNode();

    return () => {
      active = false;
    };
  }, [nodeId, loadNodeDetail]);

  const submitQuestion = useCallback(
    async (nodeId: string, question: string, tempId?: string) => {
      const trimmed = question.trim();
      if (!trimmed || loading) return;

      const pendingId = tempId ?? crypto.randomUUID();
      const pendingPair = createPendingPair(trimmed);
      pendingPair.id = pendingId;

      shouldAutoScrollRef.current = true;
      setLoading(true);
      setOptimisticPairs((prev) => {
        if (prev.some((pair) => pair.id === pendingId)) return prev;
        return [...prev, { ...pendingPair, nodeId, pendingId }];
      });

      try {
        await streamChatOnNode(nodeId, { question: trimmed }, {
          onDelta: (delta) => {
            setOptimisticPairs((prev) =>
              prev.map((pair) => {
                if (pair.id !== pendingId) return pair;

                const blocks = pair.blocks.length > 0
                  ? pair.blocks.map((block, index) =>
                      index === 0
                        ? { ...block, content: block.content + delta }
                        : block,
                    )
                  : [
                      {
                        id: crypto.randomUUID(),
                        type: "paragraph" as Block["type"],
                        content: delta,
                      },
                    ];

                return {
                  ...pair,
                  blocks,
                };
              }),
            );
          },
          onQAPairReady: (qaPair) => {
            setQaPairs((prev) => {
              const existing = prev.filter((pair) => pair.id !== pendingId);
              return mergeQAPairs(existing, [
                {
                  ...qaPair,
                  status: "complete",
                },
              ]);
            });
            setOptimisticPairs((prev) =>
              prev.filter((pair) => pair.pendingId !== pendingId),
            );
          },
          onCompleted: () => {
            setOptimisticPairs((prev) =>
              prev.map((pair) =>
                pair.pendingId === pendingId
                  ? { ...pair, status: "complete" }
                  : pair,
              ),
            );
            Promise.resolve(onNodeUpdated?.(nodeId))
              .catch(() => {
                setOptimisticPairs((prev) =>
                  prev.map((pair) =>
                    pair.pendingId === pendingId
                      ? { ...pair, status: "complete" }
                      : pair,
                  ),
                );
              });
          },
          onError: () => {
            setOptimisticPairs((prev) =>
              prev.map((pair) =>
                pair.id === pendingId ? { ...pair, status: "error" } : pair,
              ),
            );
          },
        });
      } catch {
        setOptimisticPairs((prev) =>
          prev.map((pair) =>
            pair.id === pendingId ? { ...pair, status: "error" } : pair,
          ),
        );
      } finally {
        setLoading(false);
      }
    },
    [loading, onNodeUpdated],
  );

  useEffect(() => {
    if (!node || !pendingAutoSubmit) return;
    if (pendingAutoSubmit.nodeId !== node.id) return;

    const question = pendingAutoSubmit.question.trim();
    if (!question) return;

    const pendingPair = createPendingPair(question);
    shouldAutoScrollRef.current = true;
    setOptimisticPairs((prev) => [
      ...prev,
      { ...pendingPair, nodeId: node.id, pendingId: pendingPair.id },
    ]);
    onPendingAutoSubmitConsumed?.();
    void submitQuestion(node.id, question, pendingPair.id);
  }, [
    node,
    pendingAutoSubmit,
    onPendingAutoSubmitConsumed,
    submitQuestion,
  ]);

  const handleChat = async (question: string) => {
    if (!node) return;

    const trimmed = question.trim();
    if (!trimmed) return;

    const pendingPair = createPendingPair(trimmed);
    shouldAutoScrollRef.current = true;
    setOptimisticPairs((prev) => [
      ...prev,
      { ...pendingPair, nodeId: node.id, pendingId: pendingPair.id },
    ]);
    void submitQuestion(node.id, trimmed, pendingPair.id);
  };

  const handleCreateAnchorChild = async (question: string) => {
    const sel = selectionRef.current;
    if (!sel) return;

    const trimmed = question.trim();
    if (!trimmed) return;

    const res = await createAnchor({
      block_id: sel.blockId,
      start_offset: sel.startOffset,
      end_offset: sel.endOffset,
      quoted_text: sel.text,
      child_question: trimmed,
    });

    if (res.ok) {
      clearSelection();
      onNavigateToNode?.(res.data.child_node.id, res.data.child_node);
      void Promise.resolve(onNodeUpdated(res.data.child_node.id));
      onRequestAutoSubmit?.({
        nodeId: res.data.child_node.id,
        question: trimmed,
      });
    }
  };

  const displayPairs = (() => {
    const pairs = [
      ...qaPairs,
      ...optimisticPairs.filter((pair) => pair.nodeId === (nodeId ?? "")),
    ];
    const seenIds = new Set<string>();
    const seenSignatures = new Set<string>();
    return pairs.filter((pair) => {
      const signature = getPairSignature(pair);
      if (seenIds.has(pair.id) || seenSignatures.has(signature)) return false;

      seenIds.add(pair.id);
      seenSignatures.add(signature);
      return true;
    });
  })();
  const hasPendingPair = displayPairs.some((pair) => pair.status === "pending");

  useEffect(() => {
    const el = scrollContainerRef.current;
    if (!el) return;

    const distanceFromBottom = el.scrollHeight - el.scrollTop - el.clientHeight;
    const isNearBottom = distanceFromBottom < 80;

    if (!isNearBottom && hasPendingPair) {
      shouldAutoScrollRef.current = false;
      return;
    }

    if (!shouldAutoScrollRef.current && !isNearBottom) return;

    scrollToBottom();

    if (!hasPendingPair) {
      shouldAutoScrollRef.current = false;
    }
  }, [displayPairs, hasPendingPair, scrollToBottom]);

  if (!node) {
    return (
      <div className="flex items-center justify-center h-full text-muted-foreground">
        <div className="text-center space-y-2">
          <svg
            className="w-12 h-12 mx-auto text-muted-foreground/30"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={1}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
            />
          </svg>
          <p className="text-sm">请选择一个节点开始探索</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-background">
      <div
        ref={scrollContainerRef}
        className="flex-1 overflow-auto"
        onScroll={handleScroll}
      >
        <div className="max-w-3xl mx-auto px-8 py-6">
          <AnswerView qaPairs={displayPairs} />
        </div>
      </div>

      <div className="border-t border-border">
        <div className="max-w-3xl mx-auto px-8 py-4">
          <QuestionInput
            onSubmit={handleChat}
            loading={loading}
            placeholder="继续提问，Enter 发送"
          />
        </div>
      </div>

      {selection && selection.blockId && (
        <QuotePopover
          quotedText={selection.text}
          rect={selection.rect}
          onCreateChild={handleCreateAnchorChild}
          onClose={clearSelection}
        />
      )}
    </div>
  );
}
