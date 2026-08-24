<script lang="ts">
  import type { StyleConfig } from "$lib/types/style";
  import type { Snippet } from "svelte";
  import { styleVars } from "$lib/utils/styleVars";

  const defaultStyleConfig = {
    backgroundColor: "#f8f7f4",
    color: "#9a9890",
    borderColor: "currentColor",
    borderWidth: "0.5px",
    borderRadius: "20px",
    padding: "0.125rem 0.5rem",
    fontSize: "11px",
  };

  let {
    children,
    styleConfig = {},
    class: className = "",
  }: { children: Snippet; styleConfig?: StyleConfig; class?: string } = $props();

  const mergedStyleConfig = $derived({ ...defaultStyleConfig, ...styleConfig });
</script>

<span
  class="badge {className}"
  style={styleVars(mergedStyleConfig)}
>
  {@render children()}
</span>

<style>
  .badge {
    background-color: var(--backgroundColor);
    color: var(--color);
    border: var(--borderWidth) solid var(--borderColor);
    border-radius: var(--borderRadius);
    padding: var(--padding);
    font-size: var(--fontSize);
  }
</style>
