type Props = {
  label: string;
  onClick: () => void;
  disabled?: boolean;
  danger?: boolean;
  children: React.ReactNode;
};

export default function AdminIconButton({
  label,
  onClick,
  disabled,
  danger = false,
  children,
}: Props) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      aria-label={label}
      title={label}
      className={`relative z-10 inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-control text-muted-foreground transition-colors duration-200 ease-smooth disabled:pointer-events-none disabled:opacity-30 ${
        danger
          ? "hover:bg-brand-red/10 hover:text-brand-red"
          : "hover:bg-surface-muted hover:text-foreground"
      }`}
    >
      {children}
    </button>
  );
}
