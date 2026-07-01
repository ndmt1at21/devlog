export default function Loading() {
  return (
    <div className="mx-auto max-w-[1120px] px-6 pt-11">
      <div className="mb-5 h-4 w-32 animate-pulse rounded bg-chip" />
      <div className="mb-11 h-[240px] w-full animate-pulse rounded-[18px] bg-chip" />
      <div className="grid grid-cols-1 gap-[26px] sm:grid-cols-2 lg:grid-cols-3">
        {Array.from({ length: 6 }).map((_, i) => (
          <div
            key={i}
            className="h-[340px] animate-pulse rounded-2xl bg-chip"
          />
        ))}
      </div>
    </div>
  );
}
