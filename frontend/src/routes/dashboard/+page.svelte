<script lang="ts">
  import { onMount } from "svelte";
  import { api, type ContainerStatus } from "$lib/api";
  import { createStatusMessage } from "$lib/utils/statusMessage.svelte";
  import Card from "$lib/components/Card/Card.svelte";
  import Badge from "$lib/components/Badge/Badge.svelte";
  import Button from "$lib/components/Button/Button.svelte";
  import Notification from "$lib/components/Notification/Notification.svelte";
  import type { StyleConfig } from "$lib/types/style";
  import RefreshCw from "@lucide/svelte/icons/refresh-cw";
  import Container from "@lucide/svelte/icons/container";
  import Square from "@lucide/svelte/icons/square";
  import Terminal from "@lucide/svelte/icons/terminal";
  import Play from "@lucide/svelte/icons/play";
  import Trash2 from "@lucide/svelte/icons/trash-2";
  import Bookmark from "@lucide/svelte/icons/bookmark";
  import Shield from "@lucide/svelte/icons/shield";

  const statTileStyle: StyleConfig = {
    backgroundColor: "#f9fafb",
    borderRadius: "0.5rem",
    padding: "0.75rem",
    boxShadow: "none",
  };

  type Preset = {
    id: string;
    name: string;
    tags: string[];
    services: any[];
  };

  let containers = $state<ContainerStatus[]>([]);
  let presets = $state<Preset[]>([]);
  let dockerOnline = $state(false);
  let dbOnline = $state(false);
  let hubOnline = $state(false);
  let loading = $state(true);
  let stoppingAll = $state(false);
  const status = createStatusMessage();
  let expandedLogsId = $state<string | null>(null);
  let logsContent = $state<Record<string, string>>({});

  let runningCount = $derived(containers.filter((c) => c.state === "running").length);
  let stoppedCount = $derived(containers.length - runningCount);
  let exposedPorts = $derived(
    containers
      .filter((c) => c.state === "running" && c.ports)
      .flatMap((c) => c.ports.split(", ").filter(Boolean)).length
  );

  const load = async () => {
    loading = true;
    try {
      const [healthRes, statusRes, presetsRes] = await Promise.allSettled([
        api.health(),
        api.getStatus(),
        api.listPresets(),
      ]);

      if (healthRes.status === "fulfilled") {
        dockerOnline = healthRes.value.docker;
        dbOnline = true;
      }

      if (statusRes.status === "fulfilled") {
        containers = statusRes.value.containers ?? [];
      }

      if (presetsRes.status === "fulfilled") {
        presets = presetsRes.value.presets ?? [];
      }

      try {
        await fetch("https://hub.docker.com/v2/", {
          method: "HEAD",
          mode: "no-cors",
        });
        hubOnline = true;
      } catch {
        hubOnline = false;
      }
    } finally {
      loading = false;
    }
  };

  const stopAll = async () => {
    const running = containers.filter((c) => c.state === "running");
    if (running.length === 0) return;
    stoppingAll = true;
    status.clear();
    try {
      await Promise.all(running.map((c) => api.stopContainer(c.id)));
      status.success("All containers stopped.");
      await load();
    } catch (e) {
      status.error(e);
    } finally {
      stoppingAll = false;
    }
  };

  const stopContainer = async (id: string) => {
    try {
      await api.stopContainer(id);
      await load();
    } catch (e) {
      status.error(e);
    }
  };

  const startContainer = async (id: string) => {
    try {
      await api.startContainer(id);
      await load();
    } catch (e) {
      status.error(e);
    }
  };

  const removeContainer = async (id: string) => {
    try {
      await api.removeContainer(id);
      await load();
    } catch (e) {
      status.error(e);
    }
  };

  const toggleLogs = async (container: ContainerStatus) => {
    if (expandedLogsId === container.id) {
      expandedLogsId = null;
      return;
    }
    expandedLogsId = container.id;
    if (!(container.id in logsContent)) {
      try {
        const res = await api.getLogs(container.id);
        logsContent = { ...logsContent, [container.id]: res.logs };
      } catch {
        logsContent = { ...logsContent, [container.id]: "Failed to load logs." };
      }
    }
  };

  const bootPreset = async (preset: Preset) => {
    try {
      await api.startLab({ name: preset.name, services: preset.services });
      status.success(`"${preset.name}" started!`);
      await load();
    } catch (e) {
      status.error(e);
    }
  };

  onMount(load);
