// Server-side syntax highlighting with Shiki (VS Code-quality themes). Runs only
// on the server when rendering article pages, so it ships zero highlighting JS to
// the browser. A single cached highlighter is reused across requests.
import "server-only";
import { createHighlighter, type Highlighter } from "shiki";

const THEME = "one-dark-pro";

// Languages present in the content plus common extras. Unknown langs fall back
// to plaintext, so this list just needs to cover what we actually render.
const LANGS = [
  "javascript",
  "typescript",
  "jsx",
  "tsx",
  "json",
  "bash",
  "dockerfile",
  "go",
  "python",
  "css",
  "html",
  "yaml",
  "sql",
  "text",
];

const ALIAS: Record<string, string> = {
  js: "javascript",
  ts: "typescript",
  sh: "bash",
  shell: "bash",
  zsh: "bash",
  yml: "yaml",
  plaintext: "text",
  "": "text",
};

let highlighterPromise: Promise<Highlighter> | null = null;

function getHighlighter(): Promise<Highlighter> {
  if (!highlighterPromise) {
    highlighterPromise = createHighlighter({ themes: [THEME], langs: LANGS });
  }
  return highlighterPromise;
}

/** Returns Shiki `<pre class="shiki">…</pre>` HTML for a code block. */
export async function highlightCode(code: string, lang?: string): Promise<string> {
  const hl = await getHighlighter();
  let language = (lang ?? "text").toLowerCase();
  language = ALIAS[language] ?? language;
  if (!hl.getLoadedLanguages().includes(language)) language = "text";
  return hl.codeToHtml(code, { lang: language, theme: THEME });
}
