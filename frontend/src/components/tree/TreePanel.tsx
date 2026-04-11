import type { Node, TreeNode as TreeNodeType } from "@/types/tree";
import { TreeNode } from "./TreeNode";

interface TreePanelProps {
  nodes: Node[];
  selectedNodeId: string | null;
  onSelectNode: (nodeId: string) => void;
  treeName?: string;
}

function buildTree(nodes: Node[]): TreeNodeType[] {
  const nodeMap = new Map<string, TreeNodeType>();
  const roots: TreeNodeType[] = [];

  for (const n of nodes) {
    nodeMap.set(n.id, { ...n, children: [] });
  }

  for (const n of nodes) {
    const treeNode = nodeMap.get(n.id)!;
    if (n.parent_node_id) {
      const parent = nodeMap.get(n.parent_node_id);
      if (parent) {
        parent.children.push(treeNode);
      }
    } else {
      roots.push(treeNode);
    }
  }

  return roots;
}

export function TreePanel({
  nodes,
  selectedNodeId,
  onSelectNode,
}: TreePanelProps) {
  const tree = buildTree(nodes);

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-border">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium text-muted-foreground">
            问题树
          </span>
          <span className="text-xs text-muted-foreground">
            （左侧树状结构）
          </span>
        </div>
      </div>
      <div className="flex-1 overflow-auto py-2 px-1">
        {tree.length > 0 ? (
          tree.map((root) => (
            <TreeNode
              key={root.id}
              node={root}
              selectedNodeId={selectedNodeId}
              onSelect={onSelectNode}
            />
          ))
        ) : (
          <div className="px-4 py-8 text-center text-xs text-muted-foreground">
            暂无节点
          </div>
        )}
      </div>
    </div>
  );
}
