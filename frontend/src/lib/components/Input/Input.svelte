<script lang="ts">
  import type { Snippet } from "svelte";
  import type { StyleConfig } from "$lib/types/style";
  import { styleVars } from "$lib/utils/styleVars";

  const defaultStyleConfig = {
    backgroundColor: "#f9fafb",
    focusBackgroundColor: "#ffffff",
    borderColor: "#e5e7eb",
    focusBorderColor: "#10b981",
    borderRadius: "0.5rem",
    borderWidth: "1px",
    padding: "0.5rem 0.75rem",
    gap: "0.5rem",
    fontSize: "0.875rem",
    height: "auto",
  };

  let {
    value = $bindable(),
    placeholder,
    type = "text",
    isLoading,
    icon,
    trailing,
    oninput = undefined,
    onfocus = undefined,
    onblur = undefined,
    styleConfig = {},
    class: className = "mb-3",
  }: {
    value: string;
    placeholder?: string;
    type?: string;
    isLoading?: boolean;
    icon?: Snippet;
    trailing?: Snippet;
    oninput?: () => void;
    onfocus?: () => void;
    onblur?: () => void;
    styleConfig?: StyleConfig;
    class?: string;
  } = $props();

  const mergedStyleConfig = $derived({ ...defaultStyleConfig, ...styleConfig });
</script>

<div
  class="inputWrapper flex items-center transition-colors {className}"
  style={styleVars(mergedStyleConfig)}
>
  {#if icon}
    {@render icon()}
  {/if}
  <input
    {type}
    {placeholder}
    bind:value
    {oninput}
    {onfocus}
    {onblur}
    class="flex-1 border-none bg-transparent text-gray-900 placeholder-gray-400 outline-none"
  />
  {#if isLoading}
    <div
      class="h-3.5 w-3.5 shrink-0 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
    ></div>
  {/if}
  {#if trailing}
    {@render trailing()}
  {/if}
</div>

<style>
  .inputWrapper {
    background-color: var(--backgroundColor);
    border: var(--borderWidth) solid var(--borderColor);
    border-radius: var(--borderRadius);
    padding: var(--padding);
    gap: var(--gap);
    height: var(--height);

    &:focus-within {
      background-color: var(--focusBackgroundColor);
      border-color: var(--focusBorderColor);
    }
  }

  .inputWrapper input {
    font-size: var(--fontSize);
    font-family: inherit;
  }
</style>
