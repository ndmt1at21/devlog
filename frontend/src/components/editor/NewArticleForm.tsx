"use client";

// New-article editor. Two authoring modes normalize to the same block model
// the backend stores: "markdown" sends raw README-style source (converted
// server-side), "blocks" sends the structured rich-text editor output. The
// page is only a convenience for holders of the IAM permission
// `articles:create` — the backend re-checks it on every POST.

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useMemo, useRef, useState } from "react";
import { api, ApiError } from "@/lib/api";
import { useAuth } from "@/lib/auth";
import { useT } from "@/lib/i18n/provider";
import {
  DEFAULT_LOCALE,
  LOCALES,
  LOCALE_LABELS,
  type Locale,
} from "@/lib/i18n/dictionaries";
import { track } from "@/lib/analytics";
import { blocksFromMarkdown } from "@/lib/markdown";
import {
  IMAGE_ACCEPT,
  IMAGE_TYPES,
  MAX_IMAGE_BYTES,
  baseName,
  uploadImage,
} from "@/lib/uploads";
import type {
  ArticleDetail,
  Block,
  LocalizedInput,
  NewArticleInput,
} from "@/lib/types";
import { BlockView } from "@/components/article/BlockRenderer";

type Mode = "markdown" | "blocks";
type EditorBlockType = "p" | "h" | "quote" | "code" | "diagram" | "list" | "img";

// One row of the rich-text editor. `lines` holds diagram steps / list items,
// one per line; only the fields relevant to `type` are sent.
interface EditorBlock {
  id: number;
  type: EditorBlockType;
  text: string;
  lang: string;
  code: string;
  caption: string;
  lines: string;
  ordered: boolean;
  src: string;
  alt: string;
  /** True while this row's image is being uploaded to the bucket. */
  uploading: boolean;
}

let nextId = 1;
const newBlock = (type: EditorBlockType = "p"): EditorBlock => ({
  id: nextId++,
  type,
  text: "",
  lang: "",
  code: "",
  caption: "",
  lines: "",
  ordered: false,
  src: "",
  alt: "",
  uploading: false,
});

const splitLines = (s: string) =>
  s
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean);

// blocksToRows seeds the rich-text editor from a stored article body (edit
// mode). Types the block editor can't represent (e.g. the synthesized "ad")
// are skipped; an empty result falls back to one blank row.
function blocksToRows(blocks: Block[]): EditorBlock[] {
  const rows: EditorBlock[] = [];
  for (const b of blocks) {
    switch (b.type) {
      case "code":
        rows.push({ ...newBlock("code"), lang: b.lang ?? "", code: b.code ?? "" });
        break;
      case "img":
        rows.push({
          ...newBlock("img"),
          src: b.src ?? "",
          alt: b.alt ?? "",
          caption: b.caption ?? "",
        });
        break;
      case "diagram":
        rows.push({
          ...newBlock("diagram"),
          lines: (b.steps ?? []).join("\n"),
          caption: b.caption ?? "",
        });
        break;
      case "list":
        rows.push({
          ...newBlock("list"),
          lines: (b.items ?? []).join("\n"),
          ordered: !!b.ordered,
        });
        break;
      case "p":
      case "h":
      case "quote":
        rows.push({ ...newBlock(b.type), text: b.text ?? "" });
        break;
    }
  }
  return rows.length > 0 ? rows : [newBlock()];
}

function toBlocks(rows: EditorBlock[]): Block[] {
  const out: Block[] = [];
  for (const b of rows) {
    switch (b.type) {
      case "code":
        if (b.code.trim())
          out.push({ type: "code", lang: b.lang.trim(), code: b.code });
        break;
      case "img":
        if (b.src)
          out.push({
            type: "img",
            src: b.src,
            alt: b.alt.trim(),
            caption: b.caption.trim(),
          });
        break;
      case "diagram": {
        const steps = splitLines(b.lines);
        if (steps.length > 0)
          out.push({ type: "diagram", steps, caption: b.caption.trim() });
        break;
      }
      case "list": {
        const items = splitLines(b.lines);
        if (items.length > 0)
          out.push({ type: "list", items, ordered: b.ordered });
        break;
      }
      default:
        if (b.text.trim()) out.push({ type: b.type, text: b.text.trim() });
    }
  }
  return out;
}

// LangState is one language's editable content. Category, tags and the cover
// image live outside this (shared across languages); coverAlt is per-language
// since the alt text is translated even though the image is the same.
interface LangState {
  title: string;
  excerpt: string;
  coverAlt: string;
  mode: Mode;
  markdown: string;
  rows: EditorBlock[];
}

const emptyLang = (): LangState => ({
  title: "",
  excerpt: "",
  coverAlt: "",
  mode: "markdown",
  markdown: "",
  rows: [newBlock()],
});

