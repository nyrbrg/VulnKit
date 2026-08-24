<script lang="ts">
  import type { StyleConfig } from "$lib/types/style";
  import type { Snippet } from "svelte";
  import { styleVars } from "$lib/utils/styleVars";

  type Variant = "success" | "warning" | "error";

  const defaultStyleConfig = {
    backgroundColor: "#f9fafb",
    borderColor: "#e5e7eb",
    color: "#4a5565",
    borderWidth: "1px",
    borderRadius: "0.5rem",
    padding: "0.5rem 0.75rem",
    fontSize: "0.875rem",
  };

  const variantStyles: Record<Variant, StyleConfig> = {
    success: { backgroundColor: "#ecfdf5", borderColor: "#a7f3d0", color: "#065f46" },
    warning: { backgroundColor: "#fffbeb", borderColor: "#fde68a", color: "#92400e" },
    error: { backgroundColor: "#fef2f2", borderColor: "#fecaca", color: "#b91c1c" },
  };

  let {
    variant = undefined,
    styleConfig = {},
    children,
    class: className = "",
  }: {
    variant?: Variant;
    styleConfig?: StyleConfig;
    children: Snippet;
    class?: string;
  } = $props();

  const mergedStyleConfig = $derived({
    ...defaultStyleConfig,
    ...(variant ? variantStyles[variant] : {}),
    ...styleConfig,
  });
</script>

<div
  class="notification {className}"
  style={styleVars(mergedStyleConfig)}
>
  {@render children()}
</div>

<style>
  .notification {
    background-color: var(--backgroundColor);
    color: var(--color);
    border: var(--borderWidth) solid var(--borderColor);
    border-radius: var(--borderRadius);
    padding: var(--padding);
    font-size: var(--fontSize);
  }
</style>
