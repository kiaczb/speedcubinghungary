import React from "react";

type BadgeProps = {
  label: string;
  color: string;
  className?: string;
};

const Badge: React.FC<BadgeProps> = ({ label, color, className = "" }) => {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold uppercase tracking-wide ring-1 ring-inset ${color} ${className}`}
    >
      {label}
    </span>
  );
};

export default Badge;
