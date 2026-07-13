# Stage 2 Reference: Writing Rules

The post must serve two readers at once: a human scanning on a phone, and an
AI system extracting passages to cite. Every rule below exists for one of
those two readers — answer-first paragraphs and citation capsules are for
extraction; paragraph limits and sentence variance are for humans.

## Bilingual output (Vietnamese + English)

By default every post ships in **two versions, Vietnamese (`vi`) and English
(`en`)**, unless the user asks for one language only. The two are the same
article, not two different ones: identical structure, statistics, evidence
triples, citation capsules, sources, markers, and link zones. A number or a
source that appears in one version appears, unchanged, in the other.

Workflow: draft the primary-language version fully and run the self-check on it,
then produce the second version as a native-quality localization — never a
literal, word-for-word machine translation. Default primary language is English
(sources and the banned-phrase list are English-oriented); the user may set
Vietnamese as primary.

**Vietnamese term policy.** Write natural Vietnamese, but keep untranslatable or
established technical terms in English rather than forcing an awkward calque:
- Always English: code, identifiers, commands, file paths, product/library/
  framework names, and standard jargon Vietnamese developers normally say in
  English (e.g. endpoint, middleware, container, deploy, commit, cache, token,
  session cookie, race condition, backend, framework, query).
- Translate when a clean, widely-used Vietnamese equivalent exists (e.g. "máy
  chủ" for server, "cơ sở dữ liệu" for database in general prose). When unsure,
  keep the English term — a correct English term beats a confusing translation.
- On first use, a short gloss is fine ("truy vấn (query)"), but don't gloss
  every occurrence.

**Per-language quality.** Each version independently satisfies every rule below
(answer-first, paragraph and sentence discipline, no em dashes, visual pacing).
The banned-phrase list applies to the English version; the Vietnamese version
must read as natural Vietnamese with no translationese or stiff word-for-word
phrasing. Do not machine-translate and ship.

**Files and metadata.** Each version is its own file (`[slug].vi.md` and
`[slug].en.md`, or the platform's convention), with a `lang` field in
frontmatter and an `alternate`/hreflang link to the other version. Slugs and
keywords may differ per language to match how each audience actually searches.

## Template selection

Match the topic's search intent to a content type; use its structure to shape
the outline:

| Signal in topic | Template | Structure notes |
|-----------------|----------|-----------------|
| "How to...", process, steps | How-to guide | Numbered steps as H2/H3, prerequisites section, common-mistakes section |
| "Best X", "Top N" | Listicle | Ranked items as H2s, comparison table up top, methodology note |
| Client result, before/after | Case study | Challenge → approach → results (with metrics) → lessons |
| "X vs Y", alternatives | Comparison | Verdict first, side-by-side table, per-criterion H2s, "choose X if..." |
| Broad comprehensive topic | Pillar page | 3,000+ words, broad H2 coverage, heavy internal-link zones |
| "Is X worth it" | Product review | Verdict + rating first, pros/cons, testing methodology, alternatives |
| Opinion, prediction | Thought leadership | Thesis first, evidence sections, counterarguments addressed |
| Expert quotes, multi-source | Roundup | One expert/source per H2, synthesis section |
| Code walkthrough, tool demo | Tutorial | Prerequisites, numbered steps with code blocks, troubleshooting |
| Breaking news, update | News analysis | What happened → why it matters → what to do, heavy freshness |
| Survey, original data | Data research | Methodology, findings as H2s, charts throughout, raw-data access |
| "What is X", Q&A | FAQ/knowledge | Definition first (40–60 words), question H2s, dense FAQ section |

Tell the user which template was selected. If nothing fits, use the generic
outline below.

## Title craft

The title decides the click. A flat "[Keyword]: Overview" or a bare
"What is X?" title is monotone and loses to every competitor saying the same
thing. Craft the displayed title for curiosity; leave the plain definitional
phrasing to the slug and the meta description.

Process: write 5–8 candidates spanning at least 3 of the patterns below, then
pick the one a busy reader would actually click. Do this per language: craft
the Vietnamese title natively for Vietnamese readers, never translate the
English title word-for-word.

| Pattern | Shape | Example |
|---------|-------|---------|
| Curiosity gap | Withhold the how/why, promise the reveal | "The Database Born from a Hypervisor" |
| Number + outcome | Verified stat + concrete result | "How Discord Serves Trillions of Messages on 72 Nodes" |
| Tension / contrast | Two facts that shouldn't coexist | "10x Faster, 60 Places Less Popular" |
| Bold verdict | Take a stance the post defends | "ScyllaDB Won't Replace Your SQL Database, and That's the Point" |
| Mistake / warning | Name the pain the reader fears | "When ScyllaDB Wrecks Your Data Model" |
| Intriguing question | A question with a twist, never bare "What is X?" | "Can 4 Nodes Really Beat 40?" |

Hard rules (the SEO checks still apply):
- Primary keyword, or at least its head term, appears naturally, ideally in
  the first half
- 40–60 characters (title-tag limit)
- Any number in a title comes from a Verified packet row
- The post must cash every promise the title makes: curiosity, not clickbait
- No ALL CAPS, no "SHOCKING", no listicle bait unless the template is a listicle

## Narrative arc: why → high level → details

Every post, whatever the template, walks the reader down the same slope:

1. **Why (the cause)** — open with the problem, pain, or need that makes the
   topic necessary: what breaks, what it costs, who hits it. The reader must
   feel why this matters before any solution appears.
2. **High level (the idea)** — then the bird's-eye view: what the solution or
   concept is, its core mental model, in plain words — before any code,
   configuration, or step-by-step.
3. **Details (the how)** — only then the deep-dive H2s: implementation,
   steps, numbers, edge cases, trade-offs.

Order the H2s along this slope and never invert it: don't open the body with
implementation detail, and don't introduce a mechanism before the reader knows
why they'd need it. Map the arc onto whichever template was chosen (a how-to
still opens with why the task matters before prerequisites; a tutorial states
the pain before the first numbered step).

This coexists with answer-first formatting: a comparison still leads with its
verdict and every H2 still opens with its answer — the arc governs the order
of sections, not the shape of paragraphs.

## Outline skeleton (generic)

```
# [Title per Title craft: curiosity hook + primary keyword]

## Introduction (100–150 words)
- Hook with a surprising statistic
- Problem/opportunity statement
- What the reader will learn

> **Key Takeaways** (3–5 bullets, 40–60 words combined, 1 sourced stat)

## H2 [question format] (300–400 words)
- Answer-first paragraph (40–60 words, stat + source)
- Supporting evidence / [IMAGE or CHART]
- Citation capsule
- [INTERNAL-LINK: anchor → target description]

... 6–8 H2 sections total, ordered along the narrative arc (why →
    high level → details); 60–70% as questions, the rest as statements
    for variety; each 200–400 words ...

## FAQ (3–5 questions, 40–60 word answers, each with a stat)

## Conclusion (100–150 words: key takeaways bulleted, single CTA,
   [INTERNAL-LINK: → next logical content])
```

Present the outline for approval before drafting, unless the user asked for a
one-shot finished post.

## Frontmatter

```yaml
---
title: "[Title per Title craft: curiosity hook, keyword present, 40–60 chars]"
description: "[150–160 chars, fact-dense, includes 1 statistic]"
coverImage: "[direct URL or path]"
coverImageAlt: "[Full descriptive sentence]"
date: "YYYY-MM-DD"
lastUpdated: "YYYY-MM-DD"
author: "[Name]"
tags: ["keyword1", "keyword2", "keyword3"]
lang: "en"                    # this version's language: "en" or "vi"
alternate: "[slug].vi.md"     # path/URL of the other-language version (hreflang)
---
```

Adapt field names to the user's platform convention if one is visible
(e.g., `image` / `hero` / `thumbnail`).

## Non-negotiable writing rules

### Answer-first formatting
Every H2 section opens with a 40–60 word paragraph containing (a) at least
one specific statistic with source attribution and (b) a direct answer to the
heading's implicit question. Pattern:

```markdown
## How Does X Impact Y in 2026?

In 2026, [statistic] ([Publisher](url), [Title], year). [Direct answer in
1–2 more sentences: the implication and what it means for the reader.]
```

### The evidence triple, applied at drafting time
Every public statistic carries: year anchor in prose ("In 2026," before the
number, not just in parentheses), inline citation naming publisher AND
document title, and an entry in the source block at the bottom of the post:
`[Publisher], [Title], retrieved YYYY-MM-DD, [URL]`. If any component is
missing, the statistic doesn't go in.

### Key Takeaways box
Immediately after the introduction:
- 3–5 bullets, 40–60 words combined
- Self-contained (understandable without the article)
- Contains 1 specific statistic with source name
- Label alternatives if the brand voice calls for it: "The Bottom Line",
  "At a Glance", "In Brief"

### Paragraph and sentence discipline
- Paragraphs: 40–80 words, hard cap 150; one idea per paragraph; most
  important sentence first
- Sentences: average 15–20 words, but **deliberately mix** 8-word punches
  with 25-word explanations — uniform length reads as machine-written
- Active voice; target Flesch Reading Ease 60–70 (Grade 7–8). Technical
  audiences may run Grade 10–12 with 30-word max sentences.

### Headings
- One H1 (the title), H2s for sections, H3s only under H2s — never skip levels
- 60–70% of H2s phrased as questions
- Primary keyword naturally in 2–3 headings
- An H2 roughly every 200–300 words

### Citation capsules
In each major H2 body, one 40–60 word self-contained passage an AI system
could quote verbatim: specific claim + data point + source attribution,
declarative style, makes sense in isolation. Example:

> According to Gartner's 2026 B2B Buying Survey, 58% of enterprise buyers now
> consult AI assistants before contacting a vendor (Gartner, 2026). This shift
> means B2B content must answer specific questions concisely enough for AI
> systems to extract and cite in their responses.

### Information gain markers
2–3 per post, marking original value not found elsewhere:
- `[ORIGINAL DATA]` — proprietary surveys, experiments, case-study metrics
- `[PERSONAL EXPERIENCE]` — first-hand observations, "when we tried X, Y happened"
- `[UNIQUE INSIGHT]` — analysis others haven't made, contrarian takes backed by data

Use HTML comments (`<!-- [ORIGINAL DATA] -->`) before the paragraph, or a
visible callout (`> **Our finding:** ...`). Only mark content that genuinely
comes from the user's own experience or data — ask the user for these inputs;
never fabricate first-hand experience.

### Internal linking zones
5–10 per 2,000-word post, format
`[INTERNAL-LINK: descriptive anchor text → target description]`.
Place in: introduction (→ pillar content), each H2 (→ supporting articles),
FAQ answers (→ deeper content), conclusion (→ next logical read). Never
"click here" / "read more" anchors. If the user's existing posts are visible,
resolve zones to real URLs directly.

### External reference links (relate to related sources)
Separate from the bottom source block and from internal-link zones: weave 3–6
contextual outbound links into the prose so the post relates to the wider
conversation. Link a load-bearing claim or a key term to the tier 1–3 source
that backs it, and drop in genuinely useful further reading (a fuller study, an
official doc, an opposing view) where a reader would want it.
- Descriptive anchor naming the source or its finding, never a bare URL or
  "source" / "here". Example: `[Ahrefs' analysis of 1M SERPs](url) found...`
- Link the primary/upstream publisher, not an aggregator re-reporting it.
- One outbound link per claim; don't stack several links on one sentence.
- Every inline outbound link also appears in the bottom source block.
- Relevance over quantity: a reader should want to open every link.

### Concrete evidence and examples (dẫn chứng)
A statistic states; an example convinces. In each major H2, pair the sourced
number with at least one concrete piece of supporting evidence that makes it
tangible: a named real-world case ("after X switched to Y, Z dropped 30%", with
the source), a short attributed expert quote, a before/after comparison, or a
specific scenario the reader recognizes. Prefer specifics — names, dates,
outcomes — over generic assertion. Examples must be real and sourced, or come
from the user's own experience (wrapped in an information gain marker). Never
invent a case, a quote, or an outcome.

### FAQ section
3–5 questions, 40–60 word answers, each answer contains a statistic. For MDX
platforms with an FAQ schema component, use it; otherwise plain
`## Frequently Asked Questions` with `###` questions.

### Self-promotion
Maximum 1 brand mention (author-bio context). Educational tone throughout.

## Naturalness rules (anti-AI-tells)

- **Voice**: an experienced professional talking to a peer — conversational
  yet authoritative, never a robot presenting a textbook
- **Plain words**: pick the simple, common word over the fancy one — "use"
  not "utilize", "help" not "facilitate", "start" not "commence", "buy" not
  "purchase". A smart reader new to the topic should follow every sentence on
  first read. Necessary jargon gets a one-clause plain-language explanation on
  first use ("idempotent — running it twice gives the same result"); jargon
  that isn't necessary gets cut. Same rule in Vietnamese: everyday words over
  formal or Sino-Vietnamese phrasing when a common word exists
- **Banned phrases** — never use: "in today's digital landscape",
  "it's important to note", "dive into", "delve", "game-changer",
  "navigate the landscape", "revolutionize", "seamlessly", "cutting-edge",
  "harness the power of", "leverage" (as a verb), "crucial", "elevate",
  "foster", "landscape" (overused), "multifaceted", "robust", "tapestry",
  "embark", "unlock", "unleash", "in the realm of", "unprecedented",
  "transformative", "empower", "furthermore" (as a paragraph opener)
- **No em dashes.** Replace with commas, colons, periods, or split sentences.
- **Contractions**: use them naturally ("it's", "don't", "we've")
- **Rhetorical questions**: roughly one every 200–300 words
- **Hedging where honest**: "in our experience", "we've found that"
- **Burstiness**: verify the finished draft mixes short and long sentences.
  Some short. Some long and descriptive.
- **Varied openings**: no repeated sentence openings — consecutive sentences
  (and consecutive paragraph openers) never start with the same word or
  construction
- **Break the symmetry**: avoid the "Rule of Threes" and other predictable,
  overly symmetrical structures. Don't make every list exactly three items or
  every section the same shape; vary list lengths and section rhythms the way
  a human would
- **Show, don't tell**: strong active verbs over weak adjectives ("latency
  collapsed from 90ms to 4ms", not "performance was significantly better");
  specific, concrete details over broad generalizations
- **No filler bookends**: never open with "In this article, we will
  discuss..." or close with "In conclusion...". The introduction hooks with a
  statistic and gets moving; the conclusion earns its keep with takeaways and
  a CTA, not a restatement. No generic filler paragraphs anywhere

## Visual pacing

An `[IMAGE: ...]`, `[CHART: type + data + source]`, `[VIDEO: topic]`, or
`[CALLOUT]` marker every 300–500 words. Never two consecutive markers of the
same type. Images go after H2 headings, before body text. Alt text is a full
descriptive sentence. Charts get a `<figure>` wrapper with a
`<figcaption>Source: [Name], [Year]</figcaption>`.

If the environment can generate SVG charts, embed them directly from the
Stage 1 chart plan; otherwise leave the `[CHART]` markers with complete data
so the user can build them.

### Image markers carry generation prompts

Every `[IMAGE:]` marker is self-sufficient: the user must be able to either
drop in the stock candidate or paste the prompt into an image generator
(Midjourney, DALL-E, Ideogram, etc.) without further thought. Format:

```
[IMAGE: <alt-text sentence> | stock: <direct URL from a Verified packet row, or "none"> | gen: <complete generation prompt>]
```

Prompt rules (write prompts in English regardless of post language):
- One sentence each for subject, setting, composition; concrete nouns, no
  mood-board vagueness ("a stylized CPU chip with four isolated data lanes",
  not "a technological feeling of speed")
- A style block, **identical across every prompt in the post**, so the
  illustrations read as one set (e.g. "isometric flat vector illustration,
  dark navy background, cyan and orange accents, clean geometric lines,
  no gradients")
- Aspect ratio stated: cover 1200×630 (~16:9), inline 3:2 or 16:9
- Always end with "no text, no words, no logos": generated text renders as
  gibberish
- For architecture/data-flow diagrams prefer a `[CHART]` or custom SVG; if
  one must be generated, keep shapes abstract rather than labeled

Example:

```
[IMAGE: A stylized CPU with an isolated lane of data flowing into each core, illustrating ScyllaDB's shard-per-core design. | stock: none | gen: Isometric flat vector illustration of a stylized CPU chip with four glowing lanes of data packets flowing into four separate cores, each lane fully isolated from the others, dark navy background, cyan and orange accents, clean geometric lines, no gradients, 16:9, no text, no words, no logos]
```

## Pre-handoff self-check

Before passing the draft to Stage 3, verify:
- [ ] Sections follow the narrative arc: why → high level → details, never
      detail-first
- [ ] Every H2 opens with stat + source (40–60 words)
- [ ] No paragraph over 150 words
- [ ] All statistics carry the evidence triple; source block present at bottom
- [ ] Key Takeaways box present with a stat
- [ ] 2–3 information gain markers
- [ ] Citation capsules in major H2s
- [ ] 5–10 internal-link zones marked
- [ ] 3–6 inline external reference links to related tier 1–3 sources, descriptive anchors
- [ ] Each major H2 pairs its statistic with a concrete example, case, or quote (real + sourced)
- [ ] FAQ with 3–5 stat-bearing answers
- [ ] Meta description 150–160 chars with a stat
- [ ] Zero banned phrases, zero em dashes, contractions present
- [ ] Plain words throughout; every jargon term either explained on first
      use or cut
- [ ] No repeated sentence/paragraph openings; no filler bookends
      ("In this article...", "In conclusion..."); no generic filler paragraphs
- [ ] Lists and sections vary in length and shape (no forced Rule of Threes)
- [ ] Visual marker every 300–500 words, types alternating
- [ ] Both language versions (vi + en) present with full parity, unless single-language requested
- [ ] Vietnamese version reads naturally; untranslatable technical terms kept in English
