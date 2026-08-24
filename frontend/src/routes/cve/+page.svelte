<script lang="ts">
  import { api, type CVEItem } from "$lib/api";
  import { versionToNumber } from "$lib/utils/version";
  import type { StyleConfig } from "$lib/types/style";
  import Card from "$lib/components/Card/Card.svelte";
  import Badge from "$lib/components/Badge/Badge.svelte";
  import Input from "$lib/components/Input/Input.svelte";
  import Notification from "$lib/components/Notification/Notification.svelte";
  import Pagination from "$lib/components/Pagination/Pagination.svelte";
  import PaginationSummary from "$lib/components/Pagination/PaginationSummary.svelte";
  import DetailedTable from "$lib/components/Table/DetailedTable/DetailedTable.svelte";
  import Shield from "@lucide/svelte/icons/shield";
  import Search from "@lucide/svelte/icons/search";
  import ExternalLink from "@lucide/svelte/icons/external-link";
  import Wrench from "@lucide/svelte/icons/wrench";
  import TriangleAlert from "@lucide/svelte/icons/triangle-alert";

  const PAGE_SIZE = 20;
  const DEBOUNCE_MS = 600;
  const MIN_QUERY_LENGTH = 3;

  const SEVERITY_STYLES: Record<string, { backgroundColor: string; color: string }> = {
    CRITICAL: { backgroundColor: "#FCEBEB", color: "#A32D2D" },
    HIGH: { backgroundColor: "#FAECE7", color: "#993C1D" },
    MEDIUM: { backgroundColor: "#FAEEDA", color: "#854F0B" },
    LOW: { backgroundColor: "#EAF3DE", color: "#3B6D11" },
    NONE: { backgroundColor: "#F1EFE8", color: "#5F5E5A" },
  };

  const versionChipStyle: StyleConfig = {
    backgroundColor: "#ffffff",
    color: "#4b5563",
    borderColor: "#e5e7eb",
    borderWidth: "1px",
    borderRadius: "0.25rem",
    padding: "0.125rem 0.5rem",
    fontSize: "10px",
  };

  let searchQuery = $state("");
  let results = $state<{ cve: CVEItem }[]>([]);
  let totalResults = $state(0);
  let currentPage = $state(1);
  let loading = $state(false);
  let searchError = $state("");
  let hasSearched = $state(false);

  let debounceTimer: ReturnType<typeof setTimeout>;
  let abortController: AbortController | null = null;

  let totalPages = $derived(Math.ceil(totalResults / PAGE_SIZE));

  const onSearchInput = () => {
    clearTimeout(debounceTimer);
    abortController?.abort();
    searchError = "";
    if (searchQuery.trim().length < MIN_QUERY_LENGTH) {
      results = [];
      totalResults = 0;
      hasSearched = false;
      loading = false;
      return;
    }
    debounceTimer = setTimeout(() => {
      currentPage = 1;
      search();
    }, DEBOUNCE_MS);
  };

  const search = async () => {
    abortController?.abort();
    abortController = new AbortController();
    loading = true;
    searchError = "";
    hasSearched = true;
    try {
      const startIndex = (currentPage - 1) * PAGE_SIZE;
      const data = await api.cveSearch(searchQuery, startIndex, PAGE_SIZE, abortController.signal);
      results = data.vulnerabilities ?? [];
      totalResults = data.totalResults ?? 0;
    } catch (e: any) {
      if (e.name === "AbortError") return;
      searchError = "Failed to reach NVD API. Try again later.";
      results = [];
      totalResults = 0;
    } finally {
      loading = false;
    }
  };

  const goToPage = (page: number) => {
    if (page < 1 || page > totalPages || page === currentPage) return;
    currentPage = page;
    search();
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const getSeverity = (cve: CVEItem): { label: string; score: number | null } => {
    const v31 = cve.metrics?.cvssMetricV31?.[0];
    if (v31)
      return {
        label: v31.cvssData.baseSeverity,
        score: v31.cvssData.baseScore,
      };
    const v2 = cve.metrics?.cvssMetricV2?.[0];
    if (v2) return { label: v2.baseSeverity, score: v2.cvssData.baseScore };
    return { label: "NONE", score: null };
  };

  const getDescription = (cve: CVEItem): string =>
    cve.descriptions.find((d) => d.lang === "en")?.value ?? "No description available.";

  const getExploitDbLink = (cve: CVEItem): string | null => {
    const ref = cve.references.find((r) => r.url.includes("exploit-db.com/exploits/"));
    return ref?.url ?? null;
  };

  const hasExploit = (cve: CVEItem): boolean =>
    cve.references.some((r) => r.url.includes("exploit-db.com/exploits/"));

  type AffectedItem = {
    product: string;
    image: string;
    version: string | null;
    versionEnd: string | null;
    versionEndType: "including" | "excluding" | null;
  };

  const CPE_TO_IMAGE: Record<string, string> = {
    http_server: "httpd",
    mysql: "mysql",
    mysql_server: "mysql",
    nginx: "nginx",
    nginx_server: "nginx",
    postgresql: "postgres",
    redis: "redis",
    redis_server: "redis",
    mongodb: "mongo",
    mariadb: "mariadb",
    php: "php",
  };

  const getAffected = (cve: CVEItem): AffectedItem[] => {
    const items: AffectedItem[] = [];
    try {
      const configurations = cve.configurations ?? [];
      for (const config of configurations) {
        const nodes = config.nodes ?? [];
        for (const node of nodes) {
          for (const match of node.cpeMatch ?? []) {
            if (!match.vulnerable || !match.criteria) continue;
            const parts = match.criteria.split(":");
            if (parts.length < 5) continue;
            const product = parts[4];
            const rawVersion = parts[5] !== "*" ? parts[5] : null;
            const image = CPE_TO_IMAGE[product] ?? product;
            items.push({
              product,
              image,
              version: rawVersion,
              versionEnd: match.versionEndIncluding ?? match.versionEndExcluding ?? null,
              versionEndType: match.versionEndIncluding
                ? "including"
                : match.versionEndExcluding
                  ? "excluding"
                  : null,
            });
          }
        }
      }
    } catch {}
    const seen = new Set<string>();
    return items
      .filter((item) => {
        const key = `${item.image}-${item.version ?? item.versionEnd}`;
        if (seen.has(key)) return false;
        seen.add(key);
        return true;
      })
      .slice(0, 6);
  };

  type VersionAvailability = {
    available: boolean;
    suggested: string | null;
    dynamicSupported: boolean;
  };

  const versionCache = new Map<string, VersionAvailability>();

  const isVersionTag = (tag: string): boolean => {
    return /^\d+\.\d+/.test(tag);
  };

  const checkVersionAvailability = async (
    image: string,
    version: string
  ): Promise<VersionAvailability> => {
    const key = `${image}:${version}`;
    if (versionCache.has(key)) return versionCache.get(key)!;

    try {
      const [tagsData, dynamicCheck] = await Promise.allSettled([
        api.hubTags(image),
        api.dynamicBuildCheck(image, version),
      ]);

      const tags = (
        tagsData.status === "fulfilled"
          ? (tagsData.value.results?.map((t: { name: string }) => t.name) ?? [])
          : []
      ).filter(isVersionTag);

      const dynamicSupported =
        dynamicCheck.status === "fulfilled" ? dynamicCheck.value.supported : false;

      if (tags.includes(version)) {
        const result = {
          available: true,
          suggested: null,
          dynamicSupported: false,
        };
        versionCache.set(key, result);
        return result;
      }

      const targetNum = versionToNumber(version);
      const parts = version.split(".");

      const sameMinor = tags
        .filter((t) => {
          const tParts = t.split(".");
          return tParts[0] === parts[0] && tParts[1] === parts[1] && versionToNumber(t) < targetNum;
        })
        .sort((a, b) => versionToNumber(b) - versionToNumber(a));

      const sameMajor = tags
        .filter((t) => {
          const tParts = t.split(".");
          return tParts[0] === parts[0] && versionToNumber(t) < targetNum;
        })
        .sort((a, b) => versionToNumber(b) - versionToNumber(a));

      const suggested = sameMinor[0] ?? sameMajor[0] ?? null;

      const suggestedNum = suggested ? versionToNumber(suggested) : 0;
      const finalSuggested = suggestedNum < targetNum ? suggested : null;

      const result = {
        available: false,
        suggested: finalSuggested,
        dynamicSupported,
      };
      versionCache.set(key, result);
      return result;
    } catch {
      return { available: true, suggested: null, dynamicSupported: false };
    }
  };

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });

  const labBuilderUrl = (image: string, version: string) =>
    `/lab/lab-builder/?image=${encodeURIComponent(image)}&version=${encodeURIComponent(version)}`;
