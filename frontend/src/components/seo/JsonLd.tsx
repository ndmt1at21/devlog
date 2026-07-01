// Renders a JSON-LD structured-data block. Server-safe (no client JS); the
// serialized graph is inlined into the HTML for crawlers. `<` is escaped to
// avoid breaking out of the <script> context if any field contains markup.
export function JsonLd({ data }: { data: Record<string, unknown> }) {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{
        __html: JSON.stringify(data).replace(/</g, "\\u003c"),
      }}
    />
  );
}
