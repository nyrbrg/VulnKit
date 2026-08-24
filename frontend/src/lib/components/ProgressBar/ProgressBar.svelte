<script lang="ts">
  import type { StyleConfig } from "$lib/types/style";
  import { styleVars } from "$lib/utils/styleVars";

  type Variant = "success" | "warning" | "error";

  const defaultStyleConfig = {
    backgroundColor: "#e5e7eb",
    fillColor: "#9ca3af",
    borderRadius: "9999px",
    height: "0.25rem",
  };

  const variantStyles: Record<Variant, StyleConfig> = {
    success: { fillColor: "#10b981" },
    warning: { fillColor: "#fbbf24" },
    error: { fillColor: "#f87171" },
  };

  let {
    value,
    variant = undefined,
    styleConfig = {},
    class: className = "",
  }: {
    value: number;
    variant?: Variant;
    styleConfig?: StyleConfig;
    class?: string;
  } = $props();

  const mergedStyleConfig = $derived({
    ...defaultStyleConfig,
    ...(variant ? variantStyles[variant] : {}),
    ...styleConfig,
  });

  const clampedValue = $derived(Math.min(100, Math.max(0, value)));
</script>

<div
  class="progressTrack overflow-hidden {className}"
  style={styleVars({
    backgroundColor: mergedStyleConfig.backgroundColor,
    borderRadius: mergedStyleConfig.borderRadius,
    height: mergedStyleConfig.height,
  })}
>
  <div
    class="progressFill h-full transition-all duration-500"
    style={`${styleVars({ fillColor: mergedStyleConfig.fillColor })} width:${clampedValue}%;`}
  ></div>
</div>

<style>
  .progressTrack {
    background-color: var(--backgroundColor);
    border-radius: var(--borderRadius);
    height: var(--height);
  }
  .progressFill {
    background-color: var(--fillColor);
    border-radius: var(--borderRadius);
  }
</style>
