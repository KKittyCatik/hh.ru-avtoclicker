export function LoadingScreen() {
  return (
    <div className="grid min-h-[40vh] gap-4 md:grid-cols-2 xl:grid-cols-4">
      {Array.from({ length: 4 }).map((_, index) => (
        <div key={index} className="glass-panel shimmer h-40 animate-shimmer rounded-3xl bg-white/[0.04]" />
      ))}
    </div>
  );
}