// initLangs seeds per-language editor state. In create mode every language
// starts empty (markdown authoring). In edit mode the primary language is seeded
// from the article's base fields and each translation from article.translations;
// a language with stored blocks starts in "blocks" mode (markdown isn't
// losslessly reconstructable), a still-untranslated language stays empty.
function initLangs(article?: ArticleDetail): Record<Locale, LangState> {
  const out = {} as Record<Locale, LangState>;
  for (const loc of LOCALES) out[loc] = emptyLang();
  if (!article) return out;

  const primary = LOCALES.includes(article.lang as Locale)
    ? (article.lang as Locale)
    : DEFAULT_LOCALE;
  out[primary] = {
    title: article.title,
    excerpt: article.excerpt,
    coverAlt: article.coverAlt ?? "",
    mode: "blocks",
    markdown: "",
    rows: blocksToRows(article.body),
  };
  for (const loc of LOCALES) {
    const tr = article.translations?.[loc];
    if (tr && loc !== primary) {
      out[loc] = {
        title: tr.title,
        excerpt: tr.excerpt,
        coverAlt: tr.coverAlt ?? "",
        mode: "blocks",
        markdown: "",
        rows: blocksToRows(tr.body),
      };
    }
  }
  return out;
}

const fieldCls = "field px-[14px] py-3 text-[15px]";
const labelCls = "mb-[7px] block text-[13.5px] font-semibold text-strong";

