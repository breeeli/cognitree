import { useState } from "react";
import type { Tree } from "@/types/tree";
import { Button } from "@/components/ui/button";

interface SidebarToolbarProps {
  trees: Tree[];
  currentTreeId: string | null;
  onSelectTree: (treeId: string) => void;
  onCreateNew: () => void;
}

export function SidebarToolbar({
  trees,
  currentTreeId,
  onSelectTree,
  onCreateNew,
}: SidebarToolbarProps) {
  const [showTreeMenu, setShowTreeMenu] = useState(false);

  return (
    <div className="px-4 py-4 border-b border-border space-y-3">
      <div className="flex items-center justify-between">
        <h1 className="text-base font-semibold tracking-tight">工具栏</h1>
      </div>

      <div className="flex items-center gap-2">
        <Button size="sm" className="flex-1" onClick={onCreateNew}>
          <svg
            className="w-4 h-4 mr-1.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M12 4v16m8-8H4"
            />
          </svg>
          创建新树
        </Button>
      </div>

      <div className="flex items-center gap-2">
        <div className="relative flex-1">
          <button
            className="w-full flex items-center justify-between rounded-lg border border-input bg-background px-3 py-1.5 text-sm hover:bg-muted transition-colors"
            onClick={() => setShowTreeMenu(!showTreeMenu)}
          >
            <span className="truncate text-muted-foreground">
              {currentTreeId
                ? (trees.find((t) => t.id === currentTreeId)?.title ??
                  "选择思维树")
                : "选择思维树"}
            </span>
            <svg
              className="w-3.5 h-3.5 ml-2 shrink-0 text-muted-foreground"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M19 9l-7 7-7-7"
              />
            </svg>
          </button>

          {showTreeMenu && trees.length > 0 && (
            <div className="absolute top-full left-0 right-0 mt-1 bg-popover border border-border rounded-lg shadow-lg z-50 max-h-48 overflow-auto">
              {trees.map((tree) => (
                <button
                  key={tree.id}
                  className={`w-full text-left px-3 py-2 text-sm hover:bg-muted transition-colors first:rounded-t-lg last:rounded-b-lg ${
                    tree.id === currentTreeId
                      ? "bg-primary/10 text-primary font-medium"
                      : ""
                  }`}
                  onClick={() => {
                    onSelectTree(tree.id);
                    setShowTreeMenu(false);
                  }}
                >
                  {tree.title}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" className="flex-1">
          <svg
            className="w-3.5 h-3.5 mr-1.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"
            />
          </svg>
          导入
        </Button>
        <Button variant="outline" size="sm" className="flex-1">
          <svg
            className="w-3.5 h-3.5 mr-1.5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
            />
          </svg>
          导出
        </Button>
      </div>
    </div>
  );
}
