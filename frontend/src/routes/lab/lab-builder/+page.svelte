<script lang="ts">
  import { onMount } from "svelte";
  import { api, type ServiceConfig, type DynamicBuild } from "$lib/api";
  import { KNOWN_PORTS } from "$lib/utils/ports";
  import { versionToNumber } from "$lib/utils/version";
  import { createStatusMessage } from "$lib/utils/statusMessage.svelte";
  import type { StyleConfig } from "$lib/types/style";
  import Card from "$lib/components/Card/Card.svelte";
  import Input from "$lib/components/Input/Input.svelte";
  import Badge from "$lib/components/Badge/Badge.svelte";
  import { DEFAULT_ENV } from "$lib/utils/defaultEnvs";
  import Button from "$lib/components/Button/Button.svelte";
  import Notification from "$lib/components/Notification/Notification.svelte";
  import ProgressBar from "$lib/components/ProgressBar/ProgressBar.svelte";
  import { page } from "$app/state";
  import Bookmark from "@lucide/svelte/icons/bookmark";
  import Play from "@lucide/svelte/icons/play";
  import Square from "@lucide/svelte/icons/square";
  import Search from "@lucide/svelte/icons/search";
  import Wrench from "@lucide/svelte/icons/wrench";
  import TriangleAlert from "@lucide/svelte/icons/triangle-alert";
  import RotateCcw from "@lucide/svelte/icons/rotate-ccw";
  import Check from "@lucide/svelte/icons/check";
  import CircleAlert from "@lucide/svelte/icons/circle-alert";
  import Plus from "@lucide/svelte/icons/plus";
  import X from "@lucide/svelte/icons/x";
  import Code from "@lucide/svelte/icons/code";
  import ChevronRight from "@lucide/svelte/icons/chevron-right";
  import Database from "@lucide/svelte/icons/database";

  type AddedService = {
    id: string;
    image: string;
    tag: string;
    ports: string[];
    isDynamic?: boolean;
  };

  type HubImage = {
    repo_name: string;
    short_description: string;
    pull_count: number;
    is_official: boolean;
  };

  type HubTag = { name: string };

  let searchQuery = $state("");
  let hubImages = $state<HubImage[]>([]);
  let hubTags = $state<HubTag[]>([]);
  let selectedImage = $state<HubImage | null>(null);
  let portInput = $state("");
  let searching = $state(false);
  let loadingTags = $state(false);
  let searchError = $state("");

  let versionInput = $state("");
  let selectedTag = $state("");
  let showDropdown = $state(false);
  let versionNotFound = $state(false);
  let dynamicSupported = $state(false);
  let dynamicImageName = $state("");
  let building = $state(false);
  let buildLogs = $state<string[]>([]);
  let buildDone = $state(false);

  let services = $state<AddedService[]>([]);
  let labName = $state("my-lab");
  let composeYaml = $state("");
  let loading = $state(false);
  let labRunning = $state(false);
  let stoppingLab = $state(false);
  const status = createStatusMessage();
  let presets = $state<any[]>([]);

  let localImages = $state<DynamicBuild[]>([]);

  let portWarning = $state("");
  let portCheckTimer: ReturnType<typeof setTimeout>;

  let debounceTimer: ReturnType<typeof setTimeout>;
  let versionCheckTimer: ReturnType<typeof setTimeout>;

  const compactInputStyle = (warning: boolean): StyleConfig => ({
    backgroundColor: warning ? "#fffbeb" : "#ffffff",
    focusBackgroundColor: warning ? "#fffbeb" : "#ffffff",
    borderColor: warning ? "#fcd34d" : "#e5e7eb",
    focusBorderColor: warning ? "#fbbf24" : "#10b981",
    borderRadius: "0.375rem",
    padding: "0.375rem 0.5rem",
    fontSize: "0.75rem",
    height: "2rem",
  });

  let versionInputStyle = $derived(compactInputStyle(versionNotFound));
  let portInputStyle = $derived(compactInputStyle(!!portWarning));

  let tagSuggestions = $derived(hubTags.map((t) => t.name));

  let filteredTags = $derived(
    versionInput
      ? tagSuggestions.filter((t) => t.startsWith(versionInput)).slice(0, 8)
      : tagSuggestions.slice(0, 8)
  );

  let selectedServices = $derived(
    services.map((s) => {
      const baseName = s.image.split("/").pop() ?? s.image;
      return {
        name: baseName,
        image: s.image,
        version: s.tag,
        ports: s.ports,
        env_vars: s.isDynamic ? {} : (DEFAULT_ENV[baseName] ?? {}),
      } as ServiceConfig;
    })
  );

  const onSearchInput = () => {
    clearTimeout(debounceTimer);
    hubImages = [];
    selectedImage = null;
    hubTags = [];
    searchError = "";
    versionInput = "";
    selectedTag = "";
    if (searchQuery.trim().length < 2) return;
    debounceTimer = setTimeout(searchDockerHub, 300);
  };

  const searchDockerHub = async () => {
    searching = true;
    searchError = "";
    try {
      const data = await api.hubSearch(searchQuery);
      hubImages = data.results ?? [];
    } catch {
      searchError = "Docker Hub unavailable";
    } finally {
      searching = false;
    }
  };

  const findFreeHostPort = (defaultPortMapping: string): string => {
    const [defaultHostPort, containerPort] = defaultPortMapping.split(":");
    const usedHostPorts = new Set(services.flatMap((s) => s.ports.map((p) => p.split(":")[0])));

    let hostPort = parseInt(defaultHostPort, 10);
    if (Number.isNaN(hostPort)) return defaultPortMapping;

    while (usedHostPorts.has(String(hostPort))) {
      hostPort++;
    }

    return `${hostPort}:${containerPort ?? defaultHostPort}`;
  };

  const selectImage = async (img: HubImage) => {
    selectedImage = img;
    hubTags = [];
    selectedTag = "";
    versionInput = "";
    versionNotFound = false;
    buildLogs = [];
    buildDone = false;
    loadingTags = true;

    const baseName = img.repo_name.split("/").pop() ?? img.repo_name;
    portInput = KNOWN_PORTS[baseName] ? findFreeHostPort(KNOWN_PORTS[baseName]) : "";

    try {
      const data = await api.hubTags(img.repo_name);
      hubTags = data.results ?? [];
      if (hubTags.length > 0) {
        versionInput = hubTags[0].name;
        selectedTag = hubTags[0].name;
      }
    } catch {
      searchError = "Failed to load tags";
    } finally {
      loadingTags = false;
    }
  };

  const onVersionInput = () => {
    clearTimeout(versionCheckTimer);
    showDropdown = true;
    versionNotFound = false;
    dynamicSupported = false;
    buildLogs = [];
    buildDone = false;

    if (!versionInput || !selectedImage) return;

    const exact = tagSuggestions.find((t) => t === versionInput);
    if (exact) {
      selectedTag = versionInput;
      return;
    }

    selectedTag = "";

    if (versionInput.length >= 2) {
      const filtered = tagSuggestions.filter((t) => t.startsWith(versionInput));
      if (filtered.length === 0) {
        versionNotFound = true;
        versionCheckTimer = setTimeout(checkDynamicBuild, 400);
      }
    }
  };

  const checkDynamicBuild = async () => {
    if (!selectedImage || !versionInput) return;
    try {
      const check = await api.dynamicBuildCheck(selectedImage.repo_name, versionInput);
      dynamicSupported = check.supported;
      dynamicImageName = check.image_name ?? "";
    } catch {}
  };

  const selectTag = (tag: string) => {
    versionInput = tag;
    selectedTag = tag;
    versionNotFound = false;
    showDropdown = false;
  };

  const isVersionTag = (tag: string): boolean => /^\d+(\.\d+){0,2}$/.test(tag);

  const useSuggestedVersion = () => {
    const versionTags = tagSuggestions.filter(isVersionTag);
    if (versionTags.length === 0) return;

    const target = versionToNumber(versionInput);
    const closest = versionTags.reduce((best, tag) =>
      Math.abs(versionToNumber(tag) - target) < Math.abs(versionToNumber(best) - target)
        ? tag
        : best
    );
    selectTag(closest);
  };

  let showBuildPanel = $state(false);

  const startDynamicBuild = async () => {
    if (!selectedImage) return;
    building = true;
    buildLogs = [];
    buildDone = false;
    showBuildPanel = true;
    showDropdown = false;
    versionNotFound = false;

    try {
      await api.dynamicBuild(
        selectedImage.repo_name,
        versionInput,
        (line) => {
          buildLogs = [...buildLogs, line];
        },
        (imageName) => {
          buildDone = true;
          building = false;
          dynamicImageName = imageName;
          selectedTag = versionInput;
          versionNotFound = false;
          loadLocalImages();
        },
        (err) => {
          buildLogs = [...buildLogs, `✗ Error: ${err}`];
          building = false;
        }
      );
    } catch (e: any) {
      buildLogs = [...buildLogs, `✗ ${e.message}`];
      building = false;
    }
  };

  const onPortInput = () => {
    clearTimeout(portCheckTimer);
    portWarning = "";
    if (!portInput || !portInput.includes(":")) return;

    portCheckTimer = setTimeout(async () => {
      const hostPort = portInput.split(":")[0];
      try {
        const res = await api.checkPort(hostPort);
        if (res.in_use) {
          portWarning = `Port ${hostPort} is already in use by ${res.container || "another container"}`;
        }
      } catch {}
    }, 400);
  };

  const addService = () => {
    if (!selectedImage || !selectedTag) return;
    const ports = portInput
      .split(",")
      .map((p) => p.trim())
      .filter(Boolean);
    const imageToUse = buildDone && dynamicImageName ? dynamicImageName : selectedImage.repo_name;

    services = [
      ...services,
      {
        id: `${imageToUse}-${Date.now()}`,
        image: imageToUse,
        tag: buildDone ? "" : selectedTag,
        ports,
        isDynamic: buildDone,
      },
    ];

    searchQuery = "";
    hubImages = [];
    selectedImage = null;
    hubTags = [];
    selectedTag = "";
    versionInput = "";
    portInput = "";
    portWarning = "";
    versionNotFound = false;
    building = false;
    buildLogs = [];
    buildDone = false;
    showBuildPanel = false;
    refreshCompose();
  };

  const removeService = (id: string) => {
    services = services.filter((s) => s.id !== id);
    refreshCompose();
  };

  $effect(() => {
    if (selectedServices.length > 0) refreshCompose();
    else composeYaml = "";
  });

  const refreshCompose = async () => {
    if (selectedServices.length === 0) {
      composeYaml = "";
      return;
    }
    try {
      const res = await api.generateCompose({
        name: labName,
        services: selectedServices,
      });
      composeYaml = res.yaml;
    } catch {}
  };

  const startLab = async () => {
    if (selectedServices.length === 0) return;
    loading = true;
    status.clear();
    try {
      await api.startLab({ name: labName, services: selectedServices });
      labRunning = true;
      status.success("Lab started successfully!");
    } catch (e: any) {
      const firstLine =
        e.message?.split("\n").find((l: string) => l.includes("Error")) ?? e.message;
      status.error(firstLine);
    } finally {
      loading = false;
    }
  };

  const stopLab = async () => {
    stoppingLab = true;
    status.clear();
    try {
      await api.stopLab(labName);
      labRunning = false;
      status.success("Lab stopped.");
    } catch (e) {
      status.error(e);
    } finally {
      stoppingLab = false;
    }
  };

  const saveAsPreset = async () => {
    const name = prompt("Preset name:");
    if (!name) return;
    const tag = prompt("Tag (e.g. SQLi, XSS, RCE):") ?? "";
    await api.savePreset({
      name,
      tags: tag ? [tag] : [],
      services: selectedServices,
    });
    await loadPresets();
  };

  const loadPresets = async () => {
    try {
      const res = await api.listPresets();
      presets = res.presets ?? [];
    } catch {}
  };

  const loadLocalImages = async () => {
    try {
      const res = await api.listDynamicBuilds();
      localImages = res.builds ?? [];
    } catch {}
  };

  const loadPresetIntoBuilder = (preset: any) => {
    services = preset.services.map((svc: any) => ({
      id: `${svc.image}-${Date.now()}-${Math.random()}`,
      image: svc.image,
      tag: svc.version,
      ports: svc.ports,
      isDynamic: false,
    }));
    labName = preset.name;
    status.success(`"${preset.name}" loaded into builder`);
    setTimeout(() => status.clear(), 3000);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  const loadLocalImageIntoBuilder = (build: DynamicBuild) => {
    const baseName = build.software;
    services = [
      {
        id: `${build.image_name}-${Date.now()}`,
        image: build.image_name,
        tag: "",
        ports: KNOWN_PORTS[baseName] ? [KNOWN_PORTS[baseName]] : [],
        isDynamic: true,
      },
    ];
    labName = build.image_name;
    status.success(`"${build.image_name}" loaded into builder`);
    setTimeout(() => status.clear(), 3000);
    window.scrollTo({ top: 0, behavior: "smooth" });
  };

  onMount(async () => {
    const params = page.url.searchParams;
    const imageParam = params.get("image");
    const versionParam = params.get("version");

    if (imageParam && versionParam) {
      try {
        const check = await api.dynamicBuildCheck(imageParam, versionParam);
        const tagsData = await api.hubTags(imageParam);
        const tags =
          tagsData.results
            ?.map((t: { name: string }) => t.name)
            .filter((t: string) => /^\d+\.\d+/.test(t)) ?? [];
        const isOnHub = tags.includes(versionParam);

        if (isOnHub) {
          services = [
            {
              id: `${imageParam}-${Date.now()}`,
              image: imageParam,
              tag: versionParam,
              ports: KNOWN_PORTS[imageParam] ? [KNOWN_PORTS[imageParam]] : [],
            },
          ];
          await refreshCompose();
          status.success(`Pre-loaded from CVE: ${imageParam}:${versionParam}`);
          setTimeout(() => status.clear(), 4000);
        } else {
          searchQuery = imageParam;
          await searchDockerHub();
          const match = hubImages.find(
            (img) => img.repo_name === imageParam || img.repo_name.endsWith(`/${imageParam}`)
          );
          if (match) {
            await selectImage(match);
            versionInput = versionParam;
            selectedTag = "";
            versionNotFound = true;
            dynamicSupported = check.supported;
            dynamicImageName = check.image_name ?? "";
          }
          status.success(`${imageParam}:${versionParam} is not available on Docker Hub.`);
          setTimeout(() => status.clear(), 5000);
        }
      } catch {
        services = [
          {
            id: `${imageParam}-${Date.now()}`,
            image: imageParam,
            tag: versionParam,
            ports: KNOWN_PORTS[imageParam] ? [KNOWN_PORTS[imageParam]] : [],
          },
        ];
        await refreshCompose();
      }
    }

    await Promise.all([loadPresets(), loadLocalImages()]);
  });
</script>

<div class="flex flex-col gap-5">
  <div class="flex items-start justify-between">
    <div>
      <h1 class="text-base font-medium text-gray-900">Lab builder</h1>
      <p class="mt-0.5 text-xs text-gray-400">
        Search Docker images and configure your lab environment
      </p>
    </div>
    <div class="flex gap-2">
      <Button onclick={saveAsPreset} disabled={services.length === 0}>
        {#snippet icon()}
          <Bookmark size={13} />
        {/snippet}
        Save as preset
      </Button>
      {#if labRunning}
        <Button
          disabled={stoppingLab}
          onclick={stopLab}
          styleConfig={{
            backgroundColor: "#fef2f2",
            color: "#dc2626",
            borderColor: "#fecaca",
            hoverBackgroundColor: "#fee2e2",
          }}
        >
          {#snippet icon()}
            <Square size={13} />
          {/snippet}
          {stoppingLab ? "Stopping..." : "Stop lab"}
        </Button>
      {:else}
        <Button
          disabled={loading || services.length === 0}
          onclick={startLab}
          styleConfig={{
            backgroundColor: "#009966",
            color: "#ffffff",
            hoverBackgroundColor: "#007a55",
            borderColor: "#009966",
          }}
        >
          {#snippet icon()}
            <Play size={13} />
          {/snippet}
          {loading ? "Starting..." : "Start lab"}
        </Button>
      {/if}
    </div>
  </div>

  {#if status.message}
    <Notification variant={status.variant}>
      {status.message}
    </Notification>
  {/if}

  <Card>
    <div class="mb-3 flex items-center gap-1.5 text-xs font-medium text-gray-500">
      <Search size={13} />
      Docker image search
    </div>

    <Input
      bind:value={searchQuery}
      placeholder="e.g. mysql, nginx, postgres..."
      isLoading={searching}
      oninput={() => onSearchInput()}
    />

    {#if searchError}
      <p class="mt-2 text-xs text-red-600">{searchError}</p>
    {/if}

    {#if hubImages.length > 0}
      <div class="mt-3 mb-2 grid grid-cols-2 gap-1.5">
        {#each hubImages as img}
          <button
            onclick={() => selectImage(img)}
            class="rounded-lg border px-3 py-2.5 text-left transition-colors
              {selectedImage?.repo_name === img.repo_name
              ? 'border-emerald-500 bg-emerald-50'
              : 'border-gray-100 bg-gray-50 hover:border-gray-200 hover:bg-gray-100'}"
          >
            <div class="mb-1 flex items-center gap-1.5">
              <span class="text-sm font-medium text-gray-900">{img.repo_name}</span>
              {#if img.is_official}
                <Badge
                  class="leading-none font-medium"
                  styleConfig={{
                    backgroundColor: "#dbeafe",
                    color: "#1d4ed8",
                    borderWidth: "0",
                    borderRadius: "0.25rem",
                    padding: "0.125rem 0.375rem",
                    fontSize: "9px",
                  }}>Official</Badge
                >
              {/if}
            </div>
            {#if img.short_description}
              <p class="mb-1.5 line-clamp-2 text-[11px] leading-snug text-gray-400">
                {img.short_description}
              </p>
            {/if}
            <span class="text-[10px] text-gray-400">
              ↓ {img.pull_count >= 1_000_000
                ? `${(img.pull_count / 1_000_000).toFixed(0)}M`
                : img.pull_count >= 1_000
                  ? `${(img.pull_count / 1_000).toFixed(0)}K`
                  : img.pull_count} pulls
            </span>
          </button>
        {/each}
      </div>
    {/if}

    {#if selectedImage}
      <div class="flex flex-col gap-3 border-t border-gray-100 pt-4">
        <div class="flex gap-3">
          <div class="flex flex-1 flex-col gap-1">
            <label class="text-[11px] font-medium text-gray-600">Version tag</label>
            {#if loadingTags}
              <div
                class="flex h-8 items-center gap-2 rounded-md border border-gray-200 bg-gray-50 px-2"
              >
                <div
                  class="h-3 w-3 animate-spin rounded-full border-2 border-gray-200 border-t-emerald-500"
                ></div>
                <span class="text-xs text-gray-400">Loading tags...</span>
              </div>
            {:else}
              <div class="relative">
                <Input
                  bind:value={versionInput}
                  placeholder="type or select a version..."
                  oninput={onVersionInput}
                  onfocus={() => (showDropdown = true)}
                  onblur={() =>
                    setTimeout(() => {
                      showDropdown = false;
                    }, 150)}
                  class="w-full font-mono"
                  styleConfig={versionInputStyle}
                />
                {#if showDropdown && filteredTags.length > 0}
                  <div
                    class="absolute top-full right-0 left-0 z-20 mt-1 overflow-hidden rounded-lg border border-gray-200 bg-white shadow-sm"
                  >
                    {#each filteredTags as tag}
                      <button
                        onmousedown={() => selectTag(tag)}
                        class="flex w-full items-center justify-between px-3 py-1.5 text-left font-mono text-xs hover:bg-gray-50
                          {selectedTag === tag
                          ? 'bg-emerald-50 text-emerald-700'
                          : 'text-gray-700'}"
                      >
                        {tag}
                        {#if tag === tagSuggestions[0]}
                          <Badge
                            styleConfig={{
                              backgroundColor: "#ecfdf5",
                              color: "#059669",
                              borderColor: "#a7f3d0",
                              borderWidth: "1px",
                              borderRadius: "0.25rem",
                              padding: "0.125rem 0.375rem",
                              fontSize: "9px",
                            }}>latest</Badge
                          >
                        {/if}
                      </button>
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}
          </div>

          <div class="flex flex-1 flex-col gap-1">
            <label class="text-[11px] font-medium text-gray-600">
              Port mapping
              <span class="font-normal text-gray-400">(e.g. 3306:3306)</span>
            </label>
            <Input
              bind:value={portInput}
              placeholder="3306:3306"
              oninput={onPortInput}
              class="w-full"
              styleConfig={portInputStyle}
            />
            {#if portWarning}
              <p class="mt-1 flex items-center gap-1 text-[10px] text-amber-700">
                <TriangleAlert size={10} />
                {portWarning}
              </p>
            {/if}
          </div>
        </div>

        {#if versionNotFound && !building && !buildDone}
          <Notification
            variant="warning"
            class="flex flex-col gap-2"
            styleConfig={{ padding: "0.75rem" }}
          >
            <div class="flex items-center gap-2 text-xs font-medium text-amber-800">
              <TriangleAlert size={13} />
              {selectedImage.repo_name}:{versionInput} not found on Docker Hub
            </div>
            <p class="text-[11px] leading-relaxed text-amber-700">
              {#if dynamicSupported}
                VulnKit can build a custom image dynamically using the system package manager. This
                may take 2-5 minutes on first run.
              {:else}
                Dynamic build is not supported for this software/version combination. Try a
                different version.
              {/if}
            </p>
            <div class="flex flex-wrap gap-2">
              {#if dynamicSupported}
                <Button
                  onclick={startDynamicBuild}
                  styleConfig={{
                    backgroundColor: "#fef3c7",
                    color: "#92400e",
                    hoverBackgroundColor: "#fde68a",
                    borderColor: "#fcd34d",
                  }}
                >
                  {#snippet icon()}
                    <Wrench size={11} />
                  {/snippet}
                  Build {selectedImage.repo_name}:{versionInput} dynamically
                </Button>
              {/if}
              <Button onclick={useSuggestedVersion}>
                {#snippet icon()}
                  <RotateCcw size={11} />
                {/snippet}
                Use closest available version
              </Button>
            </div>
          </Notification>
        {/if}

        {#if showBuildPanel}
          <div class="flex flex-col gap-2 rounded-lg border border-gray-200 bg-gray-50 p-3">
            <div class="flex items-center justify-between">
              <span class="flex items-center gap-1.5 text-xs font-medium text-gray-700">
                <Wrench size={12} />
                Building {selectedImage.repo_name}:{versionInput}
              </span>
              {#if buildDone}
                <span class="flex items-center gap-1 text-[11px] text-emerald-700">
                  <Check size={11} strokeWidth={2.5} />
                  Done
                </span>
              {:else if building}
                <span class="flex items-center gap-1.5 text-[11px] text-amber-700">
                  <div
                    class="h-2.5 w-2.5 animate-spin rounded-full border-2 border-amber-300 border-t-amber-600"
                  ></div>
                  Building...
                </span>
              {:else}
                <span class="flex items-center gap-1.5 text-[11px] text-red-600">
                  <CircleAlert size={11} />
                  Failed
                </span>
              {/if}
            </div>

            <ProgressBar
              value={buildDone || !building ? 100 : 66}
              variant={buildDone ? "success" : building ? "warning" : "error"}
            />

            <div
              class="max-h-28 overflow-y-auto rounded-md border border-gray-100 bg-white p-2 font-mono text-[10px] leading-relaxed text-gray-500"
            >
              {#each buildLogs as line}
                <div
                  class="
          {line.startsWith('✓')
                    ? 'text-emerald-600'
                    : line.startsWith('✗')
                      ? 'text-red-500'
                      : line.startsWith('→')
                        ? 'text-gray-800'
                        : 'text-gray-400'}"
                >
                  {line}
                </div>
              {/each}
            </div>

            {#if !building && !buildDone}
              <Button
                onclick={startDynamicBuild}
                class="self-start"
                styleConfig={{
                  backgroundColor: "#fffbeb",
                  color: "#b45309",
                  hoverBackgroundColor: "#fef3c7",
                  borderColor: "#fcd34d",
                }}
              >
                {#snippet icon()}
                  <RotateCcw size={11} />
                {/snippet}
                Retry
              </Button>
            {/if}
          </div>
        {/if}

        <Button
          onclick={addService}
          disabled={!selectedTag || loadingTags || building}
          styleConfig={{
            backgroundColor: "#009966",
            color: "#ffffff",
            hoverBackgroundColor: "#007a55",
            borderColor: "#009966",
          }}
        >
          {#snippet icon()}
            <Plus size={12} strokeWidth={2.5} />
          {/snippet}
          Add service — {buildDone && dynamicImageName
            ? dynamicImageName
            : `${selectedImage.repo_name}${selectedTag ? `:${selectedTag}` : ""}`}
          {#if buildDone}
            <span class="text-[10px] opacity-80">(dynamic)</span>
          {/if}
        </Button>
      </div>
    {/if}
  </Card>

  {#if services.length > 0}
    <Card>
      <div class="mb-2 text-[11px] font-medium tracking-wide text-gray-400 uppercase">
        Added services ({services.length})
      </div>
      <div class="flex flex-col gap-1.5">
        {#each services as svc}
          <div
            class="flex items-center gap-2.5 rounded-lg border border-gray-100 bg-white px-3 py-2 shadow-sm"
          >
            <div
              class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md border border-gray-200 bg-gray-100 text-[10px] font-semibold text-gray-500"
            >
              {svc.image.split("/").pop()!.slice(0, 2).toUpperCase()}
            </div>
            <div class="flex min-w-0 flex-1 items-center gap-1.5">
              <span class="truncate text-sm font-medium text-gray-900">{svc.image}</span>
              <span class="shrink-0 font-mono text-xs text-emerald-700">:{svc.tag}</span>
              {#if svc.isDynamic}
                <Badge
                  class="shrink-0"
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
              {/if}
            </div>
            {#if svc.ports.length > 0}
              <span class="shrink-0 font-mono text-[11px] text-gray-400"
                >{svc.ports.join(", ")}</span
              >
            {:else}
              <span class="shrink-0 text-[11px] text-gray-300">no ports</span>
            {/if}
            <Button
              onclick={() => removeService(svc.id)}
              title="Remove"
              class="shrink-0"
              styleConfig={{
                color: "#9ca3af",
                hoverBackgroundColor: "#fef2f2",
                hoverBorderColor: "#fecaca",
                hoverColor: "#dc2626",
                width: "1.25rem",
                height: "1.25rem",
                padding: "0",
                borderRadius: "0.25rem",
              }}
            >
              {#snippet icon()}<X size={10} strokeWidth={2.5} />{/snippet}
            </Button>
          </div>
        {/each}
      </div>
    </Card>
  {/if}

  <div class="grid grid-cols-2 gap-3.5">
    <Card>
      <div class="mb-3 flex items-center gap-1.5 text-xs font-medium text-gray-500">
        <Code size={13} />
        docker-compose preview
      </div>
      {#if composeYaml}
        <pre
          class="overflow-x-auto rounded-lg bg-gray-50 p-2.5 font-mono text-[10px] leading-relaxed whitespace-pre text-gray-500">{composeYaml}</pre>
      {:else}
        <p class="text-xs text-gray-400">Add at least one service to preview</p>
      {/if}
    </Card>

    <div class="flex flex-col gap-3.5">
      <Card>
        <div class="mb-3 flex items-center justify-between">
          <div class="flex items-center gap-1.5 text-xs font-medium text-gray-500">
            <Bookmark size={13} />
            Saved presets
          </div>
          <a href="/lab/presets" class="text-[10px] text-gray-400 hover:text-emerald-600">
            View all →
          </a>
        </div>
        {#if presets.length === 0}
          <p class="text-xs text-gray-400">No saved presets yet</p>
        {:else}
          <div class="flex flex-col gap-1">
            {#each presets.slice(0, 5) as preset}
              <button
                onclick={() => loadPresetIntoBuilder(preset)}
                class="group flex w-full items-center gap-2 rounded-lg border border-gray-100 bg-gray-50 px-2.5 py-2 text-left transition-colors hover:border-emerald-200 hover:bg-emerald-50"
              >
                <span
                  class="flex-1 truncate text-xs font-medium text-gray-800 group-hover:text-emerald-800"
                  >{preset.name}</span
                >
                <div class="flex shrink-0 items-center gap-1">
                  {#each preset.tags ?? [] as tag}
                    <Badge
                      styleConfig={{
                        backgroundColor: "#dbeafe",
                        color: "#1447e6",
                      }}>{tag}</Badge
                    >
                  {/each}
                </div>
                <ChevronRight
                  size={10}
                  class="shrink-0 text-gray-300 group-hover:text-emerald-500"
                />
              </button>
            {/each}
          </div>
        {/if}
      </Card>

      <Card>
        <div class="mb-3 flex items-center justify-between">
          <div class="flex items-center gap-1.5 text-xs font-medium text-gray-500">
            <Database size={13} />
            Local images
          </div>
          <a href="/lab/local-images" class="text-[10px] text-gray-400 hover:text-emerald-600">
            View all →
          </a>
        </div>
        {#if localImages.length === 0}
          <p class="text-xs text-gray-400">No dynamic builds yet</p>
        {:else}
          <div class="flex flex-col gap-1">
            {#each localImages.slice(0, 5) as build}
              <button
                onclick={() => loadLocalImageIntoBuilder(build)}
                class="group flex w-full items-center gap-2 rounded-lg border border-gray-100 bg-gray-50 px-2.5 py-2 text-left transition-colors hover:border-emerald-200 hover:bg-emerald-50"
              >
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-1.5">
                    <span
                      class="truncate font-mono text-xs font-medium text-gray-800 group-hover:text-emerald-800"
                      >{build.image_name}</span
                    >
                    <Badge
                      class="shrink-0"
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
                  <span class="text-[10px] text-gray-400">{build.software}:{build.version}</span>
                </div>
                <ChevronRight
                  size={10}
                  class="shrink-0 text-gray-300 group-hover:text-emerald-500"
                />
              </button>
            {/each}
          </div>
        {/if}
      </Card>
    </div>
  </div>
</div>
