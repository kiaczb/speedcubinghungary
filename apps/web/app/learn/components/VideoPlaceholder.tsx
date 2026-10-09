import { FaPlay } from "react-icons/fa";

type Props = {
  label: string;
  compact?: boolean;
  className?: string;
};

export default function VideoPlaceholder({
  label,
  compact = false,
  className = "",
}: Props) {
  if (compact) {
    return (
      <div
        role="img"
        aria-label={`${label}: a videó hamarosan érkezik`}
        className={`flex items-center gap-2.5 rounded-control border border-dashed border-line-strong bg-surface-muted/60 px-3 py-2 text-xs text-subtle-foreground sm:aspect-video sm:flex-col sm:justify-center sm:gap-2 sm:px-2 sm:py-0 sm:text-center ${className}`}
      >
        <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-surface text-muted-foreground shadow-card">
          <FaPlay className="ml-px h-2.5 w-2.5" />
        </span>
        Videó hamarosan
      </div>
    );
  }

  return (
    <div
      role="img"
      aria-label={`${label}: a videó hamarosan érkezik`}
      className={`flex aspect-video w-full flex-col items-center justify-center gap-3 rounded-card border border-dashed border-line-strong bg-surface-muted text-center ${className}`}
    >
      <span className="flex h-14 w-14 items-center justify-center rounded-full bg-surface text-foreground shadow-card">
        <FaPlay className="ml-1 h-4 w-4" />
      </span>
      <span className="px-6 text-sm font-medium text-muted-foreground">
        A videó hamarosan érkezik
      </span>
    </div>
  );
}
