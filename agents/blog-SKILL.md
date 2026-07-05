---
name: blog
description: >
  End-to-end blog post pipeline: research → write → check → SEO optimize.
  Produces complete, publish-ready articles optimized for both Google rankings
  and AI citations (ChatGPT, Perplexity, AI Overviews), with sourced statistics,
  answer-first formatting, Key Takeaways box, citation capsules, FAQ section,
  and a full quality + SEO validation pass. Use this skill whenever the user
  asks to write a blog post, article, or long-form content piece; research a
  topic for content; fact-check or verify statistics in a post; audit, score,
  or improve an existing post; run an SEO check; or mentions "blog", "article",
  "content marketing", "SEO", "rank on Google", or "AI citations" — even if
  they only ask for one stage of the pipeline.
license: MIT
---

# Blog: Research → Write → Check → SEO Optimize

A four-stage pipeline for producing blog content that ranks in Google search
and gets cited by AI assistants. Each stage can run standalone (the user may
ask only to fact-check an existing post, or only run an SEO audit), but a
"write me a blog post" request runs all four in order.

Each stage is executed by a dedicated agent with its own instruction file and
reference file:

| Stage | Agent | Reference | What the reference contains |
|-------|-------|-----------|------------------------------|
| 1. Research | `agents/researcher.md` | `references/research.md` | Source tiers, evidence triple, freshness rules, image sourcing, web-content safety |
| 2. Write | `agents/writer.md` | `references/writing-rules.md` | Templates, outline skeleton, answer-first rules, banned phrases, all formatting rules |
| 3. Check | `agents/reviewer.md` | `references/quality-check.md` | 100-point scoring rubric, AI-detection signals, priority classification, fact-check workflow |
| 4. SEO | `agents/seo.md` | `references/seo-checklist.md` | Full on-page checklist and report format |

## Execution model: agents

**If subagents are available** (a Task/agent-spawning tool exists, e.g. Claude
Code or Cowork): dispatch each stage as a subagent. The subagent prompt must
contain: (a) the full text of the stage's `agents/*.md` file, (b) the path to
its reference file with an instruction to read it first, (c) the task brief
(topic, keyword, audience, format), and (d) the prior stage's output (research
packet for the writer; draft + packet for the reviewer; cleared draft for the
SEO agent). Subagents start with no context — pass everything they need.
Stages run sequentially; each stage's output is the next stage's input.

**If subagents are not available** (plain chat): read the stage's `agents/*.md`
file yourself and execute it inline, adopting that agent's role, process, and
output format exactly. The pipeline is identical; only the execution mechanism
changes.

**The fix loop** is driven by the reviewer's machine-readable verdict. The
reviewer's scorecard ends with `BLOCKING: true|false (reason)`. On
`BLOCKING: true`, re-dispatch the **writer** with the scorecard's Prioritized
Fix List as its brief, then re-dispatch the **reviewer** on the revised draft.
The orchestrator (you) holds the loop counter — agents never loop themselves.
Maximum 3 iterations; on the third `BLOCKING: true`, stop and present the
draft with the outstanding issues honestly listed. Only a `BLOCKING: false`
draft proceeds to the SEO agent, and only the SEO-validated draft is delivered.

## Routing

Match the user's request to a starting stage:

- "Write a blog post about X" / "create an article" → run **all four stages** in order
- "Research X for a post" / "find statistics on X" → **Stage 1** only, deliver a research packet
- "Fact-check this post" / "verify these statistics" → **Stage 3** (fact-check half) only
- "Score / audit / review this post" → **Stage 3** (quality scoring) only
- "SEO check" / "optimize this for SEO" / "on-page audit" → **Stage 4** only
- "Rewrite / improve this post" → Stage 3 to diagnose, then Stage 2 rules to fix, then Stage 4 to validate

When running a single stage on an existing post, read the post first (from an
uploaded file, a pasted draft, or by fetching the URL if web access is available).

## Stage 0: Setup (before any writing request)

Clarify only what's missing — don't interrogate the user if they gave a detailed brief:

1. **Target audience** — who reads this?
2. **Primary keyword / search intent** — what query should it win?
3. **Word count** — default 2,000–2,500 words
4. **Format** — markdown (default), MDX, or HTML; auto-detect from the user's project if visible
5. **Author/brand voice** — any constraints, existing posts to link to?
6. **Language** — default **bilingual: Vietnamese + English**, delivered as two
   parallel files with full content parity. The Vietnamese version keeps
   untranslatable technical terms in English. User may request one language only.

