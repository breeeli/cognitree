import { useState } from "react";
import type { Tree, Node } from "@/types/tree";
import { SidebarToolbar } from "./SidebarToolbar";
import { TreePanel } from "../tree/TreePanel";
import { AccountPanel } from "./AccountPanel";

interface LeftPanelProps {
  trees: Tree[];
  currentTreeId: string | null;
  treeNodes: Node[];
  treeName: string;
  selectedNodeId: string | null;
  onSelectNode: (nodeId: string) => void;
  onSelectTree: (treeId: string) => void;
  onCreateTree: (title: string, question: string) => void;
}

export function LeftPanel({
  trees,
  currentTreeId,
  treeNodes,
  treeName,
  selectedNodeId,
  onSelectNode,
  onSelectTree,
  onCreateTree,
}: LeftPanelProps) {
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [newTitle, setNewTitle] = useState("");
  const [newQuestion, setNewQuestion] = useState("");

  const handleCreate = () => {
    if (!newTitle.trim() || !newQuestion.trim()) return;
    onCreateTree(newTitle.trim(), newQuestion.trim());
    setNewTitle("");
    setNewQuestion("");
    setShowCreateDialog(false);
  };

  return (
    <div className="flex flex-col h-full bg-card border-r border-border">
      <SidebarToolbar
        trees={trees}
        currentTreeId={currentTreeId}
        onSelectTree={onSelectTree}
        onCreateNew={() => setShowCreateDialog(true)}
      />

      <div className="flex-1 min-h-0 overflow-hidden">
        {currentTreeId ? (
          <TreePanel
            nodes={treeNodes}
            selectedNodeId={selectedNodeId}
            onSelectNode={onSelectNode}
            treeName={treeName}
          />
        ) : (
          <div className="flex items-center justify-center h-full px-4">
            <p className="text-sm text-muted-foreground text-center">
              创建或选择一棵思维树开始探索
            </p>
          </div>
        )}
      </div>

      <AccountPanel />

      {showCreateDialog && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card rounded-xl border border-border p-6 w-full max-w-md space-y-4 shadow-lg">
            <h3 className="font-semibold text-base">创建新的思维树</h3>
            <div className="space-y-3">
              <input
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
                placeholder="思维树标题（如：分布式系统学习）"
                value={newTitle}
                onChange={(e) => setNewTitle(e.target.value)}
                autoFocus
              />
              <input
                className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
                placeholder="根问题（如：为什么分布式系统这么复杂？）"
                value={newQuestion}
                onChange={(e) => setNewQuestion(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              />
            </div>
            <div className="flex gap-2 justify-end">
              <button
                className="px-4 py-2 text-sm rounded-lg hover:bg-muted transition-colors"
                onClick={() => {
                  setShowCreateDialog(false);
                  setNewTitle("");
                  setNewQuestion("");
                }}
              >
                取消
              </button>
              <button
                className="px-4 py-2 text-sm rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                onClick={handleCreate}
                disabled={!newTitle.trim() || !newQuestion.trim()}
              >
                创建
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
