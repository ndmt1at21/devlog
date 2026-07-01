"use client";

import { track } from "@/lib/analytics";
import { useT } from "@/lib/i18n/provider";

const ALL = "Tất cả";

export function CategoryFilter({
  categories,
  active,
  onSelect,
}: {
  categories: string[];
  active: string;
  onSelect: (cat: string) => void;
}) {
  const t = useT();
  return (
    <div className="mx-auto mt-11 max-w-[1120px] px-6">
      <div
        role="group"
        aria-label={t("common.allCategories")}
        className="flex flex-wrap gap-[9px]"
      >
        {categories.map((cat) => {
          const isActive = cat === active;
          return (
            <button
              key={cat}
              onClick={() => {
                onSelect(cat);
                track("select_category", { category: cat });
              }}
              aria-pressed={isActive}
              className="cursor-pointer rounded-full px-4 py-2 text-[14px] font-semibold transition-[filter] hover:brightness-[.98]"
              style={
                isActive
                  ? {
                      background: "var(--pill-bg)",
                      color: "var(--pill-text)",
                      border: "1px solid var(--pill-bg)",
                    }
                  : {
                      background: "var(--surface)",
                      color: "var(--c43)",
                      border: "1px solid var(--border-2)",
                    }
              }
            >
              {cat === ALL ? t("common.allCategories") : cat}
            </button>
          );
        })}
      </div>
    </div>
  );
}
