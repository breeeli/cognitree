import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { QAPair } from "@/types/tree";

interface AnswerViewProps {
  qaPairs: QAPair[];
}

export function AnswerView({ qaPairs }: AnswerViewProps) {
  if (qaPairs.length === 0) {
    return (
      <div className="text-center text-muted-foreground py-12">
        <p>还没有问答记录</p>
        <p className="text-sm mt-1">在下方输入框中提问开始探索</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {qaPairs.map((qaPair) => (
        <div key={qaPair.id} className="space-y-4">
          {/* User question — right aligned */}
          <div className="flex justify-end">
            <div className="max-w-[75%] flex items-start gap-2">
              <div className="rounded-2xl rounded-tr-sm bg-primary text-primary-foreground px-4 py-2.5">
                <p className="text-sm">{qaPair.question}</p>
              </div>
              <span className="shrink-0 w-7 h-7 rounded-full bg-primary/10 text-primary flex items-center justify-center text-xs font-medium mt-0.5">
                U
              </span>
            </div>
          </div>

          {/* AI answer — left aligned */}
          <div className="flex justify-start">
            <div className="max-w-[85%] flex items-start gap-2">
              <span className="shrink-0 w-7 h-7 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium mt-0.5">
                AI
              </span>
              <div className="rounded-2xl rounded-tl-sm bg-card border border-border px-4 py-3">
                <div className="prose prose-sm max-w-none" data-answer-content>
                  {qaPair.blocks.map((block) => (
                    <div key={block.id} data-block-id={block.id}>
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {block.content}
                      </ReactMarkdown>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
}
