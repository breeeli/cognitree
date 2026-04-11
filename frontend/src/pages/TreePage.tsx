import { useState, useEffect, useCallback } from "react";
import type { Tree, Node } from "@/types/tree";
import { listTrees, getTree, createTree } from "@/api/trees";
import { getNode, createChildNode } from "@/api/nodes";
import { LeftPanel } from "@/components/sidebar/LeftPanel";
import { WorkspacePanel } from "@/components/workspace/WorkspacePanel";

const STORAGE_KEY_TREE = "cognitree:currentTreeId";

export function TreePage() {
  const [trees, setTrees] = useState<Tree[]>([]);
  const [currentTreeId, setCurrentTreeId] = useState<string | null>(
    () => localStorage.getItem(STORAGE_KEY_TREE),
  );
  const [treeNodes, setTreeNodes] = useState<Node[]>([]);
  const [treeName, setTreeName] = useState("");
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [showCreateChild, setShowCreateChild] = useState(false);
  const [childQuestion, setChildQuestion] = useState("");

  useEffect(() => {
    listTrees().then((res) => {
      if (res.ok) setTrees(res.data);
    });
  }, []);

  const loadTree = useCallback(async (treeId: string) => {
    const res = await getTree(treeId);
    if (res.ok) {
      setCurrentTreeId(treeId);
      localStorage.setItem(STORAGE_KEY_TREE, treeId);
      setTreeNodes(res.data.nodes);
      setTreeName(res.data.tree.title);
      const root = res.data.nodes.find((n) => !n.parent_node_id);
      if (root) {
        setSelectedNodeId(root.id);
        loadNodeDetail(root.id);
      }
    } else {
      localStorage.removeItem(STORAGE_KEY_TREE);
      setCurrentTreeId(null);
    }
  }, []);

  useEffect(() => {
    const savedId = localStorage.getItem(STORAGE_KEY_TREE);
    if (savedId) {
      loadTree(savedId);
    }
  }, [loadTree]);

  const loadNodeDetail = async (nodeId: string) => {
    const res = await getNode(nodeId);
    if (res.ok) {
      setSelectedNode(res.data);
    }
  };

  const handleSelectNode = (nodeId: string) => {
    setSelectedNodeId(nodeId);
    loadNodeDetail(nodeId);
  };

  const handleCreateTree = async (title: string, question: string) => {
    const res = await createTree(title, question);
    if (res.ok) {
      setTrees((prev) => [res.data.tree, ...prev]);
      loadTree(res.data.tree.id);
    }
  };

  const handleNodeUpdated = () => {
    if (currentTreeId) {
      getTree(currentTreeId).then((res) => {
        if (res.ok) setTreeNodes(res.data.nodes);
      });
    }
  };

  const handleCreateChild = async () => {
    if (!selectedNodeId || !childQuestion.trim()) return;
    const res = await createChildNode(selectedNodeId, childQuestion.trim());
    if (res.ok) {
      setChildQuestion("");
      setShowCreateChild(false);
      handleNodeUpdated();
      handleSelectNode(res.data.id);
    }
  };

  return (
    <div className="flex h-full">
      {/* Left panel: toolbar + tree + account */}
      <div className="w-72 shrink-0 overflow-hidden">
        <LeftPanel
          trees={trees}
          currentTreeId={currentTreeId}
          treeNodes={treeNodes}
          treeName={treeName}
          selectedNodeId={selectedNodeId}
          onSelectNode={handleSelectNode}
          onSelectTree={loadTree}
          onCreateTree={handleCreateTree}
        />
      </div>

      {/* Right panel: workspace */}
      <div className="flex-1 min-w-0">
        <WorkspacePanel
          node={selectedNode}
          onNodeUpdated={handleNodeUpdated}
          onNavigateToNode={handleSelectNode}
          onCreateChild={() => setShowCreateChild(true)}
        />
      </div>

      {/* Create child question dialog */}
      {showCreateChild && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-card rounded-xl border border-border p-6 w-full max-w-md space-y-4 shadow-lg">
            <h3 className="font-semibold">提出新问题</h3>
            <p className="text-sm text-muted-foreground">
              在当前节点下创建一个新的子问题
            </p>
            <input
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
              placeholder="输入子问题..."
              value={childQuestion}
              onChange={(e) => setChildQuestion(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleCreateChild()}
              autoFocus
            />
            <div className="flex gap-2 justify-end">
              <button
                className="px-4 py-2 text-sm rounded-lg hover:bg-muted transition-colors"
                onClick={() => {
                  setShowCreateChild(false);
                  setChildQuestion("");
                }}
              >
                取消
              </button>
              <button
                className="px-4 py-2 text-sm rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                onClick={handleCreateChild}
                disabled={!childQuestion.trim()}
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
