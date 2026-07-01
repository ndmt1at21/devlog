import { Fragment } from "react";

export function Diagram({
  steps,
  caption,
}: {
  steps: string[];
  caption?: string;
}) {
  return (
    <figure className="my-[30px] rounded-[14px] border border-border bg-surface px-[18px] pb-5 pt-[26px]">
      <div className="flex flex-wrap items-center justify-center gap-2">
        {steps.map((label, i) => (
          <Fragment key={i}>
            {i > 0 && (
              <span className="text-[18px] font-semibold text-[color:var(--cc4)]">
                →
              </span>
            )}
            <span
              className="whitespace-nowrap rounded-[10px] px-[17px] py-[11px] text-[14.5px] font-semibold text-accent-ink"
              style={{
                background: "color-mix(in srgb, var(--accent) 11%, transparent)",
                border:
                  "1px solid color-mix(in srgb, var(--accent) 24%, transparent)",
              }}
            >
              {label}
            </span>
          </Fragment>
        ))}
      </div>
      {caption && (
        <figcaption className="mt-[18px] text-center text-[13px] leading-[1.5] text-subtle">
          {caption}
        </figcaption>
      )}
    </figure>
  );
}
