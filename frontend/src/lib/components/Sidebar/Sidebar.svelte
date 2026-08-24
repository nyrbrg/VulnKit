<script lang="ts">
  import type { Snippet } from "svelte";

  type SidebarSection = {
    id?: string;
    name: string;
    href: string;
    items?: {
      name: string;
      active: boolean;
      href: string;
      icon?: Snippet;
    }[];
  };

  let { sections }: { sections: SidebarSection[] } = $props();
</script>

<nav class="flex flex-col gap-1 bg-white border-r-[0.5px] border-black/80 py-3">
  {#each sections as section}
    <div class="flex flex-col mb-1">
      <a
        href={section.href}
        class="text-[10px] font-medium text-[#9a9890] tracking-wider uppercase py-2 pl-4"
        >{section.name}</a
      >
      {#if section.items}
        {#each section.items as item}
          <a
            href={item.href}
            class="flex items-center gap-2.25 py-1.75 px-4 text-[13px] text-[#5a5955] hover:bg-[#f8f7f4] active:text-[#0F6E56] active:bg-[#E1F5EE] active:border-l-[#1D9E75] active:font-medium decoration-0 border-l-2 border-solid border-transparent transition-all"
            class:active={item.active}
          >
            {#if item.icon}
              {item.icon}
            {/if}
            {item.name}
          </a>
        {/each}
      {/if}
    </div>
  {/each}
</nav>
