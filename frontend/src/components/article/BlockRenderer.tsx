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
        <h2 className="mb-2.5 mt-11 text-[25px] font-bold leading-[1.3] tracking-[-.02em] text-balance text-text">
          {renderInline(block.text ?? "")}
        </h2>
      );
    case "quote":
      return (
        <blockquote className="my-[30px] border-l-[3px] border-accent py-1.5 pl-[22px] text-[20px] font-medium leading-[1.6] text-pretty text-c3a">
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
    case "img":
      return (
        <figure className="my-[30px]">
          {/* Plain <img>: bytes come final-form from the image CDN and
              next/image optimization isn't wired up on the Workers runtime. */}
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img
            src={block.src}
            alt={block.alt ?? ""}
            loading="lazy"
            decoding="async"
            className="w-full rounded-[14px] border border-border"
          />
          {block.caption && (
            <figcaption className="mt-2.5 text-center text-[13.5px] text-muted">
              {block.caption}
            </figcaption>
          )}
        </figure>
      );
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
      // Justified from sm: up only — on narrow phones (~7 syllables/line)
      // justify gaps get too wide without Vietnamese hyphenation, so mobile
      // stays left-aligned. The last line stays left (browser default).
      return (
        <p className="mb-[22px] text-pretty sm:text-justify">
          {renderInline(block.text ?? "")}
        </p>
      );
  }
}
