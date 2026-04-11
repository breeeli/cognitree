import { Button } from "@/components/ui/button";

export function AccountPanel() {
  return (
    <div className="px-4 py-4 border-t border-border">
      <div className="text-xs font-medium text-muted-foreground mb-3">账户</div>
      <div className="flex items-center gap-3">
        <div className="w-9 h-9 rounded-full bg-primary/10 text-primary flex items-center justify-center text-sm font-medium shrink-0">
          U
        </div>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium truncate">用户名</p>
          <p className="text-xs text-muted-foreground truncate">
            user@example.com
          </p>
        </div>
        <Button variant="outline" size="sm">
          设置
        </Button>
      </div>
    </div>
  );
}
