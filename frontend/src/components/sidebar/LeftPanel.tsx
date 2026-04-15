import { useNavigate } from "react-router-dom";
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
}

export function LeftPanel({
  trees,
  currentTreeId,
  treeNodes,
  treeName,
  selectedNodeId,
  onSelectNode,
  onSelectTree,
}: LeftPanelProps) {
  const navigate = useNavigate();

  return (
    <div className="flex flex-col h-full bg-card border-r border-border">
      <SidebarToolbar
        trees={trees}
        currentTreeId={currentTreeId}
        onSelectTree={onSelectTree}
        onCreateNew={() => navigate("/trees/new")}
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
    </div>
  );
}
