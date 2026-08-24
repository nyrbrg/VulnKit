<script lang="ts">
  let {
    searchQuery,
    currentPage,
    totalPages,
    pageSize,
    totalResults,
  }: {
    searchQuery: string;
    currentPage: number;
    totalPages: number;
    pageSize: number;
    totalResults: number;
  } = $props();
</script>

<div class="flex items-center justify-between">
  <span class="text-[11px] text-gray-400">
    {totalResults.toLocaleString()} results for "{searchQuery}" — page {currentPage} of {totalPages}
  </span>
  <span class="text-[11px] text-gray-400">
    Showing {(currentPage - 1) * pageSize + 1}–{Math.min(currentPage * pageSize, totalResults)}
  </span>
</div>
