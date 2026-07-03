"use client";

import type { Block } from "@/lib/types";
import { CodeBlock } from "./blocks/CodeBlock";
import { Diagram } from "./blocks/Diagram";
import { renderInline } from "./inline";

// BlockView renders one article body block. Shared by the article page and the
// editor preview; text-bearing blocks pass through the safe inline renderer.
export function BlockView({ block, slug }: { block: Block; slug: string }) {
  switch (block.type) {
    case "h":
      return (
        <h2 className="mb-2.5 mt-11 text-[25px] font-bold leading-[1.3] tracking-[-.02em] text-text">
          {renderInline(block.text ?? "")}
        </h2>
      );
    case "quote":
      return (
        <blockquote className="my-[30px] border-l-[3px] border-accent py-1.5 pl-[22px] text-[20px] font-medium leading-[1.6] text-c3a">
          {renderInline(block.text ?? "")}
        </blockquote>
      );
    case "code":
      return (
        <CodeBlock
          lang={block.lang}
          code={block.code ?? ""}
          html={block.html}
          slug={slug}
        />
      );
    case "diagram":
      return <Diagram steps={block.steps ?? []} caption={block.caption} />;
    case "list": {
      const ListTag = block.ordered ? "ol" : "ul";
      return (
        <ListTag
          className={`mb-[22px] pl-[26px] ${block.ordered ? "list-decimal" : "list-disc"}`}
        >
          {(block.items ?? []).map((item, i) => (
            <li key={i} className="mb-1.5 pl-1">
              {renderInline(item)}
            </li>
          ))}
        </ListTag>
      );
    }
    case "p":
    default:
      return <p className="mb-[22px]">{renderInline(block.text ?? "")}</p>;
  }
}
