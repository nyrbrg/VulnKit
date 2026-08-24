<script lang="ts">
  import type { StyleConfig } from "$lib/types/style";
  import type { Snippet } from "svelte";
  import { styleVars } from "$lib/utils/styleVars";

  const defaultStyleConfig = {
    backgroundColor: "#ffffff",
    color: "inherit",
    borderColor: "#f3f4f6",
    borderWidth: "1px",
    borderRadius: "0.75rem",
    padding: "1rem",
    boxShadow: "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)",
  };

  let {
    children,
    styleConfig = {},
    class: className = "",
  }: { children: Snippet; styleConfig?: StyleConfig; class?: string } = $props();

  const mergedStyleConfig = $derived({ ...defaultStyleConfig, ...styleConfig });
</script>

<section
  class={className}
  style={styleVars(mergedStyleConfig)}
>
  {@render children()}
</section>

<style>
  section {
    background-color: var(--backgroundColor);
    color: var(--color);
    border: var(--borderWidth) solid var(--borderColor);
    border-radius: var(--borderRadius);
    padding: var(--padding);
    box-shadow: var(--boxShadow);
  }
</style>
