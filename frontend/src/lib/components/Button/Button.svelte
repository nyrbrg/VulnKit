<script lang="ts">
  import type { StyleConfig } from "$lib/types/style";
  import type { Snippet } from "svelte";
  import { styleVars } from "$lib/utils/styleVars";

  const defaultStyleConfig = {
    backgroundColor: "#ffffff",
    hoverBackgroundColor: "#f9fafb",
    color: "#4a5565",
    hoverColor: "#4a5565",
    borderColor: "#e5e7eb",
    hoverBorderColor: "#e5e7eb",
    borderRadius: "0.5rem",
    borderWidth: "1px",
    padding: "0.375rem 0.75rem",
    boxShadow: "none",
    width: "auto",
    height: "auto",
    fontSize: "0.75rem",
    gap: "0.375rem",
  };

  let {
    disabled,
    onclick = undefined,
    title = undefined,
    icon,
    children,
    styleConfig = {},
    class: className = "",
  }: {
    disabled?: boolean;
    icon?: Snippet;
    onclick?: () => void;
    title?: string;
    children?: Snippet;
    styleConfig?: StyleConfig;
    class?: string;
  } = $props();

  const mergedStyleConfig = $derived({ ...defaultStyleConfig, ...styleConfig });
</script>

<button
  {onclick}
  {disabled}
  {title}
  class="customButtom inline-flex items-center justify-center font-medium disabled:cursor-not-allowed disabled:opacity-40 {className}"
  style={styleVars(mergedStyleConfig)}
>
  {#if icon}
    {@render icon()}
  {/if}
  {#if children}
    {@render children()}
  {/if}
</button>

<style>
  .customButtom {
    background-color: var(--backgroundColor);
    color: var(--color);
    border: var(--borderWidth) solid var(--borderColor);
    border-radius: var(--borderRadius);
    padding: var(--padding);
    box-shadow: var(--boxShadow);
    width: var(--width);
    height: var(--height);
    font-size: var(--fontSize);
    gap: var(--gap);

    &:hover {
      background-color: var(--hoverBackgroundColor);
      color: var(--hoverColor);
      border-color: var(--hoverBorderColor);
    }
  }
</style>
