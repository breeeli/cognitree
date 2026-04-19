import { useState, useRef, useEffect } from "react";
import { Button } from "@/components/ui/button";

interface QuotePopoverProps {
  quotedText: string;
  rect: DOMRect;
  onCreateChild: (question: string) => void;
  onClose: () => void;
}

export function QuotePopover({
  quotedText,
  rect,
  onCreateChild,
  onClose,
}: QuotePopoverProps) {
  const [question, setQuestion] = useState("");
  const [expanded, setExpanded] = useState(false);
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(e.target as HTMLElement)
      ) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [onClose]);

  const top = rect.top - 10;
  const left = rect.left + rect.width / 2;

  const handleSubmit = () => {
    if (!question.trim()) return;
    onCreateChild(question.trim());
  };

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(quotedText);
    } catch {
      // Best-effort clipboard copy; keep popover open either way.
    }
  };

  return (
    <div
      ref={popoverRef}
      className="fixed z-50 -translate-x-1/2 -translate-y-full"
      style={{ top: `${top}px`, left: `${left}px` }}
    >
      <div className="bg-card rounded-xl border border-border shadow-lg p-4 w-80 space-y-3">
        <div className="flex items-center justify-between gap-2">
          <div className="text-xs text-muted-foreground font-medium">
            选中文本
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="sm" onClick={handleCopy}>
              复制
            </Button>
            {!expanded && (
              <Button size="sm" onClick={() => setExpanded(true)}>
                基于引用追问
              </Button>
            )}
          </div>
        </div>
        <div className="text-sm bg-muted/50 rounded-lg p-2 border-l-2 border-primary max-h-20 overflow-auto">
          {quotedText}
        </div>
        {expanded && (
          <>
            <input
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:border-ring focus:ring-2 focus:ring-ring/20"
              placeholder="基于引用内容提出子问题..."
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
            />
            <div className="flex gap-2 justify-end">
              <Button variant="ghost" size="sm" onClick={onClose}>
                取消
              </Button>
              <Button size="sm" onClick={handleSubmit} disabled={!question.trim()}>
                创建子问题
              </Button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
