// API response types — mirror the Go backend DTOs (backend/internal/handler).

export type BlockType =
  | "p"
  | "h"
  | "quote"
  | "code"
  | "diagram"
  | "list"
  | "img"
  | "ad";

export interface Block {
  type: BlockType;
  text?: string;
  lang?: string;
  code?: string;
  caption?: string;
  steps?: string[];
  /** List items for `list` blocks; may carry inline markdown spans. */
  items?: string[];
  ordered?: boolean;
  /** Image URL for `img` blocks (must live under the public image origin). */
  src?: string;
  /** Alternative text for `img` blocks. */
  alt?: string;
  /** Server-rendered Shiki HTML for `code` blocks (added during SSR). */
  html?: string;
}

/** Response of POST /uploads: PUT the bytes to uploadUrl, embed publicUrl. */
export interface UploadTicket {
  uploadUrl: string;
  publicUrl: string;
}

/** Payload for POST /articles. Body is either markdown source or editor blocks. */
export interface NewArticleInput {
  title: string;
  excerpt?: string;
  category: string;
  tags: string[];
  format: "markdown" | "blocks";
  content?: string;
  body?: Block[];
}

export interface ArticleSummary {
  slug: string;
  title: string;
  excerpt: string;
  category: string;
  author: string;
  authorInitial: string;
  read: string;
  date: string;
  /** RFC 3339 publish timestamp — used for SEO (canonical dates, sitemap lastmod). */
  publishedAt: string;
  tags: string[];
  cover?: string;
  featured: boolean;
  isSeries: boolean;
  series?: string;
  seriesBadge?: string;
}

export interface SeriesPart {
  id: string;
  part: number;
  ptitle: string;
  isCurrent: boolean;
  pLocked: boolean;
}

export interface PartLink {
  id: string;
  ptitle: string;
}

export interface ArticleDetail extends ArticleSummary {
  body: Block[];
  locked: boolean;
  inSeries: boolean;
  seriesTitle?: string;
  seriesPartLabel?: string;
  seriesParts?: SeriesPart[];
  prevPart?: PartLink;
  nextPart?: PartLink;
}

export interface Comment {
  name: string;
  text: string;
  time: string;
  initial: string;
}

export interface ReactionStatus {
  likes: number;
  liked: boolean;
  bookmarked: boolean;
}

export interface Plan {
  key: "month" | "year";
  name: string;
  price: string;
  note: string;
}

export interface SessionUser {
  name: string;
  email: string;
  premium: boolean;
  /** UI hint: the IAM "articles:create" permission (server re-checks on POST). */
  canWrite: boolean;
}

export interface MeResponse {
  authenticated: boolean;
  user?: SessionUser;
}

export interface SubscriptionState {
  active: boolean;
  plan?: string;
  status?: string;
}
