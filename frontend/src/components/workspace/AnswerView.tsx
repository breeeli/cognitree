import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { QAPair } from "@/types/tree";
import type { Components } from "react-markdown";

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
                  {qaPair.status === "pending" ? (
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
                  ) : qaPair.blocks.length > 0 ? (
                    qaPair.blocks.map((block) => (
                      <div key={block.id} data-block-id={block.id} className="space-y-3">
                        <ReactMarkdown
                          remarkPlugins={[remarkGfm]}
                          components={markdownComponents}
                        >
                          {block.content}
                        </ReactMarkdown>
                      </div>
                    ))
                  ) : (
                    <div className="text-sm text-muted-foreground">
                      AI 正在整理回答...
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

const markdownComponents: Components = {
  p: ({ children }) => <p className="m-0 first:mt-0 mt-3">{children}</p>,
  h1: ({ children }) => (
    <h1 className="text-lg font-semibold tracking-tight m-0 first:mt-0 mt-4">{children}</h1>
  ),
  h2: ({ children }) => (
    <h2 className="text-base font-semibold tracking-tight m-0 first:mt-0 mt-4">{children}</h2>
  ),
  h3: ({ children }) => (
    <h3 className="text-sm font-semibold tracking-tight m-0 first:mt-0 mt-3">{children}</h3>
  ),
  ul: ({ children }) => (
    <ul className="list-disc pl-5 space-y-1 m-0 first:mt-0 mt-3">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="list-decimal pl-5 space-y-1 m-0 first:mt-0 mt-3">{children}</ol>
  ),
  li: ({ children }) => <li className="m-0">{children}</li>,
  blockquote: ({ children }) => (
    <blockquote className="border-l-2 border-border pl-3 text-muted-foreground m-0 first:mt-0 mt-3">
      {children}
    </blockquote>
  ),
  strong: ({ children }) => (
    <strong className="font-semibold text-foreground">{children}</strong>
  ),
  em: ({ children }) => <em className="italic">{children}</em>,
  code: ({ className, children }) => {
    const isBlock = typeof className === "string" && className.includes("language-");

    if (isBlock) {
      return (
        <code className="font-mono text-[0.85em] text-foreground">
          {children}
        </code>
      );
    }

    return (
      <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-[0.85em] text-foreground">
        {children}
      </code>
    );
  },
  pre: ({ children }) => (
    <pre className="overflow-x-auto rounded-xl bg-muted/60 p-3 font-mono text-xs leading-6 m-0 first:mt-0 mt-3">
      {children}
    </pre>
  ),
  table: ({ children }) => (
    <div className="overflow-x-auto first:mt-0 mt-3">
      <table className="w-full border-collapse text-sm">{children}</table>
    </div>
  ),
  th: ({ children }) => (
    <th className="border border-border bg-muted px-2 py-1 text-left font-semibold">
      {children}
    </th>
  ),
  td: ({ children }) => (
    <td className="border border-border px-2 py-1 align-top">{children}</td>
  ),
  hr: () => <hr className="my-3 border-border" />,
  a: ({ children, href }) => (
    <a
      className="text-primary underline underline-offset-2"
      href={href}
      rel="noreferrer"
      target={href?.startsWith("http") ? "_blank" : undefined}
    >
      {children}
    </a>
  ),
};
