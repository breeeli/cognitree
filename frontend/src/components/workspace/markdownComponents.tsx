import type { Components } from "react-markdown";

export const markdownComponents: Components = {
  p: ({ children }) => <p className="m-0 first:mt-0 mt-3">{children}</p>,
  h1: ({ children }) => (
    <h1 className="text-lg font-semibold tracking-tight m-0 first:mt-0 mt-4">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="text-base font-semibold tracking-tight m-0 first:mt-0 mt-4">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="text-sm font-semibold tracking-tight m-0 first:mt-0 mt-3">
      {children}
    </h3>
  ),
  ul: ({ children }) => (
    <ul className="list-disc pl-5 space-y-1 m-0 first:mt-0 mt-3">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="list-decimal pl-5 space-y-1 m-0 first:mt-0 mt-3">
      {children}
    </ol>
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
      return <code className="font-mono text-[0.85em] text-foreground">{children}</code>;
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