Sensible defaults beat a wall of questions. State assumptions and proceed.

## Stage 1: Research

Dispatch the **researcher** agent (`agents/researcher.md`), passing the task
brief and the `references/research.md` path. What the agent does:

1. **Find 8–12 current statistics** (prefer data from the last 12–18 months) via
   web search: `[topic] study statistics data [current year]`. Every statistic
   needs: exact value, source name, URL, publication date.
2. **Apply the source tier filter** — only tier 1–3 sources qualify (government,
   academic, primary research firms, reputable industry publications). Drop
   anything from content farms or unattributed aggregators.
3. **Apply the evidence triple** to every statistic you plan to use:
   year anchor for the prose, publisher + document title for the inline
   citation, URL + retrieval date for the source block.
4. **Enforce coverage** — at least 2 independent sources for load-bearing claims;
   sources that all cite the same upstream report count as one.
5. **Find visuals**: 1 cover image concept + 3–5 inline image candidates from
   free stock platforms (Pixabay/Unsplash/Pexels), a ready-to-run AI image
   generation prompt for every slot (consistent style block across the set,
   aspect ratio, "no text no logos" suffix), and plan 2–4 data visualizations
   from the researched statistics.
6. **Identify content gaps** — skim the top-ranking pages for the target query
   and note what they all miss. That gap is the post's angle.
7. **Collect related sources and concrete evidence** — 3–5 authoritative
   further-reading sources (including at least one opposing view) for the writer
   to link contextually, plus at least one real, sourced example, case, or quote
   per major subtopic (dẫn chứng), so the post relates to the wider conversation
   and reads convincingly rather than just cited.

If web search is unavailable, the agent tells the user, asks them to supply
source material, and marks unverifiable statistics as `[NEEDS SOURCE]` rather
than inventing numbers. **Never fabricate a statistic.**

The agent returns a research packet: statistics table (value, source, URL,
date, verified?), competitive gap analysis, image candidates, chart plan.
Sanity-check the packet before Stage 2 — if fewer than 8 verified statistics
came back, send the researcher back out with narrower queries.

## Stage 2: Write

Dispatch the **writer** agent (`agents/writer.md`), passing the research
packet, the task brief, and the `references/writing-rules.md` path. What the
agent does:

1. **Select a template** from the 12 content types (how-to, listicle,
   comparison, case study, pillar page, etc.) based on search intent.
2. **Craft the title and build the outline** — the title follows the Title
   craft rules (5–8 candidates across curiosity patterns, pick the strongest;
   keyword or head term present, 40–60 chars, crafted natively per language,
   never a flat "What is X?"), then 6–8 H2 sections (60–70% phrased as
   questions), FAQ, conclusion. The orchestrator presents the outline to the
   user for approval before the full draft **unless** the user asked for the
   finished post in one shot.
3. **Write the draft** obeying the non-negotiables:
   - Every H2 opens answer-first: a 40–60 word paragraph with a sourced statistic
   - Key Takeaways box right after the introduction (3–5 bullets, one stat)
   - Paragraphs 40–80 words (hard cap 150); sentences average 15–20 words with deliberate variance
   - Citation capsule (40–60 word self-contained quotable passage) in each major H2
   - 2–3 information gain markers (`[ORIGINAL DATA]`, `[PERSONAL EXPERIENCE]`, `[UNIQUE INSIGHT]`)
   - 5–10 internal linking zones marked `[INTERNAL-LINK: anchor → target description]`
   - 3–6 inline external reference links to the related tier 1–3 sources (descriptive anchors), distinct from internal links
   - Each major H2 pairs its statistic with a concrete example, case, or attributed quote (dẫn chứng), not just the number
   - FAQ section: 3–5 questions, 40–60 word answers, each with a statistic
   - Zero banned AI-tell phrases (full list in the reference), natural contractions, no em dashes
   - Meta description 150–160 chars containing one statistic
4. **Place visuals**: an `[IMAGE]`, `[CHART]`, or `[CALLOUT]` marker every
   300–500 words, never two of the same type consecutively. Each `[IMAGE]`
   marker is self-sufficient: alt sentence, stock candidate URL (or "none"),
   and a complete AI generation prompt sharing one style block across the post.