</script>

<div class="flex flex-col gap-5">
  <div>
    <h1 class="text-base font-medium text-gray-900">CVE Lookup</h1>
    <p class="mt-0.5 text-xs text-gray-400">Search vulnerabilities by software name or CVE ID</p>
  </div>

  <Input
    bind:value={searchQuery}
    placeholder="e.g. mysql, apache, CVE-2021-41773..."
    isLoading={loading}
    oninput={onSearchInput}
    class=""
  >
    {#snippet icon()}
      <Shield class="shrink-0 text-gray-400" size={15} />
    {/snippet}
    {#snippet trailing()}
      {#if !loading && searchQuery.length > 0 && searchQuery.length < MIN_QUERY_LENGTH}
        <span class="shrink-0 text-[10px] text-gray-400">
          {MIN_QUERY_LENGTH - searchQuery.length} more char{MIN_QUERY_LENGTH -
            searchQuery.length !==
          1
            ? "s"
            : ""}...
        </span>
      {/if}
    {/snippet}
  </Input>

  {#if searchError}
    <Notification variant="error">
      {searchError}
    </Notification>
  {/if}

  {#if hasSearched && !loading && totalResults > 0}
    <PaginationSummary
      {searchQuery}
      {currentPage}
      {totalPages}
      pageSize={PAGE_SIZE}
      {totalResults}
    />
  {/if}

  {#if results.length === 0 && !loading && !searchError}
    <Card>
      <div class="flex flex-col items-center gap-2 py-10">
        {#if hasSearched}
          <Shield size={28} strokeWidth={1.5} class="text-gray-300" />
          <p class="text-sm text-gray-500">No CVEs found for "{searchQuery}"</p>
          <p class="text-xs text-gray-400">Try a different search term</p>
        {:else}
          <Search size={28} strokeWidth={1.5} class="text-gray-300" />
          <p class="text-sm text-gray-500">Search for a software or CVE ID</p>
          <p class="text-xs text-gray-400">
            min. {MIN_QUERY_LENGTH} characters — e.g. mysql, apache, nginx
          </p>
        {/if}
      </div>
    </Card>
  {:else if results.length > 0}
    <DetailedTable items={results} getKey={({ cve }) => cve.id}>
      {#snippet row({ cve })}
        {@const severity = getSeverity(cve)}
        {@const description = getDescription(cve)}
        <div class="min-w-0 flex-1">
          <div class="mb-1 flex items-center gap-2">
            <span class="font-mono text-xs font-medium text-gray-900">{cve.id}</span>
            {#if hasExploit(cve)}
              <Badge
                class="font-medium"
                styleConfig={{
                  backgroundColor: "#fef2f2",
                  color: "#dc2626",
                  borderColor: "#fecaca",
                  borderWidth: "1px",
                  borderRadius: "0.25rem",
                  padding: "0.125rem 0.375rem",
                  fontSize: "9px",
                }}
              >
                EXPLOIT
              </Badge>
            {/if}
          </div>
          <p class="line-clamp-2 text-xs leading-relaxed text-gray-500">
            {description}
          </p>
          <span class="mt-1 block text-[10px] text-gray-400">{formatDate(cve.published)}</span>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <Badge styleConfig={SEVERITY_STYLES[severity.label] ?? SEVERITY_STYLES.NONE}>
            {severity.label}
          </Badge>
          {#if severity.score !== null}
            <span class="w-7 text-right text-xs font-medium text-gray-600">{severity.score}</span>
          {/if}
        </div>
      {/snippet}

      {#snippet detail({ cve })}
        {@const description = getDescription(cve)}
        {@const exploitLink = getExploitDbLink(cve)}
        {@const affected = getAffected(cve)}
        <div class="mt-3 mb-1 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
          Description
        </div>
        <p class="mb-3 text-xs leading-relaxed text-gray-600">
          {description}
        </p>

        {#if affected.length > 0}
          <div class="mb-2 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
            Affected versions
          </div>
          <div class="mb-3 flex flex-col gap-2">
            {#each affected as item}
              {#if item.version || item.versionEnd}
                {@const targetVersion = item.version ?? item.versionEnd ?? ""}
                {#await checkVersionAvailability(item.image, targetVersion)}
                  <div class="flex items-center gap-2">
                    <Badge class="font-mono" styleConfig={versionChipStyle}>
                      {item.image}:{targetVersion}
                    </Badge>
                    <div
                      class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
                    ></div>
                  </div>
                {:then availability}
                  <div class="flex flex-wrap items-center gap-2">
                    <Badge class="font-mono" styleConfig={versionChipStyle}>
                      {item.image}:{targetVersion}
                    </Badge>

                    {#if availability.available}
                      <a
                        href={labBuilderUrl(item.image, targetVersion)}
                        onclick={(e) => e.stopPropagation()}
                        class="inline-flex items-center gap-1 rounded border border-emerald-200 bg-emerald-50 px-2 py-0.5 text-[10px] whitespace-nowrap text-emerald-700 hover:bg-emerald-100"
                      >
                        <ExternalLink size={9} />
                        Open in Lab Builder
                      </a>
                    {:else}
                      <Badge
                        class="inline-flex items-center gap-1"
                        styleConfig={{
                          backgroundColor: "#fffbeb",
                          color: "#b45309",
                          borderColor: "#fde68a",
                          borderWidth: "1px",
                          borderRadius: "0.25rem",
                          padding: "0.125rem 0.5rem",
                          fontSize: "10px",
                        }}
                      >
                        <TriangleAlert size={10} />
                        not on Docker Hub
                      </Badge>

                      {#if availability.dynamicSupported}
                        <a
                          href={labBuilderUrl(item.image, targetVersion)}
                          onclick={(e) => e.stopPropagation()}
                          class="inline-flex items-center gap-1 rounded border border-amber-200 bg-amber-50 px-2 py-0.5 text-[10px] whitespace-nowrap text-amber-700 hover:bg-amber-100"
                        >
                          <Wrench size={9} />
                          Build dynamically in Lab Builder
                        </a>
                      {:else if availability.suggested}
                        <span class="text-[10px] text-gray-400">→ try</span>
                        <a
                          href={labBuilderUrl(item.image, availability.suggested)}
                          onclick={(e) => e.stopPropagation()}
                          class="inline-flex items-center gap-1 rounded border border-emerald-200 bg-emerald-50 px-2 py-0.5 font-mono text-[10px] text-emerald-700 hover:bg-emerald-100"
                        >
                          {item.image}:{availability.suggested}
                          <ExternalLink size={9} />
                        </a>
                      {/if}
                    {/if}
                  </div>
                {:catch}
                  <Badge class="font-mono" styleConfig={versionChipStyle}>
                    {item.image}:{targetVersion}
                  </Badge>
                {/await}
              {:else}
                <Badge class="font-mono" styleConfig={versionChipStyle}>
                  {item.image}
                  {item.versionEnd
                    ? `${item.versionEndType === "including" ? "≤" : "<"} ${item.versionEnd}`
                    : ""}
                </Badge>
              {/if}
            {/each}
          </div>
        {/if}

        <div class="mb-2 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
          References
        </div>
        <div class="mb-3 flex flex-col gap-1.5">
          {#each cve.references.slice(0, 5) as ref}
            <div class="flex items-center gap-2">
              {#if ref.tags && ref.tags.length > 0}
                <Badge
                  class="shrink-0"
                  styleConfig={{
                    backgroundColor: "#ffffff",
                    color: "#9ca3af",
                    borderColor: "#e5e7eb",
                    borderWidth: "1px",
                    borderRadius: "0.25rem",
                    padding: "0.125rem 0.375rem",
                    fontSize: "9px",
                  }}
                >
                  {ref.tags[0]}
                </Badge>
              {/if}
              <a
                href={ref.url}
                target="_blank"
                rel="noopener noreferrer"
                onclick={(e) => e.stopPropagation()}
                class="truncate text-[11px] text-blue-600 hover:underline"
              >
                {ref.url}
              </a>
            </div>
          {/each}
        </div>

        {#if exploitLink}
          <a
            href={exploitLink}
            target="_blank"
            rel="noopener noreferrer"
            onclick={(e) => e.stopPropagation()}
            class="inline-flex items-center gap-1.5 rounded-lg border border-red-200 bg-red-50 px-3 py-1.5 text-xs font-medium text-red-700 hover:bg-red-100"
          >
            <ExternalLink size={11} />
            View on Exploit-DB
          </a>
        {/if}
      {/snippet}
    </DetailedTable>

    {#if totalPages > 1}
      <Pagination
        {currentPage}
        {totalPages}
        pageSize={PAGE_SIZE}
        {totalResults}
        onPageChange={goToPage}
      />
    {/if}
  {/if}
</div>
