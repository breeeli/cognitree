import { useState } from "react";
import { Button } from "@/components/ui/button";

interface QuestionInputProps {
  onSubmit: (question: string) => void;
  loading?: boolean;
  placeholder?: string;
}

export function QuestionInput({
  onSubmit,
  loading = false,
  placeholder = "继续提问...",
}: QuestionInputProps) {
  const [question, setQuestion] = useState("");

  const handleSubmit = () => {
    const trimmed = question.trim();
    if (!trimmed || loading) return;
    onSubmit(trimmed);
    setQuestion("");
  };

  return (
    <div className="flex gap-3 items-end">
      <textarea
        className="flex-1 rounded-xl border border-input bg-background px-4 py-3 text-sm outline-none focus:border-primary focus:ring-2 focus:ring-primary/20 resize-none min-h-[80px] max-h-[160px]"
        placeholder={placeholder}
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSubmit();
          }
        }}
        rows={3}
        disabled={loading}
      />
      <Button
        className="h-10"
        onClick={handleSubmit}
        disabled={!question.trim() || loading}
      >
        {loading ? "思考中..." : "发送"}
      </Button>
    </div>
  );
}
