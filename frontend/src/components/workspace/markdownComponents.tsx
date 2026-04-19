import type { Components } from "react-markdown";

export const markdownComponents: Components = {
  p: ({ children }) => <p className="m-0 mt-3 leading-7 first:mt-0">{children}</p>,
  h1: ({ children }) => (
    <h1 className="m-0 mt-6 text-3xl font-bold tracking-tight text-foreground first:mt-0">
      {children}
    </h1>
  ),
  h2: ({ children }) => (
    <h2 className="m-0 mt-5 text-2xl font-bold tracking-tight text-foreground first:mt-0">
      {children}
    </h2>
  ),
  h3: ({ children }) => (
    <h3 className="m-0 mt-4 text-lg font-semibold tracking-tight text-foreground first:mt-0">
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
    <blockquote className="m-0 mt-4 border-l-4 border-primary/30 pl-4 italic text-muted-foreground first:mt-0">
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
    <pre className="m-0 mt-4 overflow-x-auto rounded-xl bg-muted/60 p-4 font-mono text-xs leading-6 first:mt-0">
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
