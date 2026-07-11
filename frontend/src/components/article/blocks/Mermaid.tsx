"use client";

import { useEffect, useId, useRef, useState } from "react";
import { useTheme } from "@/components/theme/ThemeProvider";

// Renders a Mermaid diagram from an article `code` block whose language is
// "mermaid". Mermaid needs the DOM, so it runs client-only via a dynamic
// import: no Mermaid JS ships to pages without a diagram, and nothing executes
// on the Workers SSR runtime (where the code path that highlights code blocks
// skips lang="mermaid"). On any parse/render error we fall back to the raw
// source so the content is never lost.
export function Mermaid({ code }: { code: string }) {
  const baseId = useId().replace(/[^a-zA-Z0-9]/g, "");
  const seq = useRef(0);
  const { theme } = useTheme();
  const [svg, setSvg] = useState("");
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    // A fresh id per attempt: React strict-mode double-invokes effects, and
    // Mermaid keys its temp render node by id — reusing one can collide.
    const renderId = `mmd-${baseId}-${seq.current++}`;
    (async () => {
      try {
        const mermaid = (await import("mermaid")).default;
        mermaid.initialize({
          startOnLoad: false,
          securityLevel: "strict",
          theme: theme === "dark" ? "dark" : "default",
          fontFamily: "inherit",
        });
        const { svg } = await mermaid.render(renderId, code);
        if (!cancelled) {
          setSvg(svg);
          setError(false);
        }
      } catch {
        if (!cancelled) {
          setError(true);
          setSvg("");
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [code, theme, baseId]);

  if (error) {
    return (
      <pre className="my-[26px] overflow-x-auto rounded-xl border border-border bg-surface px-[18px] py-4 font-mono text-[13px] leading-[1.7] text-subtle [tab-size:2]">
        {code}
      </pre>
    );
  }

  // securityLevel "strict" sanitizes the output (HTML labels off, scripts
  // stripped), so injecting the generated SVG is safe. Empty on first paint
  // and during SSR — the effect fills it after hydration.
  return (
    <figure
      className="my-[30px] flex justify-center overflow-x-auto rounded-[14px] border border-border bg-surface px-[18px] py-[26px] [&_svg]:h-auto [&_svg]:max-w-full"
      aria-label="diagram"
      dangerouslySetInnerHTML={{ __html: svg }}
    />
  );
}
