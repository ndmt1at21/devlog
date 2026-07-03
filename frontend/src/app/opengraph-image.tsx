import { ImageResponse } from "next/og";
import { SITE_NAME } from "@/lib/seo";

// Branded default social-share card (Open Graph + Twitter). Next serves this for
// any page that doesn't provide its own image, and `absoluteUrl("/opengraph-image")`
// is reused as the Organization logo in JSON-LD.
export const alt = "jamti — nhật ký lập trình";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OgImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          background: "#2A2117",
          padding: "72px 80px",
          color: "#FFFDF6",
          fontFamily: "sans-serif",
        }}
      >
        <div style={{ display: "flex", alignItems: "center", gap: 20 }}>
          <div
            style={{
              width: 64,
              height: 64,
              borderRadius: 16,
              background: "#FFC531",
              color: "#4A3512",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              fontSize: 44,
              fontWeight: 800,
            }}
          >
            j
          </div>
          <div style={{ fontSize: 40, fontWeight: 700 }}>{SITE_NAME}</div>
        </div>

        <div
          style={{
            display: "flex",
            fontSize: 68,
            fontWeight: 800,
            lineHeight: 1.15,
            letterSpacing: "-0.02em",
            maxWidth: 900,
          }}
        >
          Nhật ký lập trình — kiến thức, dự án & ghi chú kỹ thuật.
        </div>

        <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
          <div style={{ width: 40, height: 6, background: "#FFC531" }} />
          <div style={{ fontSize: 30, color: "#c9bfae" }}>
            Blog & series về lập trình
          </div>
        </div>
      </div>
    ),
    { ...size },
  );
}
