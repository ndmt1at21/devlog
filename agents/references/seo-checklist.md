# Stage 4 Reference: SEO Optimize

Post-writing on-page validation. Apply fixes directly to the draft where the
fix is unambiguous; report anything needing site-level context (canonical
conventions, existing internal pages) as a prioritized fix for the user.

**Bilingual posts:** run every check below on each language version. Title, meta
description, slug, and keyword are per-language — localize them to how that
audience searches, don't reuse the English string for Vietnamese. Then run the
hreflang check (§12).

## 1. Title tag
| Check | Pass criteria |
|-------|---------------|
| Length | 40–60 characters (no SERP truncation) |
| Keyword | Primary keyword, or its head term, in the first half |
| Hook | Creates curiosity or tension: a verified number, a contrast, a bold verdict, an intriguing question, or a power word (Guide, Best, How, Why, Proven). A flat "[Keyword]: overview" / bare "What is X?" pattern fails |
| Honesty | The post cashes every promise the title makes; numbers in the title are sourced |
| Truncation safety | No critical meaning lost if cut at 60 chars |
| Uniqueness | Specific to this content, not generic |

## 2. Meta description
| Check | Pass criteria |
|-------|---------------|
| Length | 150–160 characters |
| Statistic | Contains at least one specific number |
| Value proposition | Ends with a clear reader benefit |
| Keyword | Appears once, naturally |
| Action implied | learn / discover / find out / see |

## 3. Heading hierarchy
| Check | Pass criteria |
|-------|---------------|
| Single H1 | Exactly one (the title) |
| No skipped levels | H1→H2→H3 only |
| Keyword in headings | In 2–3 headings, naturally |
| Question ratio | 60–70% of H2s are questions |
| H2 count | 6–8 for a standard post |
| Length | Each heading under 70 characters |

## 4. Internal links
| Check | Pass criteria |
|-------|---------------|
| Count | 3–10 per post (5–10 per 2,000 words) |
| Anchor text | Descriptive — never "click here"/"read more" |
| Distribution | Spread through the post, not clustered |
| No self-links | Post doesn't link to itself |
| Deduplication | Each URL appears at most once in body content |

For duplicates: normalize URLs (strip trailing slashes, fragments, query
params), keep the instance with the most descriptive anchor, remove the rest.
Google records only 1–2 anchor texts per URL per page. Max ~50 total links
per page.

If the user's site content is visible, verify links point to real pages and
flag one-way links that would benefit from a backlink from the target page.

## 5. External links
| Check | Pass criteria |
|-------|---------------|
| Source tier | Tier 1–3 only (see research reference) |
| Count | At least 3 authoritative outbound links |
| Reachable | Verify top links resolve, if web access available |
| Rel attributes | nofollow for sponsored/UGC, noopener for new tabs |
| No gratuitous competitor links | Only when genuinely necessary |

## 6. Evidence triple audit (citations)
For every statistic, confirm all three components survived the writing stage:
- Year anchor in prose before the number ("In 2026, ...")
- Inline citation names publisher AND document title
- Source block at the bottom lists URL + `retrieved YYYY-MM-DD`

Any statistic failing the triple: fix the citation or remove the claim.

## 7. Canonical URL
| Check | Pass criteria |
|-------|---------------|
| Present | Defined in frontmatter or meta |
| Format | Full absolute URL |
| Self-referencing | Points to the page itself |
| Trailing slash | Consistent with the site's convention |

## 8. Open Graph tags
| Check | Pass criteria |
|-------|---------------|
| og:title | Present, matches/complements title |
| og:description | Present, 150–160 chars |
| og:image | Present, ≥1200×630, absolute URL |
| og:type | "article" |
| og:url | Matches canonical |

## 9. Twitter Card
| Check | Pass criteria |
|-------|---------------|
| twitter:card | "summary_large_image" |
| twitter:title | Under 70 chars |
| twitter:description | Under 200 chars |
| twitter:image | Same as or similar to og:image |

Mark N/A if the site has no X/Twitter presence.

## 10. URL slug
| Check | Pass criteria |
|-------|---------------|
| Length | Path under 75 characters |
| Keyword | Primary keyword or close variant present |
| No dates | No /2025/ or /2026/ segments |
| Characters | Lowercase letters, numbers, hyphens only |
| Stop words | Minimal "the/a/and/of" |
| Clean | No .html/.php extension |

## 11. Image checks
- Every image has a full-sentence descriptive alt text
- Cover image + og:image present in frontmatter
- Only direct image URLs (no photo-page URLs)

## 12. Bilingual / hreflang (when two versions exist)
| Check | Pass criteria |
|-------|---------------|
| lang declared | Each version's frontmatter sets `lang` ("vi" or "en") |
| Reciprocal alternate | Each version links to the other via hreflang/`alternate` |
| Per-language slug | Vietnamese slug uses the Vietnamese keyword (ASCII-folded), not the English one |
| Per-language meta | Title tag and meta description localized, not copied across |
| Canonical per version | Each version self-canonicalizes to its own URL |

## Report format

```
## SEO Validation Report: [Title]

**File**: [path] | **Overall**: [X/Y passed] — [PASS / NEEDS WORK / FAIL]

| # | Check | Status | Details | Fix |
|---|-------|--------|---------|-----|
| 1 | Title length | PASS | 52 chars | — |
| 2 | Meta description stat | FAIL | No number found | Add key stat from Key Takeaways |
| ... |

### Fixes applied directly
- [What was changed in the draft]

### Priority fixes needing your input
1. [Most impactful — what and where]
2. ...
```

Status values: **PASS**, **FAIL** (fix provided), **WARN** (edge case,
recommendation given), **N/A** (doesn't apply, with reason).