// NewArticleForm doubles as the edit form: pass `article` to prefill the fields
// and switch the submit to PUT. In edit mode the block editor is seeded from the
// stored body (markdown can't be reconstructed losslessly), so the mode starts
// on "blocks".
export function NewArticleForm({ article }: { article?: ArticleDetail } = {}) {
  const router = useRouter();
  const { user, loading } = useAuth();
  const t = useT();
  const editing = !!article;

  // Per-language content; the currently edited language is `activeLang`.
  const [langs, setLangs] = useState<Record<Locale, LangState>>(() =>
    initLangs(article),
  );
  const [activeLang, setActiveLang] = useState<Locale>(() =>
    article && LOCALES.includes(article.lang as Locale)
      ? (article.lang as Locale)
      : DEFAULT_LOCALE,
  );

  // Shared across languages.
  const [category, setCategory] = useState(article?.category ?? "");
  const [tagsRaw, setTagsRaw] = useState(article ? article.tags.join(", ") : "");
  const [cover, setCover] = useState(article?.cover ?? "");
  const [coverUploading, setCoverUploading] = useState(false);

  const [preview, setPreview] = useState(false);
  const [error, setError] = useState("");
  const [busy, setBusy] = useState(false);
  // In-flight body-image uploads: markdown toolbar vs. blocks "insert image".
  const [mdUploading, setMdUploading] = useState(false);
  const [insertingImage, setInsertingImage] = useState(false);

  // Last-focused text field, so the inline-format toolbar knows its target.
  const activeField = useRef<HTMLTextAreaElement | null>(null);

  const previewBlocks = useMemo<Block[]>(() => {
    const s = langs[activeLang];
    return s.mode === "markdown"
      ? blocksFromMarkdown(s.markdown)
      : toBlocks(s.rows);
  }, [langs, activeLang]);

  if (!user) {
    if (loading) {
      return (
        <div
          className="mx-auto max-w-[760px] px-6 pb-24 pt-16 text-center text-muted"
          aria-busy="true"
        >
          …
        </div>
      );
    }
    return (
      <Notice title={t("editor.needLoginTitle")} body={t("editor.needLogin")}>
        <Link href="/login" className="btn-accent inline-block px-6 py-3 text-[15px] no-underline">
          {t("editor.loginCta")}
        </Link>
      </Notice>
    );
  }
  if (!user.canWrite) {
    return (
      <Notice
        title={t("editor.noPermissionTitle")}
        body={t("editor.noPermission")}
      />
    );
  }
  // Only the author may edit their own article (server-computed by user id).
  // The backend re-checks on PUT; this guard just avoids showing a form that
  // can't be saved.
  if (editing && article && !article.editable) {
    return (
      <Notice title={t("editor.notAuthorTitle")} body={t("editor.notAuthor")} />
    );
  }

  // Accessors bound to the active language, so the field/block JSX below stays
  // language-agnostic. Setters patch only the active language's slice.
  const cur = langs[activeLang];
  const title = cur.title;
  const excerpt = cur.excerpt;
  const coverAlt = cur.coverAlt;
  const mode = cur.mode;
  const markdown = cur.markdown;
  const rows = cur.rows;

  const patchLang = (patch: Partial<LangState>) =>
    setLangs((ls) => ({ ...ls, [activeLang]: { ...ls[activeLang], ...patch } }));
  const setTitle = (v: string) => patchLang({ title: v });
  const setExcerpt = (v: string) => patchLang({ excerpt: v });
  const setCoverAlt = (v: string) => patchLang({ coverAlt: v });
  const setMode = (v: Mode) => patchLang({ mode: v });
  const setMarkdown = (v: string | ((s: string) => string)) =>
    setLangs((ls) => {
      const prev = ls[activeLang].markdown;
      const next = typeof v === "function" ? v(prev) : v;
      return { ...ls, [activeLang]: { ...ls[activeLang], markdown: next } };
    });
  const setRows = (v: EditorBlock[] | ((rs: EditorBlock[]) => EditorBlock[])) =>
    setLangs((ls) => {
      const prev = ls[activeLang].rows;
      const next = typeof v === "function" ? v(prev) : v;
      return { ...ls, [activeLang]: { ...ls[activeLang], rows: next } };
    });

  // Whether a language has any authored content (drives the tab indicator and
  // "publish now, translate later": empty languages are simply omitted).
  const langHasContent = (loc: Locale) => {
    const s = langs[loc];
    if (s.title.trim()) return true;
    return s.mode === "markdown"
      ? s.markdown.trim() !== ""
      : toBlocks(s.rows).length > 0;
  };

  const update = (id: number, patch: Partial<EditorBlock>) =>
    setRows((rs) => rs.map((r) => (r.id === id ? { ...r, ...patch } : r)));

  const move = (id: number, dir: -1 | 1) =>
    setRows((rs) => {
      const i = rs.findIndex((r) => r.id === id);
      const j = i + dir;
      if (i < 0 || j < 0 || j >= rs.length) return rs;
      const next = [...rs];
      [next[i], next[j]] = [next[j], next[i]];
      return next;
    });

  const remove = (id: number) =>
    setRows((rs) => (rs.length > 1 ? rs.filter((r) => r.id !== id) : rs));

  // Client-side mirror of the backend's upload limits; returns "" when valid.
  const validateImage = (file: File) => {
    if (!IMAGE_TYPES.includes(file.type)) return t("editor.errImageType");
    if (file.size > MAX_IMAGE_BYTES) return t("editor.errImageSize");
    return "";
  };

  const uploadError = (err: unknown) =>
    setError(err instanceof ApiError ? err.message : t("editor.errUpload"));

  // Markdown mode: upload, then splice a standalone ![alt](url) line at the
  // caret of the markdown field (or append when it wasn't focused).
  const insertMarkdownImage = async (file: File) => {
    const invalid = validateImage(file);
    if (invalid) return setError(invalid);
    setError("");
    setMdUploading(true);
    try {
      const url = await uploadImage(file);
      const snippet = `![${baseName(file.name)}](${url})`;
      const el = activeField.current;
      if (el && el.dataset.target === "md") {
        const start = el.selectionStart ?? el.value.length;
        const end = el.selectionEnd ?? start;
        const next = `${el.value.slice(0, start)}\n\n${snippet}\n\n${el.value.slice(end)}`;
        setMarkdown(next);
        const caret = start + snippet.length + 4;
        requestAnimationFrame(() => {
          el.focus();
          el.setSelectionRange(caret, caret);
        });
      } else {
        setMarkdown((md) => (md.trim() ? `${md.trimEnd()}\n\n${snippet}\n` : `${snippet}\n`));
      }
    } catch (err) {
      uploadError(err);
    } finally {
      setMdUploading(false);
    }
  };

  // Blocks mode: upload for one img row, defaulting alt to the file name.
  const uploadRowImage = async (id: number, file: File) => {
    const invalid = validateImage(file);
    if (invalid) return setError(invalid);
    setError("");
    update(id, { uploading: true });
    try {
      const url = await uploadImage(file);
      setRows((rs) =>
        rs.map((r) =>
          r.id === id
            ? { ...r, uploading: false, src: url, alt: r.alt || baseName(file.name) }
            : r,
        ),
      );
    } catch (err) {
      update(id, { uploading: false });
      uploadError(err);
    }
  };

  // Blocks mode: one-click "insert image" — upload, then append a new img block
  // (alt defaulted to the file name). The author can reorder it with ↑/↓.
  const insertImageBlock = async (file: File) => {
    const invalid = validateImage(file);
    if (invalid) return setError(invalid);
    setError("");
    setInsertingImage(true);
    try {
      const url = await uploadImage(file);
      setRows((rs) => [
        ...rs,
        { ...newBlock("img"), src: url, alt: baseName(file.name) },
      ]);
    } catch (err) {
      uploadError(err);
    } finally {
      setInsertingImage(false);
    }
  };

  // Wrap the selection of the last-focused textarea in inline markers. State is
  // updated through the field's data-target key so React stays the source of
  // truth; the DOM element is only read for the selection range.
  const applyInline = (before: string, after: string) => {
    const el = activeField.current;
    if (!el) return;
    const start = el.selectionStart ?? 0;
    const end = el.selectionEnd ?? 0;
    const selected = el.value.slice(start, end);
    const next =
      el.value.slice(0, start) + before + selected + after + el.value.slice(end);
    const target = el.dataset.target;
    if (target === "md") setMarkdown(next);
    else if (target) update(Number(target), { text: next });
    requestAnimationFrame(() => {
      el.focus();
      el.setSelectionRange(start + before.length, end + before.length);
    });
  };

  // Cover upload: same presigned direct-to-R2 flow as body images; stores the
  // returned public CDN URL.
  const uploadCover = async (file: File) => {
    const invalid = validateImage(file);
    if (invalid) return setError(invalid);
    setError("");
    setCoverUploading(true);
    try {
      setCover(await uploadImage(file));
    } catch (err) {
      uploadError(err);
    } finally {
      setCoverUploading(false);
    }
  };

  // buildLang turns one language's editor state into a payload entry. An empty
  // language yields null (skipped); a partially-filled one yields "incomplete"
  // so submit can flag it. The cover image is shared, so coverAlt only rides
  // along when a cover exists.
  const buildLang = (
    loc: Locale,
  ): "empty" | "incomplete" | LocalizedInput => {
    const s = langs[loc];
    const hasTitle = s.title.trim() !== "";
    const body = s.mode === "blocks" ? toBlocks(s.rows) : [];
    const hasBody =
      s.mode === "markdown" ? s.markdown.trim() !== "" : body.length > 0;
    if (!hasTitle && !hasBody) return "empty";
    if (!hasTitle || !hasBody) return "incomplete";
    const meta = {
      lang: loc,
      title: s.title.trim(),
      excerpt: s.excerpt.trim() || undefined,
      coverAlt: cover.trim() ? s.coverAlt.trim() || undefined : undefined,
    };
    return s.mode === "markdown"
      ? { ...meta, format: "markdown", content: s.markdown }
      : { ...meta, format: "blocks", body };
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    if (!category.trim()) return setError(t("editor.errCategory"));

    const complete: LocalizedInput[] = [];
    for (const loc of LOCALES) {
      const built = buildLang(loc);
      if (built === "empty") continue;
      if (built === "incomplete") {
        return setError(
          t("editor.errIncompleteLang", { lang: LOCALE_LABELS[loc] }),
        );
      }
      complete.push(built);
    }
    if (complete.length === 0) return setError(t("editor.errNoContent"));

    const tags = tagsRaw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean)
      .slice(0, 8);
    // The first complete language (vi preferred, per LOCALES order) is primary;
    // the rest are translations.
    const [primary, ...rest] = complete;
    const payload: NewArticleInput = {
      category: category.trim(),
      cover: cover.trim() || undefined,
      tags,
      ...primary,
      translations: rest.length > 0 ? rest : undefined,
    };

    setBusy(true);
    try {
      const saved =
        editing && article
          ? await api.updateArticle(article.slug, payload)
          : await api.createArticle(payload);
      track(editing ? "edit_article" : "create_article", {
        slug: saved.slug,
        format: primary.format,
        langs: complete.length,
      });
      router.push(`/articles/${saved.slug}`);
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("error.body"));
      setBusy(false);
    }
  };

  const focusProps = {
    onFocus: (e: React.FocusEvent<HTMLTextAreaElement>) => {
      activeField.current = e.currentTarget;
    },
  };

  return (
    <div className="mx-auto max-w-[760px] px-6 pb-24 pt-9">
      <h1 className="m-0 mb-2 text-[30px] font-extrabold tracking-[-.03em] text-text">
        {editing ? t("editor.editTitle") : t("editor.title")}
      </h1>
      <p className="m-0 mb-[26px] text-[15px] text-muted">
        {editing ? t("editor.editSub") : t("editor.sub")}
      </p>

      <form onSubmit={submit} className="rounded-2xl border border-border bg-surface p-7">
        {/* --- language tabs (title/excerpt/body/coverAlt are per-language) --- */}
        <div className="mb-5">
          <span className={labelCls}>{t("editor.langSection")}</span>
          <div
            role="tablist"
            aria-label={t("editor.langSection")}
            className="flex gap-1.5"
          >
            {LOCALES.map((loc) => {
              const filled = langHasContent(loc);
              const selected = activeLang === loc;
              return (
                <button
                  key={loc}
                  type="button"
                  role="tab"
                  aria-selected={selected}
                  onClick={() => {
                    // The previously focused textarea belongs to the outgoing
                    // language and is about to unmount; drop the ref so the
                    // inline toolbar can't act on a detached node.
                    activeField.current = null;
                    setActiveLang(loc);
                  }}
                  className={`flex cursor-pointer items-center gap-1.5 rounded-full border px-4 py-1.5 text-[13.5px] font-semibold transition-colors ${
                    selected
                      ? "border-accent-ink text-accent-ink"
                      : "border-border2 text-c5b hover:border-hover"
                  }`}
                >
                  {LOCALE_LABELS[loc]}
                  <span
                    title={
                      filled
                        ? t("editor.langHasContent")
                        : t("editor.langNoContent")
                    }
                    className={`h-1.5 w-1.5 rounded-full ${
                      filled ? "bg-accent-ink" : "bg-border2"
                    }`}
                    aria-hidden
                  />
                </button>
              );
            })}
          </div>
          <p className="mt-2 text-[12.5px] leading-[1.6] text-subtle">
            {t("editor.langHint")}
          </p>
        </div>

        {/* --- metadata --- */}
        <label htmlFor="art-title" className={labelCls}>
          {t("editor.titleLabel")}
        </label>
        <input
          id="art-title"
          value={title}
          maxLength={300}
          onChange={(e) => setTitle(e.target.value)}
          placeholder={t("editor.titlePlaceholder")}
          className={`${fieldCls} mb-[18px]`}
        />

        <div className="mb-[18px] grid gap-[18px] sm:grid-cols-2">
          <div>
            <label htmlFor="art-category" className={labelCls}>
              {t("editor.categoryLabel")}
            </label>
            <input
              id="art-category"
              value={category}
              maxLength={80}
              onChange={(e) => setCategory(e.target.value)}
              placeholder={t("editor.categoryPlaceholder")}
              className={fieldCls}
            />
          </div>
          <div>
            <label htmlFor="art-tags" className={labelCls}>
              {t("editor.tagsLabel")}
            </label>
            <input
              id="art-tags"
              value={tagsRaw}
              onChange={(e) => setTagsRaw(e.target.value)}
              placeholder={t("editor.tagsPlaceholder")}
              className={fieldCls}
            />
          </div>
        </div>

        <label htmlFor="art-excerpt" className={labelCls}>
          {t("editor.excerptLabel")}
        </label>
        <textarea
          id="art-excerpt"
          value={excerpt}
          maxLength={500}
          rows={2}
          onChange={(e) => setExcerpt(e.target.value)}
          placeholder={t("editor.excerptPlaceholder")}
          className={`${fieldCls} mb-[22px] resize-y`}
        />

        {/* --- cover image (optional) --- */}
        <span className={labelCls}>{t("editor.coverLabel")}</span>
        <div className="mb-[22px]">
          {cover && (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={cover}
              alt={coverAlt}
              className="mb-2.5 h-[200px] w-full rounded-[10px] border border-border object-cover"
            />
          )}
          <div className="flex flex-wrap items-center gap-2">
            <label
              aria-busy={coverUploading}
              className={`inline-flex items-center gap-2 rounded-[9px] border border-border2 px-4 py-2 text-[13.5px] font-semibold text-c5b transition-colors hover:border-hover hover:text-strong ${
                coverUploading ? "cursor-wait opacity-60" : "cursor-pointer"
              }`}
            >
              {coverUploading
                ? t("editor.uploading")
                : cover
                  ? t("editor.replaceImage")
                  : t("editor.coverUpload")}
              <input
                type="file"
                accept={IMAGE_ACCEPT}
                disabled={coverUploading}
                className="sr-only"
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  e.target.value = "";
                  if (file) void uploadCover(file);
                }}
              />
            </label>
            {cover && (
              <button
                type="button"
                onClick={() => setCover("")}
                className="cursor-pointer rounded-[9px] border border-border2 px-4 py-2 text-[13.5px] font-semibold text-c5b transition-colors hover:border-hover hover:text-strong"
              >
                {t("editor.coverRemove")}
              </button>
            )}
          </div>
          {cover && (
            <>
              <label htmlFor="cover-alt" className="sr-only">
                {t("editor.coverAltLabel")}
              </label>
              <input
                id="cover-alt"
                value={coverAlt}
                maxLength={300}
                onChange={(e) => setCoverAlt(e.target.value)}
                placeholder={t("editor.coverAltPlaceholder")}
                className={`${fieldCls} mt-2.5`}
              />
            </>
          )}
          <p className="mt-2 text-[12.5px] leading-[1.6] text-subtle">
            {t("editor.coverHint")}
          </p>
        </div>

        {/* --- format mode --- */}
        <div
          role="radiogroup"
          aria-label={t("editor.formatLabel")}
          className="mb-4 flex gap-1.5"
        >
          {(["markdown", "blocks"] as const).map((m) => (
            <button
              key={m}
              type="button"
              role="radio"
              aria-checked={mode === m}
              onClick={() => setMode(m)}
              className={`cursor-pointer rounded-full border px-4 py-1.5 text-[13.5px] font-semibold transition-colors ${
                mode === m
                  ? "border-accent-ink text-accent-ink"
                  : "border-border2 text-c5b hover:border-hover"
              }`}
            >
              {m === "markdown" ? t("editor.modeMarkdown") : t("editor.modeRich")}
            </button>
          ))}
        </div>

        {/* --- edit / preview tabs + inline toolbar --- */}
        <div className="mb-3 flex items-center justify-between gap-2">
          <div role="tablist" aria-label={t("editor.formatLabel")} className="flex gap-1.5">
            {([false, true] as const).map((p) => (
              <button
                key={String(p)}
                type="button"
                role="tab"
                aria-selected={preview === p}
                aria-controls={p ? "editor-preview" : "editor-edit"}
                onClick={() => setPreview(p)}
                className={`cursor-pointer rounded-[9px] px-3.5 py-1.5 text-[13.5px] font-semibold transition-colors ${
                  preview === p ? "bg-hoverbg text-strong" : "text-muted hover:text-strong"
                }`}
              >
                {p ? t("editor.tabPreview") : t("editor.tabEdit")}
              </button>
            ))}
          </div>
          {!preview && (
            <div className="flex gap-1" aria-label={t("editor.inlineHint")}>
              <ToolbarBtn label={t("editor.bold")} onApply={() => applyInline("**", "**")}>
                <b>B</b>
              </ToolbarBtn>
              <ToolbarBtn label={t("editor.italic")} onApply={() => applyInline("*", "*")}>
                <i>I</i>
              </ToolbarBtn>
              <ToolbarBtn label={t("editor.inlineCode")} onApply={() => applyInline("`", "`")}>
                <span className="font-mono">{"<>"}</span>
              </ToolbarBtn>
              <ToolbarBtn label={t("editor.link")} onApply={() => applyInline("[", "](https://)")}>
                🔗
              </ToolbarBtn>
              {mode === "markdown" && (
                <label
                  aria-label={t("editor.insertImage")}
                  title={t("editor.insertImage")}
                  aria-busy={mdUploading}
                  onMouseDown={(e) => e.preventDefault()}
                  className={`flex h-8 w-8 items-center justify-center rounded-[7px] border border-border2 text-[13px] text-c5b transition-colors hover:border-hover hover:text-strong ${
                    mdUploading ? "cursor-wait opacity-50" : "cursor-pointer"
                  }`}
                >
                  🖼️
                  <input
                    type="file"
                    accept={IMAGE_ACCEPT}
                    disabled={mdUploading}
                    className="sr-only"
                    onChange={(e) => {
                      const file = e.target.files?.[0];
                      e.target.value = "";
                      if (file) void insertMarkdownImage(file);
                    }}
                  />
                </label>
              )}
            </div>
          )}
        </div>

        {preview ? (
          <div
            id="editor-preview"
            role="tabpanel"
            className="mb-4 min-h-[220px] rounded-[14px] border border-border bg-bg px-5 py-2 text-[17px] leading-[1.85] text-body"
          >
            {previewBlocks.length === 0 ? (
              <p className="py-4 text-[14px] text-muted">{t("editor.previewEmpty")}</p>
            ) : (
              previewBlocks.map((b, i) => <BlockView key={i} block={b} slug="preview" />)
            )}
          </div>
        ) : (
          <div id="editor-edit" role="tabpanel" className="mb-4">
            {mode === "markdown" ? (
              <>
                <label htmlFor="art-md" className="sr-only">
                  {t("editor.markdownLabel")}
                </label>
                <textarea
                  id="art-md"
                  data-target="md"
                  value={markdown}
                  rows={16}
                  onChange={(e) => setMarkdown(e.target.value)}
                  placeholder={t("editor.markdownPlaceholder")}
                  className={`${fieldCls} resize-y font-mono text-[13.5px] leading-[1.7]`}
                  {...focusProps}
                />
                <p className="mt-2 text-[12.5px] leading-[1.6] text-subtle">
                  {t("editor.markdownHint")}
                </p>
              </>
            ) : (
              <>
                {rows.map((row, i) => (
                  <BlockRow
                    key={row.id}
                    row={row}
                    index={i}
                    count={rows.length}
                    onUpdate={update}
                    onMove={move}
                    onRemove={remove}
                    onUploadImage={uploadRowImage}
                    focusProps={focusProps}
                  />
                ))}
                <div className="flex flex-wrap items-center gap-2">
                  <button
                    type="button"
                    onClick={() => setRows((rs) => [...rs, newBlock()])}
                    className="cursor-pointer rounded-[9px] border border-dashed border-border2 px-4 py-2 text-[13.5px] font-semibold text-c5b transition-colors hover:border-hover hover:text-strong"
                  >
                    {t("editor.addBlock")}
                  </button>
                  <label
                    aria-busy={insertingImage}
                    title={t("editor.insertImageBlock")}
                    className={`inline-flex items-center gap-2 rounded-[9px] border border-dashed border-border2 px-4 py-2 text-[13.5px] font-semibold text-c5b transition-colors hover:border-hover hover:text-strong ${
                      insertingImage ? "cursor-wait opacity-60" : "cursor-pointer"
                    }`}
                  >
                    🖼️{" "}
                    {insertingImage
                      ? t("editor.uploading")
                      : t("editor.insertImageBlock")}
                    <input
                      type="file"
                      accept={IMAGE_ACCEPT}
                      disabled={insertingImage}
                      className="sr-only"
                      onChange={(e) => {
                        const file = e.target.files?.[0];
                        e.target.value = "";
                        if (file) void insertImageBlock(file);
                      }}
                    />
                  </label>
                </div>
                <p className="mt-2 text-[12.5px] leading-[1.6] text-subtle">
                  {t("editor.inlineHint")}
                </p>
              </>
            )}
          </div>
        )}

        <div role="alert" className="mb-2 min-h-[20px] text-[13px] font-medium text-danger">
          {error}
        </div>
        <button
          type="submit"
          disabled={busy}
          className="btn-accent w-full py-[13px] text-[15px] disabled:opacity-60"
        >
          {busy
            ? editing
              ? t("editor.editSubmitting")
              : t("editor.submitting")
            : editing
              ? t("editor.editSubmit")
              : t("editor.submit")}
        </button>
      </form>
    </div>
  );
}

