<script lang="ts">
  import Button from "$lib/components/Button/Button.svelte";
  import ChevronLeft from "@lucide/svelte/icons/chevron-left";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";

  let {
    currentPage,
    totalPages,
    pageSize,
    totalResults,
    onPageChange,
  }: {
    currentPage: number;
    totalPages: number;
    pageSize: number;
    totalResults: number;
    onPageChange: (page: number) => void;
  } = $props();

  const arrowStyle = {
    color: "#6b7280",
    hoverBackgroundColor: "#f9fafb",
    width: "1.75rem",
    height: "1.75rem",
    padding: "0",
    borderRadius: "0.25rem",
  };

  let pages = $derived.by((): (number | "...")[] => {
    if (totalPages <= 7) {
      return Array.from({ length: totalPages }, (_, i) => i + 1);
    }
    const result: (number | "...")[] = [];
    const delta = 2;
    const left = currentPage - delta;
    const right = currentPage + delta;
    result.push(1);
    if (left > 2) result.push("...");
    for (let i = Math.max(2, left); i <= Math.min(totalPages - 1, right); i++) {
      result.push(i);
    }
    if (right < totalPages - 1) result.push("...");
    if (totalPages > 1) result.push(totalPages);
    return result;
  });
</script>

<div class="flex items-center justify-between">
  <span class="text-[11px] text-gray-400">
    Showing {(currentPage - 1) * pageSize + 1}-{Math.min(currentPage * pageSize, totalResults)} of {totalResults.toLocaleString()}
  </span>
  <div class="flex items-center gap-1">
    <Button
      onclick={() => onPageChange(currentPage - 1)}
      disabled={currentPage === 1}
      styleConfig={arrowStyle}
    >
      {#snippet icon()}<ChevronLeft size={12} />{/snippet}
    </Button>

    {#each pages as p}
      {#if p === "..."}
        <span class="flex h-7 w-7 items-center justify-center text-xs text-gray-400">...</span>
      {:else}
        <button
          onclick={() => onPageChange(p as number)}
          class="h-7 w-7 rounded border text-xs font-medium transition-colors
            {currentPage === p
            ? 'border-emerald-300 bg-emerald-50 text-emerald-700'
            : 'border-gray-200 text-gray-500 hover:bg-gray-50'}"
        >
          {p}
        </button>
      {/if}
    {/each}

    <Button
      onclick={() => onPageChange(currentPage + 1)}
      disabled={currentPage === totalPages}
      styleConfig={arrowStyle}
    >
      {#snippet icon()}<ChevronRight size={12} />{/snippet}
    </Button>
  </div>
</div>
