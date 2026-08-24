<script lang="ts">
  import { onMount } from "svelte";
  import { api, type DefaultLab } from "$lib/api";
  import { createStatusMessage } from "$lib/utils/statusMessage.svelte";
  import type { StyleConfig } from "$lib/types/style";
  import Card from "$lib/components/Card/Card.svelte";
  import Badge from "$lib/components/Badge/Badge.svelte";
  import Button from "$lib/components/Button/Button.svelte";
  import Notification from "$lib/components/Notification/Notification.svelte";
  import Square from "@lucide/svelte/icons/square";
  import Play from "@lucide/svelte/icons/play";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";
  import ExternalLink from "@lucide/svelte/icons/external-link";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import Bookmark from "@lucide/svelte/icons/bookmark";

  type Preset = {
    id: string;
    name: string;
    tags: string[];
    services: {
      name: string;
      image: string;
      version: string;
      ports: string[];
      env_vars: Record<string, string>;
    }[];
    created_at: string;
  };

  const TAG_COLORS: Record<string, { backgroundColor: string; color: string }> = {
    SQLi: { backgroundColor: "#E6F1FB", color: "#0C447C" },
    XSS: { backgroundColor: "#FBEAF0", color: "#993556" },
    RCE: { backgroundColor: "#FAECE7", color: "#993C1D" },
    LFI: { backgroundColor: "#FAEEDA", color: "#854F0B" },
    SSRF: { backgroundColor: "#E1F5EE", color: "#0F6E56" },
  };

  const DIFF_STYLES: Record<string, { backgroundColor: string; color: string }> = {
    apprentice: { backgroundColor: "#E1F5EE", color: "#0F6E56" },
    practitioner: { backgroundColor: "#FAEEDA", color: "#854F0B" },
    expert: { backgroundColor: "#FAECE7", color: "#993C1D" },
  };

  const defaultTagColor = { backgroundColor: "#E6F1FB", color: "#0C447C" };

  const countPillStyle: StyleConfig = {
    backgroundColor: "#f3f4f6",
    color: "#9ca3af",
    borderColor: "#e5e7eb",
    borderWidth: "1px",
    borderRadius: "9999px",
    padding: "0.125rem 0.5rem",
    fontSize: "10px",
  };

  const imageVersionChipStyle: StyleConfig = {
    backgroundColor: "#f3f4f6",
    color: "#6a7282",
    borderColor: "#e5e7eb",
    borderWidth: "1px",
    borderRadius: "0.25rem",
    padding: "0.125rem 0.375rem",
    fontSize: "10px",
  };

  let presets = $state<Preset[]>([]);
  let loading = $state(true);
  const status = createStatusMessage();
  let activeFilter = $state("All");
  let expandedIds = $state<Set<string>>(new Set());
  let composeYamls = $state<Record<string, string>>({});

  let defaultLabs = $state<DefaultLab[]>([]);
  let defaultLabsRunning = $state(false);
  let startingDefaultLabs = $state(false);
  let stoppingDefaultLabs = $state(false);
  let expandedDefaultIds = $state<Set<string>>(new Set());

  const ALL_TAGS = ["All", "SQLi", "XSS", "RCE", "LFI", "SSRF"];

  let filteredPresets = $derived(
    activeFilter === "All" ? presets : presets.filter((p) => p.tags.includes(activeFilter))
  );

  let availableTags = $derived(
    ALL_TAGS.filter((t) => t === "All" || presets.some((p) => p.tags.includes(t)))
  );

  const load = async () => {
    loading = true;
    try {
      const [presetsRes, defaultLabsRes, statusRes] = await Promise.allSettled([
        api.listPresets(),
        api.listDefaultLabs(),
        api.defaultLabStatus(),
      ]);

      if (presetsRes.status === "fulfilled") {
        presets = presetsRes.value.presets ?? [];
      }
      if (defaultLabsRes.status === "fulfilled") {
        defaultLabs = defaultLabsRes.value.labs ?? [];
      }
      if (statusRes.status === "fulfilled") {
        defaultLabsRunning = statusRes.value.running;
      }
    } catch (e) {
      status.error(e, "Error loading presets");
    } finally {
      loading = false;
    }
  };

  const toggleExpand = async (preset: Preset) => {
    const next = new Set(expandedIds);
    if (next.has(preset.id)) {
      next.delete(preset.id);
    } else {
      next.add(preset.id);
      if (!composeYamls[preset.id]) {
        try {
          const res = await api.generateCompose({
            name: preset.name,
            services: preset.services,
          });
          composeYamls = { ...composeYamls, [preset.id]: res.yaml };
        } catch {
          composeYamls = { ...composeYamls, [preset.id]: "Failed to generate preview." };
        }
      }
    }
    expandedIds = next;
  };

  const toggleDefaultExpand = (id: string) => {
    const next = new Set(expandedDefaultIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    expandedDefaultIds = next;
  };

  const startDefaultLabs = async () => {
    startingDefaultLabs = true;
    status.clear();
    try {
      await api.startDefaultLab();
      defaultLabsRunning = true;
      status.success("Default labs started successfully!");
    } catch (e) {
      status.error(e);
    } finally {
      startingDefaultLabs = false;
    }
  };

  const stopDefaultLabs = async () => {
    stoppingDefaultLabs = true;
    status.clear();
    try {
      await api.stopDefaultLab();
      defaultLabsRunning = false;
      status.success("Default labs stopped.");
    } catch (e) {
      status.error(e);
    } finally {
      stoppingDefaultLabs = false;
    }
  };

  const openLab = (path: string) => {
    window.open(`http://localhost:8888${path}`, "_blank");
  };

  const bootPreset = async (preset: Preset) => {
    status.clear();
    try {
      await api.startLab({ name: preset.name, services: preset.services });
      status.success(`"${preset.name}" started successfully!`);
    } catch (e) {
      status.error(e);
    }
  };

  const deletePreset = async (preset: Preset) => {
    if (!confirm(`Delete "${preset.name}"?`)) return;
    try {
      await api.deletePreset(preset.id);
      status.success(`"${preset.name}" deleted.`);
      await load();
    } catch (e) {
      status.error(e);
    }
  };

  const formatDate = (iso: string) => {
    return new Date(iso).toLocaleDateString("en-US", {
      month: "short",
      day: "numeric",
      year: "numeric",
    });
  };

  onMount(load);
</script>

<div class="flex flex-col gap-5">
  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-base font-medium text-gray-900">Presets</h1>
      <p class="mt-0.5 text-xs text-gray-400">
        Saved lab configurations — boot any environment in one click
      </p>
    </div>
  </div>

  {#if status.message}
    <Notification variant={status.variant}>
      {status.message}
    </Notification>
  {/if}

  <div>
    <div class="mb-3 flex items-center gap-2">
      <span class="text-xs font-medium text-gray-600">Default labs</span>
      <Badge styleConfig={countPillStyle}>
        {defaultLabs.length}
      </Badge>
      <div class="ml-auto flex items-center gap-2">
        {#if defaultLabsRunning}
          <div class="flex items-center gap-2">
            <div class="flex items-center gap-1.5">
              <div class="h-1.5 w-1.5 rounded-full bg-emerald-500"></div>
              <span class="text-[11px] text-emerald-700">running</span>
            </div>
            <Button onclick={stopDefaultLabs} disabled={stoppingDefaultLabs}>
              {#snippet icon()}
                {#if stoppingDefaultLabs}
                  <div
                    class="h-3 w-3 animate-spin rounded-full border-2 border-gray-300 border-t-gray-600"
                  ></div>
                {:else}
                  <Square size={11} />
                {/if}
              {/snippet}
              {stoppingDefaultLabs ? "Stopping..." : "Stop"}
            </Button>
          </div>
        {:else}
          <Button
            onclick={startDefaultLabs}
            disabled={startingDefaultLabs}
            styleConfig={{
              backgroundColor: "#009966",
              color: "#ffffff",
              hoverBackgroundColor: "#007a55",
              borderColor: "#009966",
            }}
          >
            {#snippet icon()}
              {#if startingDefaultLabs}
                <div
                  class="h-3 w-3 animate-spin rounded-full border-2 border-white/30 border-t-white"
                ></div>
              {:else}
                <Play size={11} />
              {/if}
            {/snippet}
            {startingDefaultLabs ? "Starting..." : "Boot default labs"}
          </Button>
        {/if}
      </div>
    </div>

    {#if loading}
      <Card>
        <div class="flex items-center gap-2 py-3 text-xs text-gray-400">
          <div
            class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
          ></div>
          Loading...
        </div>
      </Card>
    {:else if defaultLabs.length === 0}
      <Card>
        <p class="py-2 text-xs text-gray-400">No default labs found.</p>
      </Card>
    {:else}
      <Card styleConfig={{ padding: "0" }} class="overflow-hidden">
        {#each defaultLabs as lab}
          <div class="border-b border-gray-100 last:border-b-0">
            <div
              class="flex cursor-pointer items-center gap-3 px-4 py-3 transition-colors hover:bg-gray-50"
              onclick={() => toggleDefaultExpand(lab.id)}
              role="button"
              tabindex="0"
              onkeydown={(e) => e.key === "Enter" && toggleDefaultExpand(lab.id)}
            >
              <ChevronRight
                size={13}
                class="shrink-0 text-gray-400 transition-transform {expandedDefaultIds.has(lab.id)
                  ? 'rotate-90'
                  : ''}"
              />
              <div class="min-w-0 flex-1">
                <div class="mb-1 text-sm font-medium text-gray-900">{lab.title}</div>
                <div class="flex items-center gap-1.5">
                  <Badge class="font-mono" styleConfig={imageVersionChipStyle}>
                    vulnkit-labs:latest
                  </Badge>
                </div>
              </div>
              <Badge styleConfig={DIFF_STYLES[lab.difficulty] ?? DIFF_STYLES.apprentice}>
                {lab.difficulty}
              </Badge>
              <Badge styleConfig={TAG_COLORS[lab.category] ?? defaultTagColor}>
                {lab.category}
              </Badge>
              <div
                class="flex shrink-0 gap-1.5"
                onclick={(e) => e.stopPropagation()}
                role="presentation"
              >
                {#if defaultLabsRunning}
                  <Button
                    onclick={() => openLab(lab.path)}
                    title="Open lab in browser"
                    styleConfig={{
                      color: "#9ca3af",
                      hoverBackgroundColor: "#ecfdf5",
                      hoverBorderColor: "#a7f3d0",
                      hoverColor: "#047857",
                      width: "1.5rem",
                      height: "1.5rem",
                      padding: "0",
                      borderRadius: "0.25rem",
                    }}
                  >
                    {#snippet icon()}<ExternalLink size={10} />{/snippet}
                  </Button>
                {:else}
                  <Button
                    onclick={startDefaultLabs}
                    disabled={startingDefaultLabs}
                    title="Boot default labs first"
                    styleConfig={{
                      color: "#9ca3af",
                      hoverBackgroundColor: "#ecfdf5",
                      hoverBorderColor: "#a7f3d0",
                      hoverColor: "#047857",
                      width: "1.5rem",
                      height: "1.5rem",
                      padding: "0",
                      borderRadius: "0.25rem",
                    }}
                  >
                    {#snippet icon()}<Play size={10} />{/snippet}
                  </Button>
                {/if}
                <Button
                  disabled={true}
                  title="Default labs cannot be deleted"
                  styleConfig={{
                    color: "#9ca3af",
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

            {#if expandedDefaultIds.has(lab.id)}
              <div class="border-t border-gray-100 bg-gray-50 px-4 pb-4">
                <div
                  class="mt-3 mb-1 text-[10px] font-medium tracking-wider text-gray-400 uppercase"
                >
                  Description
                </div>
                <p class="mb-3 text-xs leading-relaxed text-gray-600">{lab.description}</p>
                <div class="mb-3 rounded-lg border border-gray-100 bg-white px-3 py-2.5">
                  <div class="mb-1 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
                    Objective
                  </div>
                  <p class="text-xs text-gray-800">{lab.objective}</p>
                </div>
                <div class="mb-1 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
                  Hint
                </div>
                <p class="mb-3 text-xs text-gray-500 italic">{lab.hint}</p>
                <div class="mb-1 text-[10px] font-medium tracking-wider text-gray-400 uppercase">
                  Solution
                </div>
                <div class="flex items-center gap-2">
                  <code
                    class="flex-1 rounded-lg border border-gray-100 bg-white px-3 py-2 font-mono text-[11px] text-gray-800"
                  >
                    {lab.solution}
                  </code>
                  {#if defaultLabsRunning}
                    <Button
                      onclick={() => openLab(lab.path)}
                      class="whitespace-nowrap"
                      styleConfig={{
                        backgroundColor: "#ecfdf5",
                        color: "#047857",
                        borderColor: "#a7f3d0",
                        hoverBackgroundColor: "#d1fae5",
                        padding: "0.5rem 0.75rem",
                      }}
                    >
                      {#snippet icon()}<ExternalLink size={11} />{/snippet}
                      Open lab
                    </Button>
                  {:else}
                    <Button
                      onclick={startDefaultLabs}
                      disabled={startingDefaultLabs}
                      class="whitespace-nowrap"
                      styleConfig={{
                        backgroundColor: "#059669",
                        color: "#ffffff",
                        borderColor: "#059669",
                        hoverBackgroundColor: "#047857",
                        padding: "0.5rem 0.75rem",
                      }}
                    >
                      {#snippet icon()}<Play size={11} />{/snippet}
                      Boot to open
                    </Button>
                  {/if}
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </Card>
    {/if}
  </div>

  <div>
    <div class="mb-3 flex items-center gap-2">
      <span class="text-xs font-medium text-gray-600">My presets</span>
      <Badge styleConfig={countPillStyle}>
        {presets.length}
      </Badge>
    </div>

    {#if presets.length > 0}
      <div class="mb-3 flex items-center gap-2">
        <span class="text-[11px] text-gray-400">Filter:</span>
        {#each availableTags as tag}
          <button
            onclick={() => (activeFilter = tag)}
            class="rounded-full border px-3 py-1 text-[11px] transition-colors
              {activeFilter === tag
              ? 'border-emerald-300 bg-emerald-50 font-medium text-emerald-700'
              : 'border-gray-200 bg-white text-gray-500 hover:bg-gray-50'}"
          >
            {tag}
          </button>
        {/each}
      </div>
    {/if}

    {#if loading}
      <Card>
        <div class="flex items-center gap-2 py-3 text-xs text-gray-400">
          <div
            class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
          ></div>
          Loading...
        </div>
      </Card>
    {:else if presets.length === 0}
      <Card>
        <div class="flex flex-col items-center gap-2 py-10">
          <Bookmark size={28} strokeWidth={1.5} class="text-gray-300" />
          <p class="text-sm text-gray-500">No presets saved yet</p>
          <p class="text-xs text-gray-400">Build a lab and save it as a preset to see it here</p>
        </div>
      </Card>
    {:else if filteredPresets.length === 0}
      <Card>
        <p class="py-2 text-xs text-gray-400">No presets match the selected filter.</p>
      </Card>
    {:else}
      <Card styleConfig={{ padding: "0" }} class="overflow-hidden">
        {#each filteredPresets as preset}
          <div class="border-b border-gray-100 last:border-b-0">
            <div
              class="flex cursor-pointer items-center gap-3 px-4 py-3 transition-colors hover:bg-gray-50"
              onclick={() => toggleExpand(preset)}
              role="button"
              tabindex="0"
              onkeydown={(e) => e.key === "Enter" && toggleExpand(preset)}
            >
              <ChevronRight
                size={13}
                class="shrink-0 text-gray-400 transition-transform {expandedIds.has(preset.id)
                  ? 'rotate-90'
                  : ''}"
              />
              <div class="min-w-0 flex-1">
                <div class="mb-1 text-sm font-medium text-gray-900">{preset.name}</div>
                <div class="flex flex-wrap items-center gap-1.5">
                  {#each preset.services as svc}
                    <Badge class="font-mono" styleConfig={imageVersionChipStyle}>
                      {svc.image}:{svc.version}
                    </Badge>
                  {/each}
                </div>
              </div>
              <div class="flex shrink-0 items-center gap-1.5">
                {#each preset.tags ?? [] as tag}
                  <Badge styleConfig={TAG_COLORS[tag] ?? defaultTagColor}>{tag}</Badge>
                {/each}
              </div>
              <span class="w-24 shrink-0 text-right text-[11px] text-gray-400">
                {formatDate(preset.created_at)}
              </span>
              <div
                class="flex shrink-0 gap-1.5"
                onclick={(e) => e.stopPropagation()}
                role="presentation"
              >
                <Button
                  onclick={() => bootPreset(preset)}
                  title="Boot preset"
                  styleConfig={{
                    color: "#9ca3af",
                    hoverBackgroundColor: "#ecfdf5",
                    hoverBorderColor: "#a7f3d0",
                    hoverColor: "#047857",
                    width: "1.5rem",
                    height: "1.5rem",
                    padding: "0",
                    borderRadius: "0.25rem",
                  }}
                >
                  {#snippet icon()}<Play size={10} />{/snippet}
                </Button>
                <Button
                  onclick={() => deletePreset(preset)}
                  title="Delete preset"
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

            {#if expandedIds.has(preset.id)}
              <div class="border-t border-gray-100 bg-gray-50 px-4 pb-4">
                <div
                  class="mt-3 mb-2 text-[10px] font-medium tracking-wider text-gray-400 uppercase"
                >
                  docker-compose preview
                </div>
                {#if composeYamls[preset.id]}
                  <pre
                    class="overflow-x-auto rounded-lg border border-gray-100 bg-white p-3 font-mono text-[10px] leading-relaxed whitespace-pre text-gray-500">{composeYamls[
                      preset.id
                    ]}</pre>
                {:else}
                  <div class="flex items-center gap-2 text-xs text-gray-400">
                    <div
                      class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
                    ></div>
                    Generating preview...
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </Card>
    {/if}
  </div>
</div>
