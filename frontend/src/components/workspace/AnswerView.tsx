import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { QAPair } from "@/types/tree";
import { markdownComponents } from "./markdownComponents";

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

          <div className="flex justify-start">
            <div className="max-w-[85%] flex items-start gap-2">
                <span className="shrink-0 w-7 h-7 rounded-full bg-muted text-muted-foreground flex items-center justify-center text-xs font-medium mt-0.5">
                  AI
                </span>
                <div className="rounded-2xl rounded-tl-sm bg-card border border-border px-4 py-3">
                  <div className="space-y-3 text-sm leading-7" data-answer-content>
                    {qaPair.blocks.length > 0 ? (
                      qaPair.blocks.map((block) => (
                        <div
                          key={block.id}
                          data-block-id={block.id}
                          className="space-y-3"
                        >
                          <ReactMarkdown
                            remarkPlugins={[remarkGfm]}
                            components={markdownComponents}
                          >
                            {block.content}
                          </ReactMarkdown>
                        </div>
                      ))
                    ) : qaPair.status === "pending" ? (
                      <div className="text-sm text-muted-foreground">
                        <span className="inline-flex items-center gap-2">
                          <span className="h-2 w-2 rounded-full bg-primary/60 animate-pulse" />
                          AI 正在思考...
                        </span>
                      </div>
                    ) : qaPair.status === "error" ? (
                      <div className="text-sm text-danger">
                        回答发送失败，请稍后重试。
                      </div>
                    ) : (
                      <div className="text-sm text-muted-foreground">
                        AI 正在整理回答...
                      </div>
                    )}
                    {qaPair.status === "pending" && qaPair.blocks.length > 0 && (
                      <div className="text-xs text-muted-foreground">
                        <span className="inline-flex items-center gap-2">
                          <span className="h-2 w-2 rounded-full bg-primary/60 animate-pulse" />
                          AI 正在生成...
                        </span>
                      </div>
                    )}
                    {qaPair.status === "error" && qaPair.blocks.length > 0 && (
                      <div className="text-xs text-danger">
                        回答发送失败，请稍后重试。
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
        </div>
      ))}
    </div>
  );
}
