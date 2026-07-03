import type { ReactNode } from "react";

// Inline markdown spans preserved in block text by the backend converter:
// **bold**, *italic*, `code`, [label](url). Rendered as React elements — never
// via dangerouslySetInnerHTML — so user content cannot inject markup. Nesting
// is intentionally not supported (mirrors backend/internal/content).
const INLINE_RE =
  /\*\*([^*]+)\*\*|\*([^*\s][^*]*)\*|`([^`]+)`|\[([^\]]+)\]\(([^)\s]+)\)/g;

// Allowed link targets; anything else (javascript:, data:, …) renders as text.
const SAFE_URL_RE = /^(https?:\/\/|mailto:|\/|#)/i;

/** Render text with inline markdown spans as React nodes (XSS-safe). */
export function renderInline(text: string): ReactNode {
  if (!/[*`[]/.test(text)) return text;

  const out: ReactNode[] = [];
  let last = 0;
  let key = 0;
  for (const m of text.matchAll(INLINE_RE)) {
    const i = m.index ?? 0;
    if (i > last) out.push(text.slice(last, i));

    if (m[1] != null) {
      out.push(<strong key={key++}>{m[1]}</strong>);
    } else if (m[2] != null) {
      out.push(<em key={key++}>{m[2]}</em>);
    } else if (m[3] != null) {
      out.push(
        <code
          key={key++}
          className="rounded-md bg-hoverbg px-1.5 py-0.5 font-mono text-[0.85em] text-strong"
        >
          {m[3]}
        </code>,
      );
    } else if (m[4] != null && m[5] != null) {
      if (SAFE_URL_RE.test(m[5])) {
        const external = /^https?:\/\//i.test(m[5]);
        out.push(
          <a
            key={key++}
            href={m[5]}
            className="font-medium text-accent-ink underline underline-offset-2 hover:opacity-75"
            {...(external ? { target: "_blank", rel: "noopener noreferrer" } : {})}
          >
            {m[4]}
          </a>,
        );
      } else {
        out.push(m[4]);
      }
    }
    last = i + m[0].length;
  }
  if (last < text.length) out.push(text.slice(last));
  return out;
}
