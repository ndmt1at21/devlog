// Client-side image upload: ask the backend for a presigned ticket, PUT the
// bytes straight to the bucket (they never transit the API), return the public
// CDN URL to embed. Type/size limits mirror the backend's authoritative checks.
import { api } from "./api";

export const IMAGE_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
  "image/avif",
];
export const IMAGE_ACCEPT = IMAGE_TYPES.join(",");
export const MAX_IMAGE_BYTES = 5 * 1024 * 1024;

/** Upload one image and resolve to its public URL. Throws ApiError from the
 * ticket request, or a plain Error when the bucket PUT itself fails. */
export async function uploadImage(file: File): Promise<string> {
  const ticket = await api.createUpload({ type: file.type, size: file.size });
  const res = await fetch(ticket.uploadUrl, {
    method: "PUT",
    headers: { "Content-Type": file.type },
    body: file,
  });
  if (!res.ok) throw new Error(`upload failed: ${res.status}`);
  return ticket.publicUrl;
}

/** File name without extension — the default alt text suggestion. */
export function baseName(name: string): string {
  const i = name.lastIndexOf(".");
  return (i > 0 ? name.slice(0, i) : name).trim();
}
