<script lang="ts" generics="T">
  import type { Snippet } from "svelte";
  import Card from "$lib/components/Card/Card.svelte";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";

  let {
    items,
    getKey,
    row,
    detail,
  }: {
    items: T[];
    getKey: (item: T) => string;
    row: Snippet<[T]>;
    detail: Snippet<[T]>;
  } = $props();

  let expandedKeys = $state<Set<string>>(new Set());

  $effect(() => {
    items;
    expandedKeys = new Set();
  });

  const toggleExpand = (key: string) => {
    const next = new Set(expandedKeys);
    if (next.has(key)) next.delete(key);
    else next.add(key);
    expandedKeys = next;
  };
</script>

<Card styleConfig={{ padding: "0" }} class="overflow-hidden">
  {#each items as item}
    {@const key = getKey(item)}
    {@const expanded = expandedKeys.has(key)}
    <div class="border-b border-gray-100 last:border-b-0">
      <div
        class="flex cursor-pointer items-start gap-3 px-4 py-3 transition-colors hover:bg-gray-50"
        onclick={() => toggleExpand(key)}
        role="button"
        tabindex="0"
        onkeydown={(e) => e.key === "Enter" && toggleExpand(key)}
      >
        <ChevronRight
          size={13}
          class="mt-0.5 shrink-0 text-gray-400 transition-transform {expanded ? 'rotate-90' : ''}"
        />
        {@render row(item)}
      </div>
      {#if expanded}
        <div class="border-t border-gray-100 bg-gray-50 px-4 pb-4 pl-9">
          {@render detail(item)}
        </div>
      {/if}
    </div>
  {/each}
</Card>
