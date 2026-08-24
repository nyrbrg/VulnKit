<script lang="ts">
  import "../app.css";
  import { page } from "$app/state";
  import Sidebar from "$lib/components/Sidebar/Sidebar.svelte";
  import Topbar from "$lib/components/Topbar/Topbar.svelte";
</script>

<div class="grid grid-cols-[220px_1fr] grid-rows-[48px_1fr] min-h-screen">
  <Topbar />

  <Sidebar
    sections={[
      {
        name: "Dashboard",
        href: "/",
      },
      {
        name: "Lab",
        href: "/lab",
        items: [
          {
            name: "Lab builder",
            href: "/lab/lab-builder",
            active: page.url.pathname === "/lab/lab-builder",
          },
          {
            name: "Presets",
            href: "/lab/presets",
            active: page.url.pathname === "/lab/presets",
          },
          {
            name: "Local Images",
            href: "/lab/local-images",
            active: page.url.pathname === "/lab/local-images",
          },
        ],
      },
      {
        name: "CVE",
        href: "/cve",
      },
    ]}
  />

  <main class="p-6 overflow-y-auto">
    <slot />
  </main>
</div>
