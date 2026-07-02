"use client";

import { useRouter } from "next/navigation";
import { track } from "@/lib/analytics";

export function TagList({ tags }: { tags: string[] }) {
  const router = useRouter();
  if (!tags.length) return null;

  return (
    <div className="mb-[22px] flex flex-wrap gap-2">
      {tags.map((tag) => (
        <button
          key={tag}
          onClick={() => {
            track("select_tag", { tag });
            router.push(`/?q=${encodeURIComponent(tag)}`);
          }}
          className="cursor-pointer rounded-full px-3 py-[5px] text-[13px] font-semibold text-accent-ink transition-all"
          style={{
            background: "color-mix(in srgb, var(--accent) 10%, transparent)",
            border: "1px solid color-mix(in srgb, var(--accent) 22%, transparent)",
          }}
        >
          #{tag}
        </button>
      ))}
    </div>
  );
}