</script>

<div class="flex flex-col gap-5">
  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-base font-medium text-gray-900">Dashboard</h1>
      <p class="mt-0.5 text-xs text-gray-400">Overview of your lab environment</p>
    </div>
    <Button onclick={load} disabled={loading}>
      {#snippet icon()}
        <RefreshCw size={13} class={loading ? "animate-spin" : ""} />
      {/snippet}
      {loading ? "Refreshing..." : "Refresh"}
    </Button>
  </div>

  {#if status.message}
    <Notification variant={status.variant}>
      {status.message}
    </Notification>
  {/if}

  <div class="grid grid-cols-4 gap-2.5">
    <Card styleConfig={statTileStyle}>
      <div class="mb-1 text-[11px] text-gray-400">Running</div>
      <div class="text-xl font-medium {runningCount > 0 ? 'text-emerald-700' : 'text-gray-400'}">
        {runningCount}
      </div>
      <div class="mt-0.5 text-[11px] text-gray-400">
        of {containers.length} containers
      </div>
    </Card>

    <Card styleConfig={statTileStyle}>
      <div class="mb-1 text-[11px] text-gray-400">Stopped</div>
      <div class="text-xl font-medium {stoppedCount > 0 ? 'text-red-500' : 'text-gray-400'}">
        {stoppedCount}
      </div>
      <div class="mt-0.5 text-[11px] text-gray-400">containers</div>
    </Card>

    <Card styleConfig={statTileStyle}>
      <div class="mb-1 text-[11px] text-gray-400">Exposed ports</div>
      <div class="text-xl font-medium text-gray-800">{exposedPorts}</div>
      <div class="mt-0.5 text-[11px] text-gray-400">active mappings</div>
    </Card>

    <Card styleConfig={statTileStyle}>
      <div class="mb-1 text-[11px] text-gray-400">Presets</div>
      <div class="text-xl font-medium text-gray-800">{presets.length}</div>
      <div class="mt-0.5 text-[11px] text-gray-400">saved</div>
    </Card>
  </div>

  <Card>
    <div class="mb-4 flex items-center gap-1.5 text-xs font-medium text-gray-500">
      <Container size={13} />
      Containers
      <div class="ml-auto flex items-center gap-2">
        {#if runningCount > 0}
          <Button
            onclick={stopAll}
            disabled={stoppingAll}
            styleConfig={{
              backgroundColor: "#fef2f2",
              color: "#dc2626",
              borderColor: "#fecaca",
              hoverBackgroundColor: "#fee2e2",
              padding: "0.25rem 0.625rem",
              borderRadius: "0.375rem",
              fontSize: "11px",
              gap: "0.25rem",
            }}
          >
            {#snippet icon()}<Square size={11} />{/snippet}
            {stoppingAll ? "Stopping..." : "Stop all"}
          </Button>
        {/if}
      </div>
    </div>

    {#if loading}
      <div class="flex items-center gap-2 py-4 text-xs text-gray-400">
        <div
          class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
        ></div>
        Loading containers...
      </div>
    {:else if containers.length === 0}
      <p class="py-2 text-xs text-gray-400">
        No VulnKit containers found. Start a lab to see containers here.
      </p>
    {:else}
      <div class="flex flex-col gap-1.5">
        {#each containers as container}
          <div class="overflow-hidden rounded-lg border border-gray-100 bg-gray-50">
            <div class="flex items-center gap-2.5 px-3 py-2">
              <span class="w-28 shrink-0 truncate text-sm font-medium text-gray-900"
                >{container.name}</span
              >

              <span class="flex-1 truncate font-mono text-[11px] text-gray-400"
                >{container.image}</span
              >

              {#if container.ports}
                <span class="shrink-0 font-mono text-[11px] text-gray-400">{container.ports}</span>
              {/if}

              <span class="w-20 shrink-0 text-right text-[11px] text-gray-400"
                >{container.status}</span
              >
              <Badge
                styleConfig={{
                  backgroundColor: container.state === "running" ? "#ecfdf5" : "#f3f4f6",
                  color: container.state === "running" ? "oklch(50.8% 0.118 165.612)" : "#6a7282",
                }}
              >
                {container.state}
              </Badge>
              <div class="flex shrink-0 gap-1">
                <Button
                  onclick={() => toggleLogs(container)}
                  title="Logs"
                  styleConfig={{
                    color: "#9ca3af",
                    hoverBackgroundColor: "#f3f4f6",
                    width: "1.5rem",
                    height: "1.5rem",
                    padding: "0",
                    borderRadius: "0.25rem",
                  }}
                >
                  {#snippet icon()}<Terminal size={11} />{/snippet}
                </Button>

                {#if container.state === "running"}
                  <Button
                    onclick={() => stopContainer(container.id)}
                    title="Stop"
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
                    {#snippet icon()}<Square size={10} />{/snippet}
                  </Button>
                {:else}
                  <Button
                    onclick={() => startContainer(container.id)}
                    title="Start"
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
                    onclick={() => removeContainer(container.id)}
                    title="Remove"
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
                {/if}
              </div>
            </div>
            {#if expandedLogsId === container.id}
              <div class="border-t border-gray-100 bg-white px-3 py-2">
                {#if logsContent[container.id] !== undefined}
                  <pre
                    class="max-h-64 overflow-auto rounded-md bg-gray-900 p-2 font-mono text-[10px] leading-relaxed whitespace-pre text-gray-100">{logsContent[
                      container.id
                    ] || "(no logs)"}</pre>
                {:else}
                  <div class="flex items-center gap-2 text-xs text-gray-400">
                    <div
                      class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
                    ></div>
                    Loading logs...
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </Card>

  <div class="grid grid-cols-2 gap-3">
    <Card>
      <div class="mb-4 flex items-center gap-1.5 text-xs font-medium text-gray-500">
        <Bookmark size={13} />
        Quick presets
      </div>
      {#if presets.length === 0}
        <p class="text-xs text-gray-400">No saved presets yet.</p>
      {:else}
        <div class="flex flex-col gap-1.5">
          {#each presets as preset}
            <div
              class="flex items-center gap-2 rounded-lg border border-gray-100 bg-gray-50 px-2.5 py-2"
            >
              <span class="flex-1 truncate text-xs font-medium text-gray-800">{preset.name}</span>
              <div class="flex shrink-0 items-center gap-1">
                {#each preset.tags ?? [] as tag}
                  <Badge
                    styleConfig={{
                      backgroundColor: "#dbeafe",
                      color: "#1447e6",
                    }}
                  >
                    {tag}
                  </Badge>
                {/each}
              </div>
              <Button
                onclick={() => bootPreset(preset)}
                title="Boot preset"
                class="shrink-0"
                styleConfig={{
                  color: "#9ca3af",
                  hoverBackgroundColor: "#ecfdf5",
                  hoverBorderColor: "#a7f3d0",
                  hoverColor: "#047857",
                  width: "1.25rem",
                  height: "1.25rem",
                  padding: "0",
                  borderRadius: "0.25rem",
                }}
              >
                {#snippet icon()}<Play size={9} />{/snippet}
              </Button>
            </div>
          {/each}
        </div>
      {/if}
    </Card>

    <Card>
      <div class="mb-4 flex items-center gap-1.5 text-xs font-medium text-gray-500">
        <Shield size={13} />
        System
      </div>
      <div class="flex flex-col gap-3">
        {#each [{ label: "Docker Engine", online: dockerOnline }, { label: "Database", online: dbOnline }, { label: "Docker Hub", online: hubOnline }] as item}
          <div class="flex items-center justify-between">
            <span class="text-xs text-gray-600">{item.label}</span>
            <div class="flex items-center gap-1.5">
              <div
                class="h-1.5 w-1.5 rounded-full {item.online ? 'bg-emerald-500' : 'bg-red-400'}"
              ></div>
              <span class="text-[11px] {item.online ? 'text-emerald-700' : 'text-red-500'}">
                {item.online ? "online" : "offline"}
              </span>
            </div>
          </div>
          {#if item.label !== "Docker Hub"}
            <div class="h-px bg-gray-100"></div>
          {/if}
        {/each}
      </div>
    </Card>
  </div>
</div>
