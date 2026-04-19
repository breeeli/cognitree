import { useState, useEffect, useCallback, useRef } from "react";
import type { Tree, Node } from "@/types/tree";
import { listTrees, getTree } from "@/api/trees";
import { getNode } from "@/api/nodes";
import { LeftPanel } from "@/components/sidebar/LeftPanel";
import { WorkspacePanel } from "@/components/workspace/WorkspacePanel";
import { CreateTreeWorkspace } from "@/components/workspace/CreateTreeWorkspace";

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
  const [isCreatingTree, setIsCreatingTree] = useState(false);
  const currentTreeIdRef = useRef<string | null>(currentTreeId);
  const selectedNodeIdRef = useRef<string | null>(selectedNodeId);

  currentTreeIdRef.current = currentTreeId;
  selectedNodeIdRef.current = selectedNodeId;

  useEffect(() => {
    listTrees().then((res) => {
      if (res.ok) setTrees(res.data);
    });
  }, []);

  const loadNodeDetail = useCallback(async (nodeId: string) => {
    const res = await getNode(nodeId);
    if (res.ok) {
      if (selectedNodeIdRef.current === nodeId) {
        setSelectedNode(res.data);
      }
    }
  }, []);

  const setActiveNode = useCallback(
    (nodeId: string, nextNode?: Node) => {
      selectedNodeIdRef.current = nodeId;
      setSelectedNodeId(nodeId);
      if (nextNode) {
        setSelectedNode(nextNode);
        setTreeNodes((prev) =>
          prev.some((node) => node.id === nextNode.id) ? prev : [...prev, nextNode],
        );
      }
      void loadNodeDetail(nodeId);
    },
    [loadNodeDetail],
  );

  const loadTree = useCallback(
    async (treeId: string) => {
      const res = await getTree(treeId);
      if (res.ok) {
        currentTreeIdRef.current = treeId;
        setCurrentTreeId(treeId);
        localStorage.setItem(STORAGE_KEY_TREE, treeId);
        setTreeNodes(res.data.nodes);
        setTreeName(res.data.tree.title);
        const root = res.data.nodes.find((n) => !n.parent_node_id);
        if (root) {
          setActiveNode(root.id);
        }
      } else {
        localStorage.removeItem(STORAGE_KEY_TREE);
        currentTreeIdRef.current = null;
        setCurrentTreeId(null);
      }
    },
    [setActiveNode],
  );

  useEffect(() => {
    const savedId = localStorage.getItem(STORAGE_KEY_TREE);
    if (savedId) {
      void loadTree(savedId);
    }
  }, [loadTree]);

  const handleSelectNode = (nodeId: string, nextNode?: Node) => {
    setIsCreatingTree(false);
    setActiveNode(nodeId, nextNode);
  };

  const handleSelectTree = async (treeId: string) => {
    setIsCreatingTree(false);
    await loadTree(treeId);
  };

  const handleNodeUpdated = async (targetNodeId?: string) => {
    const activeTreeId = currentTreeIdRef.current;
    if (activeTreeId) {
      const res = await getTree(activeTreeId);
      if (res.ok) setTreeNodes(res.data.nodes);
    }

    const nodeIdToRefresh = targetNodeId ?? selectedNodeIdRef.current;
    if (nodeIdToRefresh) {
      await loadNodeDetail(nodeIdToRefresh);
    }
  };

  const handlePendingAutoSubmitConsumed = () => {
    setPendingAutoSubmit(null);
  };

  const handleCreateNewTree = () => {
    localStorage.removeItem(STORAGE_KEY_TREE);
    currentTreeIdRef.current = null;
    selectedNodeIdRef.current = null;
    setCurrentTreeId(null);
    setTreeNodes([]);
    setTreeName("");
    setSelectedNodeId(null);
    setSelectedNode(null);
    setPendingAutoSubmit(null);
    setIsCreatingTree(true);
  };

  const handleTreeCreated = async ({
    treeId,
    rootNodeId,
    tree,
  }: {
    treeId: string;
    rootNodeId: string;
    tree: Tree;
  }) => {
    setIsCreatingTree(false);
    setTrees((prev) => [tree, ...prev.filter((item) => item.id !== treeId)]);
    await loadTree(treeId);
    setActiveNode(rootNodeId);
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
          onSelectTree={handleSelectTree}
          onCreateNew={handleCreateNewTree}
        />
      </div>

      <div className="flex-1 min-w-0">
        {isCreatingTree ? (
          <CreateTreeWorkspace onCreated={handleTreeCreated} />
        ) : (
          <WorkspacePanel
            node={selectedNode}
            onNodeUpdated={handleNodeUpdated}
            onNavigateToNode={handleSelectNode}
            onRequestAutoSubmit={setPendingAutoSubmit}
            pendingAutoSubmit={pendingAutoSubmit}
            onPendingAutoSubmitConsumed={handlePendingAutoSubmitConsumed}
          />
        )}
      </div>
    </div>
  );
}