function Notice({
  title,
  body,
  children,
}: {
  title: string;
  body: string;
  children?: React.ReactNode;
}) {
  return (
    <div className="mx-auto max-w-[520px] px-6 pb-24 pt-16 text-center">
      <h1 className="m-0 mb-2 text-[26px] font-extrabold tracking-[-.02em] text-text">
        {title}
      </h1>
      <p className="m-0 mb-6 text-[15px] leading-[1.7] text-muted">{body}</p>
      {children}
    </div>
  );
}

// Toolbar buttons keep the textarea selection alive by cancelling the
// mousedown focus steal, so applyInline still sees the user's selection.
function ToolbarBtn({
  label,
  onApply,
  children,
}: {
  label: string;
  onApply: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      onMouseDown={(e) => e.preventDefault()}
      onClick={onApply}
      className="flex h-8 w-8 cursor-pointer items-center justify-center rounded-[7px] border border-border2 text-[13px] text-c5b transition-colors hover:border-hover hover:text-strong"
    >
      {children}
    </button>
  );
}

function BlockRow({
  row,
  index,
  count,
  onUpdate,
  onMove,
  onRemove,
  onUploadImage,
  focusProps,
}: {
  row: EditorBlock;
  index: number;
  count: number;
  onUpdate: (id: number, patch: Partial<EditorBlock>) => void;
  onMove: (id: number, dir: -1 | 1) => void;
  onRemove: (id: number) => void;
  onUploadImage: (id: number, file: File) => void;
  focusProps: {
    onFocus: (e: React.FocusEvent<HTMLTextAreaElement>) => void;
  };
}) {
  const t = useT();
  const n = { n: index + 1 };
  const types: Array<{ value: EditorBlockType; label: string }> = [
    { value: "p", label: t("editor.blockP") },
    { value: "h", label: t("editor.blockH") },
    { value: "quote", label: t("editor.blockQuote") },
    { value: "code", label: t("editor.blockCode") },
    { value: "diagram", label: t("editor.blockDiagram") },
    { value: "list", label: t("editor.blockList") },
    { value: "img", label: t("editor.blockImg") },
  ];

  return (
    <div className="mb-3 rounded-[14px] border border-border bg-bg p-3.5">
      <div className="mb-2.5 flex items-center gap-1.5">
        <label htmlFor={`blk-type-${row.id}`} className="sr-only">
          {t("editor.blockTypeLabel", n)}
        </label>
        <select
          id={`blk-type-${row.id}`}
          value={row.type}
          onChange={(e) =>
            onUpdate(row.id, { type: e.target.value as EditorBlockType })
          }
          className="field w-auto cursor-pointer px-3 py-1.5 text-[13.5px] font-semibold"
        >
          {types.map((o) => (
            <option key={o.value} value={o.value}>
              {o.label}
            </option>
          ))}
        </select>
        <span className="flex-1" />
        <IconBtn
          label={t("editor.moveUp", n)}
          disabled={index === 0}
          onClick={() => onMove(row.id, -1)}
        >
          ↑
        </IconBtn>
        <IconBtn
          label={t("editor.moveDown", n)}
          disabled={index === count - 1}
          onClick={() => onMove(row.id, 1)}
        >
          ↓
        </IconBtn>
        <IconBtn
          label={t("editor.removeBlock", n)}
          disabled={count === 1}
          onClick={() => onRemove(row.id)}
        >
          ✕
        </IconBtn>
      </div>

      {(row.type === "p" || row.type === "h" || row.type === "quote") && (
        <>
          <label htmlFor={`blk-text-${row.id}`} className="sr-only">
            {t("editor.textLabel", n)}
          </label>
          <textarea
            id={`blk-text-${row.id}`}
            data-target={String(row.id)}
            value={row.text}
            rows={row.type === "p" ? 3 : 1}
            onChange={(e) => onUpdate(row.id, { text: e.target.value })}
            placeholder={t("editor.textPlaceholder")}
            className={`${fieldCls} resize-y`}
            {...focusProps}
          />
        </>
      )}

      {row.type === "code" && (
        <>
          <label htmlFor={`blk-lang-${row.id}`} className="sr-only">
            {t("editor.langLabel", n)}
          </label>
          <input
            id={`blk-lang-${row.id}`}
            value={row.lang}
            maxLength={40}
            onChange={(e) => onUpdate(row.id, { lang: e.target.value })}
            placeholder={t("editor.langPlaceholder")}
            className={`${fieldCls} mb-2 w-[180px] font-mono text-[13px]`}
          />
          <label htmlFor={`blk-code-${row.id}`} className="sr-only">
            {t("editor.codeLabel", n)}
          </label>
          <textarea
            id={`blk-code-${row.id}`}
            value={row.code}
            rows={6}
            onChange={(e) => onUpdate(row.id, { code: e.target.value })}
            placeholder={t("editor.codePlaceholder")}
            className={`${fieldCls} resize-y font-mono text-[13px] leading-[1.7]`}
            spellCheck={false}
          />
        </>
      )}

      {row.type === "img" && (
        <>
          {row.src && (
            // eslint-disable-next-line @next/next/no-img-element
            <img
              src={row.src}
              alt={row.alt}
              className="mb-2.5 max-h-[240px] rounded-[10px] border border-border"
            />
          )}
          <label
            aria-busy={row.uploading}
            className={`inline-flex items-center gap-2 rounded-[9px] border border-border2 px-4 py-2 text-[13.5px] font-semibold text-c5b transition-colors hover:border-hover hover:text-strong ${
              row.uploading ? "cursor-wait opacity-60" : "cursor-pointer"
            }`}
          >
            {row.uploading
              ? t("editor.uploading")
              : row.src
                ? t("editor.replaceImage")
                : t("editor.uploadImage")}
            <span className="sr-only">{t("editor.imageLabel", n)}</span>
            <input
              type="file"
              accept={IMAGE_ACCEPT}
              disabled={row.uploading}
              className="sr-only"
              onChange={(e) => {
                const file = e.target.files?.[0];
                e.target.value = "";
                if (file) onUploadImage(row.id, file);
              }}
            />
          </label>
          <label htmlFor={`blk-alt-${row.id}`} className="sr-only">
            {t("editor.altLabel", n)}
          </label>
          <input
            id={`blk-alt-${row.id}`}
            value={row.alt}
            maxLength={300}
            onChange={(e) => onUpdate(row.id, { alt: e.target.value })}
            placeholder={t("editor.altPlaceholder")}
            className={`${fieldCls} mt-2`}
          />
          <label htmlFor={`blk-imgcap-${row.id}`} className="sr-only">
            {t("editor.imgCaptionLabel", n)}
          </label>
          <input
            id={`blk-imgcap-${row.id}`}
            value={row.caption}
            maxLength={300}
            onChange={(e) => onUpdate(row.id, { caption: e.target.value })}
            placeholder={t("editor.captionPlaceholder")}
            className={`${fieldCls} mt-2`}
          />
        </>
      )}

      {(row.type === "diagram" || row.type === "list") && (
        <>
          <label htmlFor={`blk-lines-${row.id}`} className="sr-only">
            {row.type === "diagram"
              ? t("editor.stepsLabel", n)
              : t("editor.itemsLabel", n)}
          </label>
          <textarea
            id={`blk-lines-${row.id}`}
            value={row.lines}
            rows={3}
            onChange={(e) => onUpdate(row.id, { lines: e.target.value })}
            placeholder={
              row.type === "diagram"
                ? t("editor.stepsPlaceholder")
                : t("editor.itemsPlaceholder")
            }
            className={`${fieldCls} resize-y`}
          />
          {row.type === "diagram" && (
            <>
              <label htmlFor={`blk-caption-${row.id}`} className="sr-only">
                {t("editor.captionLabel", n)}
              </label>
              <input
                id={`blk-caption-${row.id}`}
                value={row.caption}
                maxLength={300}
                onChange={(e) => onUpdate(row.id, { caption: e.target.value })}
                placeholder={t("editor.captionPlaceholder")}
                className={`${fieldCls} mt-2`}
              />
            </>
          )}
          {row.type === "list" && (
            <label className="mt-2 flex cursor-pointer items-center gap-2 text-[13.5px] font-medium text-strong">
              <input
                type="checkbox"
                checked={row.ordered}
                onChange={(e) => onUpdate(row.id, { ordered: e.target.checked })}
                className="h-4 w-4 accent-[var(--accent-ink)]"
              />
              {t("editor.orderedLabel")}
            </label>
          )}
        </>
      )}
    </div>
  );
}

function IconBtn({
  label,
  disabled,
  onClick,
  children,
}: {
  label: string;
  disabled?: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      type="button"
      aria-label={label}
      title={label}
      disabled={disabled}
      onClick={onClick}
      className="flex h-7 w-7 cursor-pointer items-center justify-center rounded-[7px] border border-border2 text-[12px] text-c5b transition-colors hover:border-hover hover:text-strong disabled:cursor-default disabled:opacity-40"
    >
      {children}
    </button>
  );
}
