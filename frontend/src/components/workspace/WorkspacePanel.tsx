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
  onNodeUpdated: () => void;
  onNavigateToNode?: (nodeId: string) => void;
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
}

export function WorkspacePanel({
  node,
  onNodeUpdated,
  onNavigateToNode,
  onRequestAutoSubmit,
  pendingAutoSubmit,
  onPendingAutoSubmitConsumed,
}: WorkspacePanelProps) {
  const [qaPairs, setQaPairs] = useState<QAPair[]>([]);
  const [optimisticPairs, setOptimisticPairs] = useState<OptimisticPair[]>([]);
  const [loading, setLoading] = useState(false);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const shouldAutoScrollRef = useRef(false);
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
    el.scrollTop = el.scrollHeight;
  }, []);

  const loadNodeDetail = useCallback(async (nodeId: string) => {
    const res = await getNode(nodeId);
    if (res.ok) {
      setQaPairs(
        (res.data.qa_pairs ?? []).map((pair) => ({
          ...pair,
          status: pair.status ?? "complete",
        })),
      );
    }
  }, []);

  useEffect(() => {
    let active = true;

    const syncNode = async () => {
      if (!node) {
        setQaPairs([]);
        setOptimisticPairs([]);
        return;
      }

      await loadNodeDetail(node.id);
      if (!active) return;
      if (currentNodeIdRef.current !== node.id) return;
    };

    void syncNode();

    return () => {
      active = false;
    };
  }, [node?.id, loadNodeDetail]);

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
        return [...prev, { ...pendingPair, nodeId }];
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
          onCompleted: () => {
            setOptimisticPairs((prev) =>
              prev.filter((pair) => pair.id !== pendingId),
            );
            if (currentNodeIdRef.current === nodeId) {
              onNodeUpdated();
            }
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
      { ...pendingPair, nodeId: node.id },
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
      { ...pendingPair, nodeId: node.id },
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
      onNodeUpdated();
      onRequestAutoSubmit?.({
        nodeId: res.data.child_node.id,
        question: trimmed,
      });
      onNavigateToNode?.(res.data.child_node.id);
    }
  };

  const displayPairs = [
    ...qaPairs,
    ...optimisticPairs.filter((pair) => pair.nodeId === (node?.id ?? "")),
  ];
  const hasPendingPair = displayPairs.some((pair) => pair.status === "pending");

  useEffect(() => {
    if (!shouldAutoScrollRef.current) return;

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
      <div ref={scrollContainerRef} className="flex-1 overflow-auto">
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
