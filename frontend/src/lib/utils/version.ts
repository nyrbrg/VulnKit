export const versionToNumber = (v: string): number => {
  const parts = v.split(".").map((p) => parseInt(p, 10));
  return (parts[0] || 0) * 10000 + (parts[1] || 0) * 100 + (parts[2] || 0);
};
