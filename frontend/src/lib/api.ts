import type { DockerHubImage, DockerHubTag } from "./types/docker";

const BASE = `${import.meta.env.VITE_API_URL}/api`;

export type ServiceConfig = {
  name: string;
  image: string;
  version: string;
  ports: string[];
  env_vars: Record<string, string>;
};

export type LabConfig = {
  name: string;
  services: ServiceConfig[];
};

export type ContainerStatus = {
  id: string;
  name: string;
  image: string;
  status: string;
  state: string;
  ports: string;
};

export type Preset = {
  id: string;
  name: string;
  tags: string[];
  services: ServiceConfig[];
  created_at: string;
};

export type DefaultLab = {
  id: string;
  title: string;
  category: string;
  difficulty: "apprentice" | "practitioner" | "expert";
  description: string;
  objective: string;
  hint: string;
  solution: string;
  path: string;
};

export type CVEItem = {
  id: string;
  sourceIdentifier: string;
  published: string;
  lastModified: string;
  vulnStatus: string;
  descriptions: { lang: string; value: string }[];
  metrics?: {
    cvssMetricV31?: {
      cvssData: {
        baseScore: number;
        baseSeverity: string;
      };
    }[];
    cvssMetricV2?: {
      cvssData: {
        baseScore: number;
      };
      baseSeverity: string;
    }[];
  };
  references: { url: string; source: string; tags?: string[] }[];
  configurations?: any;
};

export type NVDResponse = {
  resultsPerPage: number;
  startIndex: number;
  totalResults: number;
  vulnerabilities: { cve: CVEItem }[];
};

type DynamicBuildCheck = {
  supported: boolean;
  base_image?: string;
  package?: string;
  image_name?: string;
  message?: string;
};

export type DynamicBuild = {
  id: string;
  software: string;
  version: string;
  requested_version: string;
  image_name: string;
  build_dir: string;
  created_at: string;
};

const req = async <T>(path: string, options?: RequestInit): Promise<T> => {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error ?? "Request failed");
  return data as T;
};

export const api = {
  health: () => req<{ status: string; docker: boolean }>("/health"),

  generateCompose: (config: LabConfig) =>
    req<{ yaml: string }>("/labs/generate", {
      method: "POST",
      body: JSON.stringify(config),
    }),

  startLab: (config: LabConfig) =>
    req<{ status: string; output: string }>("/labs/start", {
      method: "POST",
      body: JSON.stringify(config),
    }),

  stopLab: (labName: string) =>
    req<{ status: string }>("/labs/stop", {
      method: "POST",
      body: JSON.stringify({ lab_name: labName }),
    }),

  stopContainer: (id: string) =>
    req<{ status: string }>(`/labs/containers/${id}/stop`, { method: "POST" }),

  startContainer: (id: string) =>
    req<{ status: string }>(`/labs/containers/${id}/start`, { method: "POST" }),

  removeContainer: (id: string) =>
    req<{ status: string }>(`/labs/containers/${id}`, { method: "DELETE" }),

  getStatus: () => req<{ containers: ContainerStatus[] }>("/labs/status"),

  getLogs: (id: string, tail = 100) => req<{ logs: string }>(`/labs/logs/${id}?tail=${tail}`),

  listPresets: () => req<{ presets: Preset[] }>("/presets"),

  savePreset: (preset: Omit<Preset, "id" | "created_at">) =>
    req<{ status: string }>("/presets", {
      method: "POST",
      body: JSON.stringify(preset),
    }),

  deletePreset: (id: string) => req<{ status: string }>(`/presets/${id}`, { method: "DELETE" }),

  hubSearch: (query: string) =>
    req<{ results: DockerHubImage[] }>(`/hub/search?q=${encodeURIComponent(query)}`),

  hubTags: (image: string) =>
    req<{ results: DockerHubTag[] }>(`/hub/tags?image=${encodeURIComponent(image)}`),

  listDefaultLabs: () => req<{ labs: DefaultLab[] }>("/labs/default"),
  defaultLabStatus: () => req<{ running: boolean }>("/labs/default/status"),

  startDefaultLab: () => req<{ status: string }>("/labs/default/start", { method: "POST" }),

  stopDefaultLab: () => req<{ status: string }>("/labs/default/stop", { method: "POST" }),

  cveSearch: (q: string, startIndex = 0, pageSize = 20, signal?: AbortSignal) =>
    req<NVDResponse>(
      `/cve/search?q=${encodeURIComponent(q)}&startIndex=${startIndex}&pageSize=${pageSize}`,
      { signal }
    ),
  cveDetail: (id: string) => req<NVDResponse>(`/cve/${encodeURIComponent(id)}`),

  dynamicBuildCheck: (image: string, version: string) =>
    req<DynamicBuildCheck>(
      `/labs/dynamic/check?image=${encodeURIComponent(image)}&version=${encodeURIComponent(version)}`
    ),

  dynamicBuildStream: (image: string, version: string): EventSource => {
    return new EventSource(
      `${import.meta.env.VITE_API_URL ?? "http://localhost:8080"}/api/labs/dynamic/build`
    );
  },
  dynamicBuild: async (
    image: string,
    version: string,
    onLog: (line: string) => void,
    onDone: (imageName: string) => void,
    onError: (msg: string) => void
  ) => {
    const BASE = import.meta.env.VITE_API_URL ?? "http://localhost:8080";
    const res = await fetch(`${BASE}/api/labs/dynamic/build`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ image, version }),
    });

    if (!res.body) {
      onError("No response body");
      return;
    }

    const reader = res.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let finished = false;

    while (!finished) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });

      const messages = buffer.split("\n\n");
      buffer = messages.pop() ?? "";

      for (const message of messages) {
        if (!message.trim()) continue;

        const lines = message.split("\n");
        let eventType = "log";
        let data = "";

        for (const line of lines) {
          if (line.startsWith("event: ")) {
            eventType = line.slice(7).trim();
          } else if (line.startsWith("data: ")) {
            data = line.slice(6).trim();
          }
        }

        if (!data) continue;

        if (eventType === "done") {
          onDone(data);
          finished = true;
          break;
        } else if (eventType === "error") {
          onError(data);
          finished = true;
          break;
        } else {
          onLog(data);
        }
      }
    }

    if (!finished && buffer.trim()) {
      const lines = buffer.split("\n");
      let eventType = "log";
      let data = "";
      for (const line of lines) {
        if (line.startsWith("event: ")) eventType = line.slice(7).trim();
        else if (line.startsWith("data: ")) data = line.slice(6).trim();
      }
      if (data) {
        if (eventType === "done") onDone(data);
        else if (eventType === "error") onError(data);
      }
    }
  },

  listDynamicBuilds: () => req<{ builds: DynamicBuild[] }>("/labs/dynamic/builds"),

  getBuildDockerfile: (id: string) =>
    req<{ dockerfile: string }>(`/labs/dynamic/builds/${id}/dockerfile`),

  deleteDynamicBuild: (id: string) =>
    req<{ status: string }>(`/labs/dynamic/builds/${id}`, { method: "DELETE" }),
  checkPort: (port: string) =>
    req<{ in_use: boolean; container?: string }>(`/labs/ports/check?port=${port}`),
};
