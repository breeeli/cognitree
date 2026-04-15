import { useState, useEffect, useCallback } from "react";
import type { Tree, Node } from "@/types/tree";
import { listTrees, getTree } from "@/api/trees";
import { getNode } from "@/api/nodes";
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
  const [pendingAutoSubmit, setPendingAutoSubmit] = useState<{
    nodeId: string;
    question: string;
  } | null>(null);

  useEffect(() => {
    listTrees().then((res) => {
      if (res.ok) setTrees(res.data);
    });
  }, []);

  const loadNodeDetail = useCallback(async (nodeId: string) => {
    const res = await getNode(nodeId);
    if (res.ok) {
      setSelectedNode(res.data);
    }
  }, []);

  const loadTree = useCallback(
    async (treeId: string) => {
      const res = await getTree(treeId);
      if (res.ok) {
        setCurrentTreeId(treeId);
        localStorage.setItem(STORAGE_KEY_TREE, treeId);
        setTreeNodes(res.data.nodes);
        setTreeName(res.data.tree.title);
        const root = res.data.nodes.find((n) => !n.parent_node_id);
        if (root) {
          setSelectedNodeId(root.id);
          void loadNodeDetail(root.id);
        }
      } else {
        localStorage.removeItem(STORAGE_KEY_TREE);
        setCurrentTreeId(null);
      }
    },
    [loadNodeDetail],
  );

  useEffect(() => {
    const savedId = localStorage.getItem(STORAGE_KEY_TREE);
    if (savedId) {
      void loadTree(savedId);
    }
  }, [loadTree]);

  const handleSelectNode = (nodeId: string) => {
    setSelectedNodeId(nodeId);
    void loadNodeDetail(nodeId);
  };

  const handleNodeUpdated = () => {
    if (currentTreeId) {
      getTree(currentTreeId).then((res) => {
        if (res.ok) setTreeNodes(res.data.nodes);
      });
    }

    if (selectedNodeId) {
      void loadNodeDetail(selectedNodeId);
    }
  };

  const handlePendingAutoSubmitConsumed = () => {
    setPendingAutoSubmit(null);
  };

  return (
    <div className="flex h-full">
      <div className="w-72 shrink-0 overflow-hidden">
        <LeftPanel
          trees={trees}
          currentTreeId={currentTreeId}
          treeNodes={treeNodes}
          treeName={treeName}
          selectedNodeId={selectedNodeId}
          onSelectNode={handleSelectNode}
          onSelectTree={loadTree}
        />
      </div>

      <div className="flex-1 min-w-0">
        <WorkspacePanel
          node={selectedNode}
          onNodeUpdated={handleNodeUpdated}
          onNavigateToNode={handleSelectNode}
          onRequestAutoSubmit={setPendingAutoSubmit}
          pendingAutoSubmit={pendingAutoSubmit}
          onPendingAutoSubmitConsumed={handlePendingAutoSubmitConsumed}
        />
      </div>
    </div>
  );
}
