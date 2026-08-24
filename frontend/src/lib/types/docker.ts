export type DockerHubImage = {
  repo_name: string;
  short_description: string;
  pull_count: number;
  is_official: boolean;
};

export type DockerHubTag = {
  name: string;
};