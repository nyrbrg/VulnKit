<script lang="ts">
  import { onMount } from "svelte";
  import { api, type DynamicBuild } from "$lib/api";
  import { createStatusMessage } from "$lib/utils/statusMessage.svelte";
  import Card from "$lib/components/Card/Card.svelte";
  import Badge from "$lib/components/Badge/Badge.svelte";
  import Notification from "$lib/components/Notification/Notification.svelte";
  import Button from "$lib/components/Button/Button.svelte";
  import Database from "@lucide/svelte/icons/database";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";
  import ExternalLink from "@lucide/svelte/icons/external-link";
  import Trash2 from "@lucide/svelte/icons/trash-2";

  let builds = $state<DynamicBuild[]>([]);
  let loading = $state(true);
  const status = createStatusMessage();
  let expandedIds = $state<Set<string>>(new Set());
  let dockerfileContents = $state<Record<string, string>>({});

  const load = async () => {
    loading = true;
    try {
      const res = await api.listDynamicBuilds();
      builds = res.builds ?? [];
    } catch (e) {
      status.error(e, "Error loading builds");
    } finally {
      loading = false;
    }
  };

  const toggleExpand = async (build: DynamicBuild) => {
    const next = new Set(expandedIds);
    if (next.has(build.id)) {
      next.delete(build.id);
    } else {
      next.add(build.id);
      if (!(build.id in dockerfileContents)) {
        try {
          const res = await api.getBuildDockerfile(build.id);
          dockerfileContents = {
            ...dockerfileContents,
            [build.id]: res.dockerfile,
          };
        } catch {
          dockerfileContents = {
            ...dockerfileContents,
            [build.id]: "Failed to load Dockerfile.",
          };
        }
      }
    }
    expandedIds = next;
  };

  const deleteBuild = async (build: DynamicBuild) => {
    if (
      !confirm(
        `Delete "${build.image_name}"?\nThis will remove the Docker image and Dockerfile from disk.`
      )
    )
      return;
    try {
      await api.deleteDynamicBuild(build.id);
      status.success(`"${build.image_name}" deleted.`);
      await load();
    } catch (e) {
      status.error(e);
    }
  };

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });

  onMount(load);
</script>

<div class="flex flex-col gap-5">
  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-base font-medium text-gray-900">Local images</h1>
      <p class="mt-0.5 text-xs text-gray-400">
        Dynamically built Docker images — manage and inspect Dockerfiles
      </p>
    </div>
    {#if builds.length > 0}
      <Badge
        styleConfig={{
          backgroundColor: "#f3f4f6",
          color: "#6a7282",
          borderColor: "#e5e7eb",
          borderWidth: "1px",
          borderRadius: "0.5rem",
          padding: "0.25rem 0.5rem",
        }}
      >
        {builds.length} image{builds.length !== 1 ? "s" : ""}
      </Badge>
    {/if}
  </div>

  {#if status.message}
    <Notification variant={status.variant}>
      {status.message}
    </Notification>
  {/if}

  {#if loading}
    <Card>
      <div class="flex items-center gap-2 py-4 text-xs text-gray-400">
        <div
          class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
        ></div>
        Loading...
      </div>
    </Card>
  {:else if builds.length === 0}
    <Card>
      <div class="flex flex-col items-center gap-3 py-12">
        <Database size={32} strokeWidth={1.5} class="text-gray-300" />
        <p class="text-sm text-gray-500">No local images yet</p>
        <p class="text-xs text-gray-400">
          Build a dynamic image from the Lab Builder to see it here
        </p>
      </div>
    </Card>
  {:else}
    <Card styleConfig={{ padding: "0" }} class="overflow-hidden">
      {#each builds as build}
        <div class="border-b border-gray-100 last:border-b-0">
          <div
            class="flex cursor-pointer items-center gap-3 px-4 py-3 transition-colors hover:bg-gray-50"
            onclick={() => toggleExpand(build)}
            role="button"
            tabindex="0"
            onkeydown={(e) => e.key === "Enter" && toggleExpand(build)}
          >
            <ChevronRight
              size={13}
              class="shrink-0 text-gray-400 transition-transform {expandedIds.has(build.id)
                ? 'rotate-90'
                : ''}"
            />

            <div class="min-w-0 flex-1">
              <div class="mb-1 flex items-center gap-2">
                <span class="font-mono text-sm font-medium text-gray-900">{build.image_name}</span>
                <Badge
                  styleConfig={{
                    backgroundColor: "#fffbeb",
                    color: "#b45309",
                    borderColor: "#fde68a",
                    borderWidth: "1px",
                    borderRadius: "0.25rem",
                    padding: "0.125rem 0.375rem",
                    fontSize: "9px",
                  }}>dynamic</Badge
                >
              </div>
              <div class="flex items-center gap-1.5">
                <Badge
                  class="font-mono"
                  styleConfig={{
                    backgroundColor: "#f3f4f6",
                    color: "#6a7282",
                    borderColor: "#e5e7eb",
                    borderWidth: "1px",
                    borderRadius: "0.25rem",
                    padding: "0.125rem 0.375rem",
                    fontSize: "10px",
                  }}
                >
                  {build.software}:{build.version}
                  {#if build.requested_version !== build.version}
                    <span class="text-gray-400">(requested: {build.requested_version})</span>
                  {/if}
                </Badge>
              </div>
            </div>

            <span class="shrink-0 text-[11px] text-gray-400">{formatDate(build.created_at)}</span>

            <div
              class="flex shrink-0 gap-1.5"
              onclick={(e) => e.stopPropagation()}
              role="presentation"
            >
              <a
                href={`/lab/lab-builder?image=${encodeURIComponent(build.image_name)}&version=${encodeURIComponent(build.version)}`}
                title="Load in Lab Builder"
                class="flex h-6 w-6 items-center justify-center rounded border border-gray-200 text-gray-400 hover:border-emerald-200 hover:bg-emerald-50 hover:text-emerald-700"
              >
                <ExternalLink size={10} />
              </a>
              <Button
                onclick={() => deleteBuild(build)}
                title="Delete image and Dockerfile"
                styleConfig={{
                  color: "#9ca3af",
                  hoverBackgroundColor: "#fef2f2",
                  hoverBorderColor: "#fecaca",
                  hoverColor: "#dc2626",
                  width: "1.5rem",
                  height: "1.5rem",
                  padding: "0",
                  borderRadius: "0.25rem",
                }}
              >
                {#snippet icon()}<Trash2 size={10} />{/snippet}
              </Button>
            </div>
          </div>

          {#if expandedIds.has(build.id)}
            <div class="border-t border-gray-100 bg-gray-50 px-4 pb-4">
              <div class="mt-3 mb-2 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
                Dockerfile
              </div>
              {#if build.id in dockerfileContents}
                <pre
                  class="overflow-x-auto rounded-lg border border-gray-100 bg-white p-3 font-mono text-[10px] leading-relaxed whitespace-pre text-gray-500">{dockerfileContents[
                    build.id
                  ]}</pre>
              {:else}
                <div class="flex items-center gap-2 text-xs text-gray-400">
                  <div
                    class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
                  ></div>
                  Loading Dockerfile...
                </div>
              {/if}
              <div class="mt-2 text-[10px] text-gray-400">
                Saved at: <span class="font-mono">{build.build_dir}</span>
              </div>
            </div>
          {/if}
        </div>
      {/each}
    </Card>
  {/if}
</div>
