import { useState, useEffect, useCallback, useRef } from "react";
import type { Node, QAPair } from "@/types/tree";
import { AnswerView } from "./AnswerView";
import { QuestionInput } from "./QuestionInput";
import { QuotePopover } from "./QuotePopover";
import { useTextSelection } from "@/hooks/useTextSelection";
import { chatOnNode } from "@/api/chat";
import { getNode, createAnchor } from "@/api/nodes";
import { Button } from "@/components/ui/button";

interface WorkspacePanelProps {
  node: Node | null;
  onNodeUpdated: () => void;
  onNavigateToNode?: (nodeId: string) => void;
  onCreateChild?: () => void;
}

export function WorkspacePanel({
  node,
  onNodeUpdated,
  onNavigateToNode,
  onCreateChild,
}: WorkspacePanelProps) {
  const [qaPairs, setQaPairs] = useState<QAPair[]>([]);
  const [loading, setLoading] = useState(false);
  const { selection, clearSelection, lockSelection } = useTextSelection(
    "[data-answer-content]",
  );
  const selectionRef = useRef(selection);
  selectionRef.current = selection;

  useEffect(() => {
    if (selection) lockSelection();
  }, [selection, lockSelection]);

  const loadNodeDetail = useCallback(async (nodeId: string) => {
    const res = await getNode(nodeId);
    if (res.ok) {
      setQaPairs(res.data.qa_pairs ?? []);
    }
  }, []);

  useEffect(() => {
    if (node) {
      loadNodeDetail(node.id);
    } else {
      setQaPairs([]);
    }
  }, [node?.id, loadNodeDetail]);

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
          <p className="text-sm">选择一个节点开始探索</p>
        </div>
      </div>
    );
  }

  const handleChat = async (question: string) => {
    setLoading(true);
    const res = await chatOnNode(node.id, question);
    if (res.ok) {
      setQaPairs((prev) => [...prev, res.data.qa_pair]);
      onNodeUpdated();
    }
    setLoading(false);
  };

  const handleCreateAnchorChild = async (question: string) => {
    const sel = selectionRef.current;
    if (!sel) return;
    const res = await createAnchor({
      block_id: sel.blockId,
      start_offset: sel.startOffset,
      end_offset: sel.endOffset,
      quoted_text: sel.text,
      child_question: question,
    });
    if (res.ok) {
      clearSelection();
      onNodeUpdated();
      onNavigateToNode?.(res.data.child_node.id);
    }
  };

  return (
    <div className="flex flex-col h-full bg-background">
      <div className="border-b border-border">
        <div className="max-w-3xl mx-auto px-8 py-5">
          <div className="flex items-start justify-between gap-4">
            <div className="min-w-0 flex-1">
              <span className="text-xs text-muted-foreground">当前问题</span>
              <h2 className="text-xl font-bold mt-1 leading-tight">
                {node.question}
              </h2>
            </div>
            <div className="flex items-center gap-2 shrink-0 pt-4">
              <Button size="sm" onClick={onCreateChild}>
                提出新问题
              </Button>
            </div>
          </div>
        </div>
      </div>

      <div className="flex-1 overflow-auto">
        <div className="max-w-3xl mx-auto px-8 py-6">
          <AnswerView qaPairs={qaPairs} />
        </div>
      </div>

      <div className="border-t border-border">
        <div className="max-w-3xl mx-auto px-8 py-4">
          <QuestionInput
            onSubmit={handleChat}
            loading={loading}
            placeholder="继续提问（Enter '发送'）..."
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
