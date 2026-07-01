"use client";

import { useTheme } from "./ThemeProvider";

/** Row-style theme switch used inside the account menu. */
export function ThemeToggle() {
  const { theme, toggle } = useTheme();
  const dark = theme === "dark";
  return (
    <button
      onClick={toggle}
      className="flex w-full items-center gap-[11px] rounded-[9px] px-3 py-[9px] text-left text-sm font-medium text-strong transition-colors hover:bg-hoverbg"
      role="menuitem"
    >
      <span aria-hidden="true" className="w-[18px] text-center text-[15px]">
        {dark ? "☀️" : "🌙"}
      </span>
      {dark ? "Chế độ sáng" : "Chế độ tối"}
    </button>
  );
}
