<script lang="ts">
  import Badge from "../Badge/Badge.svelte";
  import PanelsTopLeft from "@lucide/svelte/icons/panels-top-left";
</script>

<header
  class="flex items-center justify-between col-span-full px-2.5 bg-white border-b-[0.5px] border-solid border-black/80"
>
  <div class="logo flex items-center gap-2">
    <div
      class="logo-icon flex items-center justify-center w-6.5 h-6.5 rounded-md text-black"
    >
      <PanelsTopLeft size={14} strokeWidth={2.5} />
    </div>
    <span class="text-[14px] font-medium">VulnKit</span>
  </div>
  <Badge>{import.meta.env.VITE_ENV_NAME ?? 'unknown'}</Badge>
</header>
