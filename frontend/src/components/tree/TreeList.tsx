import { useState } from "react";
import type { Tree } from "@/types/tree";
import { Button } from "@/components/ui/button";

interface TreeListProps {
  trees: Tree[];
  onSelect: (treeId: string) => void;
  onCreate: (title: string, question: string) => void;
}

export function TreeList({ trees, onSelect, onCreate }: TreeListProps) {
  const [showCreate, setShowCreate] = useState(false);
  const [title, setTitle] = useState("");
  const [question, setQuestion] = useState("");

  const handleCreate = () => {
    if (!title.trim() || !question.trim()) return;
    onCreate(title.trim(), question.trim());
    setTitle("");
    setQuestion("");
    setShowCreate(false);
  };

  return (
    <div className="flex flex-col items-center justify-center h-full p-8">
      <h1 className="text-2xl font-bold mb-2">Cognitree</h1>
      <p className="text-muted-foreground mb-8">
        选择一棵思维树开始探索，或创建新的
      </p>

      {showCreate ? (
        <div className="w-full max-w-md space-y-4 rounded-xl border border-border bg-card p-6">
          <h3 className="font-semibold">创建新的思维树</h3>
          <input
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-ring focus:ring-2 focus:ring-ring/20"
            placeholder="思维树标题（如：分布式系统学习）"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            autoFocus
          />
          <input
            className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-ring focus:ring-2 focus:ring-ring/20"
            placeholder="根问题（如：为什么分布式系统这么复杂？）"
            value={question}
            onChange={(e) => setQuestion(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleCreate()}
          />
          <div className="flex gap-2">
            <Button onClick={handleCreate}>创建</Button>
            <Button variant="outline" onClick={() => setShowCreate(false)}>
              取消
            </Button>
          </div>
        </div>
      ) : (
        <div className="w-full max-w-md space-y-3">
          <Button className="w-full" onClick={() => setShowCreate(true)}>
            + 创建新的思维树
          </Button>

          {trees.length > 0 && (
            <div className="space-y-2 mt-4">
              {trees.map((tree) => (
                <button
                  key={tree.id}
                  className="w-full text-left rounded-xl border border-border bg-card p-4 hover:bg-muted transition-colors"
                  onClick={() => onSelect(tree.id)}
                >
                  <p className="font-medium">{tree.title}</p>
                  <p className="text-xs text-muted-foreground mt-1">
                    {new Date(tree.created_at).toLocaleDateString()}
                  </p>
                </button>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