5. **Produce both language versions** (unless one was requested): draft the
   primary language, then localize to the other with full parity. The Vietnamese
   version keeps untranslatable technical terms in English and reads as natural
   Vietnamese, not a machine translation. Each version gets its own frontmatter
   (`lang`, `alternate`), meta description, and slug.

## Stage 3: Check

Dispatch the **reviewer** agent (`agents/reviewer.md`), passing the draft, the
research packet, and the `references/quality-check.md` path. Two halves:

**3a. Quality scoring (0–100)** across five categories:
Content Quality (30), SEO Optimization (25), E-E-A-T Signals (15),
Technical Elements (15), AI Citation Readiness (15). Score honestly —
a fresh draft usually lands 75–85, not 95. Report the score with a per-category
breakdown and classify every issue as Critical / High / Medium.

**3b. Fact-check** every statistical claim:
- Extract all claims containing numbers, percentages, dollar amounts, or named sources
- If web access is available, fetch each cited URL and confirm the number
  appears in matching context; score confidence 0.0–1.0
- Flag uncited claims as UNVERIFIED with a suggested search query
- Fix or drop anything scoring below 0.7 — **an unverifiable statistic gets
  deleted or replaced, never kept with softened wording**

For bilingual posts the reviewer scores and fact-checks **both** versions and
runs a translation-parity pass: identical stats and sources, matching structure,
natural Vietnamese, technical terms kept in English. A dropped section, a stat
that drifted in translation, or machine-translated Vietnamese is Critical.

The reviewer's scorecard ends with `BLOCKING: true|false (reason)`. Run the
fix loop described in the Execution model section: on `BLOCKING: true`,
re-dispatch the writer with the Prioritized Fix List, then re-dispatch the
reviewer. You hold the counter; cap at 3 iterations. The passing bar is
≥85/100 with zero Critical issues and a clean fact-check.

## Stage 4: SEO Optimize

Dispatch the **SEO** agent (`agents/seo.md`), passing the reviewer-cleared
draft, the primary keyword, any visible site context, and the
`references/seo-checklist.md` path. The agent runs the full on-page checklist:

1. Title tag (40–60 chars, keyword/head term in first half, curiosity hook)
2. Meta description (150–160 chars, statistic, value proposition)
3. Heading hierarchy (single H1, no skipped levels, keyword in 2–3 headings, question ratio)
4. Internal links (3–10, descriptive anchors, deduplicated URLs)
5. External links (tier 1–3 sources only, ≥3, verified reachable if web access exists)
6. Canonical URL, Open Graph tags, Twitter Card
7. URL slug (short, lowercase, keyword-rich, no dates or stop words)
8. Image alt text (full descriptive sentences)
9. For bilingual posts: run the checklist on each version (title, meta, slug,
   keyword localized per language) and add reciprocal hreflang/`alternate` links

The agent applies unambiguous fixes directly to the draft and returns the
validation report table (PASS/FAIL/WARN/N/A per check) plus a prioritized
fix list for anything needing the user's input (e.g., site-level canonical
conventions, real internal-link targets). Use the agent's updated draft as
the final deliverable.

## Delivery

Save the finished post as a file (markdown/MDX/HTML per Stage 0). For a bilingual
post, save both versions — `[slug].vi.md` and `[slug].en.md` (or the platform
convention) — and report both in the summary. Present it with a compact summary:

```
## Blog Post Complete: [Title]

- Versions: [vi + en, parity confirmed | single-language]
- Template: [name] | Word count: ~[N] | Reading time: [N] min
- Statistics: [N] sourced (tier 1–3), [N] unique sources
- Quality score: [N]/100 (Content [n]/30, SEO [n]/25, E-E-A-T [n]/15, Technical [n]/15, AI Citation [n]/15)
- SEO checklist: [N]/[M] passed — [remaining items]
- Fact-check: [N] verified, [N] flagged

### Next steps for you
- Resolve [INTERNAL-LINK] placeholders with your actual URLs
- Replace [IMAGE] markers with final assets (stock candidates + generation prompts are in each marker and the research packet)
- [Anything else needing user input]
```

Be honest in the summary — report real counts and real scores, and list what
still needs the user's hands.
