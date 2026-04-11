import { Button } from "@/components/ui/button";

interface ToolbarProps {
  onCreateChild: () => void;
  onBack: () => void;
  treeName: string;
}

export function Toolbar({ onCreateChild, onBack, treeName }: ToolbarProps) {
  return (
    <div className="flex items-center justify-between px-4 py-2 border-b border-border bg-card">
      <div className="flex items-center gap-3">
        <Button variant="ghost" size="sm" onClick={onBack}>
          <svg
            className="w-4 h-4 mr-1"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            strokeWidth={2}
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              d="M15 19l-7-7 7-7"
            />
          </svg>
          返回
        </Button>
        <span className="text-sm font-medium text-muted-foreground">
          {treeName}
        </span>
      </div>
      <div className="flex items-center gap-2">
        <Button variant="outline" size="sm" onClick={onCreateChild}>
          提出新问题
        </Button>
      </div>
    </div>
  );
}
