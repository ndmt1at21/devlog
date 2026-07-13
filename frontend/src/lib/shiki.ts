// Server-side syntax highlighting with Shiki's FINE-GRAINED core. The full `shiki`
// bundle inlines every grammar (~200 languages) into one file — on Cloudflare
// Workers that made handler.mjs ~15 MB. Here we build a highlighter from just the
// languages we render, so only those grammars ship. Runs only on the server
// (zero highlighting JS to the browser); a single highlighter is cached across
// requests.
//
// Engine: the pure-JS RegExp engine, NOT the Oniguruma WASM one. The Workers
// runtime (workerd) forbids compiling WebAssembly from a byte buffer at request
// time — that's a hard "code generation disallowed by embedder" trap. Shiki's
// `wasm-inlined` engine does exactly that (`WebAssembly.instantiate(bytes)` per
// request), so it throws on Cloudflare. The JS engine uses native RegExp with no
// WASM at all and is Shiki's recommended engine for Workers.
import "server-only";
import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import { createJavaScriptRegexEngine } from "@shikijs/engine-javascript";

const THEME = "one-dark-pro";

// Short forms like js/ts/py/cjs/mjs are already registered as aliases by the
// grammars themselves; this just covers the extras we relied on before. Anything
// not bundled (bash, json, css, yaml, sql, …) falls back to plaintext.
const ALIAS: Record<string, string> = {
  golang: "go",
  plaintext: "text",
  txt: "text",
  "": "text",
};

let highlighterPromise: Promise<HighlighterCore> | null = null;

function getHighlighter(): Promise<HighlighterCore> {
  if (!highlighterPromise) {
    highlighterPromise = createHighlighterCore({
      themes: [import("@shikijs/themes/one-dark-pro")],
      // Only these grammars are bundled into the Worker. To support another
      // language, add one more `import("@shikijs/langs/<name>")` — each grammar
      // is only a few KB.
      langs: [
        import("@shikijs/langs/javascript"),
        import("@shikijs/langs/typescript"),
        import("@shikijs/langs/go"),
        import("@shikijs/langs/python"),
        import("@shikijs/langs/css"),
        import("@shikijs/langs/html"),
      ],
      // Pure-JS RegExp engine (no WASM). `forgiving` degrades gracefully: a
      // grammar pattern the JS engine can't translate is skipped instead of
      // throwing mid-request, so a rare unsupported construct can't 500 an
      // article page.
      engine: createJavaScriptRegexEngine({ forgiving: true }),
    });
  }
  return highlighterPromise;
}

/**
 * Returns Shiki `<pre class="shiki">…</pre>` HTML for a code block. Unknown or
 * unbundled languages fall back to plaintext.
 */
export async function highlightCode(
  code: string,
  lang?: string,
): Promise<string> {
  const hl = await getHighlighter();
  let language = (lang ?? "text").toLowerCase();
  language = ALIAS[language] ?? language;
  // "text" is a built-in plain grammar and isn't listed in getLoadedLanguages().
  if (language !== "text" && !hl.getLoadedLanguages().includes(language)) {
    language = "text";
  }
  return hl.codeToHtml(code, { lang: language, theme: THEME });
}
