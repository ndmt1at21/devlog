import type { Block } from "./types";

// Client-side mirror of the backend converter (backend/internal/content) used
// only for the editor's live preview — the server conversion is authoritative.
const HEADING_RE = /^(#{1,6})\s+(.*)$/;
const ORDERED_RE = /^\d{1,4}[.)]\s+(.*)$/;
const HRULE_RE = /^(-{3,}|\*{3,}|_{3,})$/;

function unorderedItem(line: string): string {
  for (const marker of ["- ", "* ", "+ "]) {
    if (line.startsWith(marker)) return line.slice(marker.length).trim();
  }
  return "";
}

/** Convert README-style markdown into article body blocks (preview only). */
export function blocksFromMarkdown(src: string): Block[] {
  const lines = src.replaceAll("\r\n", "\n").split("\n");
  const blocks: Block[] = [];
  let para: string[] = [];

  const flush = () => {
    if (para.length > 0) {
      blocks.push({ type: "p", text: para.join(" ") });
      para = [];
    }
  };

  for (let i = 0; i < lines.length; ) {
    const trimmed = lines[i].trim();

    if (trimmed === "" || HRULE_RE.test(trimmed)) {
      flush();
      i++;
    } else if (trimmed.startsWith("```")) {
      flush();
      const lang = trimmed.slice(3).trim();
      i++;
      const code: string[] = [];
      while (i < lines.length && !lines[i].trim().startsWith("```")) {
        code.push(lines[i]);
        i++;
      }
      if (i < lines.length) i++; // closing fence
      blocks.push({
        type: "code",
        lang,
        code: code.join("\n").replace(/\n+$/, ""),
      });
    } else if (HEADING_RE.test(trimmed)) {
      flush();
      blocks.push({ type: "h", text: HEADING_RE.exec(trimmed)![2].trim() });
      i++;
    } else if (trimmed.startsWith(">")) {
      flush();
      const quote: string[] = [];
      while (i < lines.length && lines[i].trim().startsWith(">")) {
        const line = lines[i].trim().replace(/^>\s?/, "").trim();
        if (line !== "") quote.push(line);
        i++;
      }
      blocks.push({ type: "quote", text: quote.join(" ") });
    } else if (unorderedItem(trimmed) !== "") {
      flush();
      const items: string[] = [];
      while (i < lines.length && unorderedItem(lines[i].trim()) !== "") {
        items.push(unorderedItem(lines[i].trim()));
        i++;
      }
      blocks.push({ type: "list", items });
    } else if (ORDERED_RE.test(trimmed)) {
      flush();
      const items: string[] = [];
      while (i < lines.length && ORDERED_RE.test(lines[i].trim())) {
        items.push(ORDERED_RE.exec(lines[i].trim())![1].trim());
        i++;
      }
      blocks.push({ type: "list", items, ordered: true });
    } else {
      para.push(trimmed);
      i++;
    }
  }
  flush();
  return blocks;
}
