"use client";

import { useState } from "react";
import { api, ApiError } from "@/lib/api";
import type { Comment } from "@/lib/types";
import { useT } from "@/lib/i18n/provider";

export function Comments({
  slug,
  initialComments,
}: {
  slug: string;
  initialComments: Comment[];
}) {
  const t = useT();
  const [comments, setComments] = useState<Comment[]>(initialComments);
  const [name, setName] = useState("");
  const [text, setText] = useState("");
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!name.trim()) {
      setError(t("comments.errName"));
      return;
    }
    if (!text.trim()) {
      setError(t("comments.errText"));
      return;
    }
    setBusy(true);
    try {
      const created = await api.createComment(slug, {
        name: name.trim(),
        text: text.trim(),
      });
      setComments((prev) => [created, ...prev]);
      setText("");
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("comments.errFailed"));
    } finally {
      setBusy(false);
    }
  };

  return (
    <section className="mt-[54px] border-t border-border pt-[38px]">
      <h3 className="m-0 mb-1.5 text-[22px] font-bold tracking-[-.02em] text-text">
        {t("comments.title", { count: comments.length })}
      </h3>
      <p className="m-0 mb-6 text-[14px] text-subtle">
        {t("comments.subtitle")}
      </p>

      <form
        onSubmit={submit}
        className="mb-[30px] rounded-[14px] border border-border bg-surface p-5"
      >
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder={t("comments.namePlaceholder")}
          aria-label={t("comments.namePlaceholder")}
          className="field mb-[11px] px-[14px] py-[11px] text-[14.5px]"
        />
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder={t("comments.textPlaceholder")}
          aria-label={t("comments.textPlaceholder")}
          rows={3}
          className="field resize-y px-[14px] py-[11px] text-[14.5px] leading-[1.55]"
        />
        <div className="mt-[13px] flex items-center justify-between gap-3.5">
          <span role="alert" className="text-[13px] font-medium text-danger">
            {error}
          </span>
          <button
            type="submit"
            disabled={busy}
            className="btn-accent whitespace-nowrap rounded-[9px] px-[22px] py-2.5 text-[14.5px] font-semibold disabled:opacity-60"
          >
            {busy ? t("comments.submitting") : t("comments.submit")}
          </button>
        </div>
      </form>

      <div className="flex flex-col gap-[22px]">
        {comments.map((c, i) => (
          <div key={`${c.name}-${i}`} className="flex gap-[13px]">
            <span
              aria-hidden="true"
              className="avatar-accent h-[38px] w-[38px] flex-none text-[15px]"
            >
              {c.initial}
            </span>
            <div className="flex-1">
              <div className="mb-1 flex items-baseline gap-[9px]">
                <span className="text-[14.5px] font-semibold text-text">
                  {c.name}
                </span>
                <span className="text-[12.5px] text-faint">{c.time}</span>
              </div>
              <p className="m-0 text-[14.5px] leading-[1.6] text-strong">
                {c.text}
              </p>
            </div>
          </div>
        ))}
        {comments.length === 0 && (
          <div className="py-[18px] text-center text-[14px] text-faint">
            {t("comments.empty")}
          </div>
        )}
      </div>
    </section>
  );
}
