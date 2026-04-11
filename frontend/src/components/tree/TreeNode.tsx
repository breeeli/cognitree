import { useState } from "react";
import type { TreeNode as TreeNodeType } from "@/types/tree";
import { cn } from "@/lib/utils";

interface TreeNodeProps {
  node: TreeNodeType;
  selectedNodeId: string | null;
  onSelect: (nodeId: string) => void;
  depth?: number;
}

export function TreeNode({
  node,
  selectedNodeId,
  onSelect,
  depth = 0,
}: TreeNodeProps) {
  const [expanded, setExpanded] = useState(true);
  const hasChildren = node.children.length > 0;
  const isSelected = node.id === selectedNodeId;

  return (
    <div>
      <div
        className={cn(
          "flex items-center gap-1.5 py-1.5 px-2 rounded-md cursor-pointer text-sm transition-colors",
          isSelected
            ? "bg-primary/10 text-primary font-medium"
            : "text-foreground/70 hover:bg-muted hover:text-foreground",
        )}
        style={{ paddingLeft: `${depth * 16 + 8}px` }}
        onClick={() => onSelect(node.id)}
      >
        {hasChildren ? (
          <button
            className="w-4 h-4 flex items-center justify-center text-muted-foreground hover:text-foreground shrink-0"
            onClick={(e) => {
              e.stopPropagation();
              setExpanded(!expanded);
            }}
          >
            <svg
              className={cn(
                "w-3 h-3 transition-transform",
                expanded && "rotate-90",
              )}
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M9 5l7 7-7 7"
              />
            </svg>
          </button>
        ) : (
          <span className="w-4 h-4 flex items-center justify-center shrink-0">
            <span
              className={cn(
                "w-1.5 h-1.5 rounded-full",
                isSelected
                  ? "bg-primary"
                  : node.status === "answered"
                    ? "bg-success"
                    : "bg-muted-foreground/40",
              )}
            />
          </span>
        )}
        <span className="truncate">{node.question}</span>
      </div>

      {hasChildren && expanded && (
        <div>
          {node.children.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              selectedNodeId={selectedNodeId}
              onSelect={onSelect}
              depth={depth + 1}
            />
          ))}
        </div>
      )}
    </div>
  );
}
