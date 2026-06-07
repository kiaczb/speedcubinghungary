import React from "react";

type BadgeProps = {
  label: string;
  color: string;
  className?: string;
};

const Badge: React.FC<BadgeProps> = ({ label, color, className }) => {
  return (
    <span
      className={`inline-block px-2 py-1 text-white text-xs font-bold rounded ${color} ${
        className || ""
      }`}
    >
      {label}
    </span>
  );
};

export default Badge;
