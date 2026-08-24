export type StatusVariant = "success" | "error";

const errorText = (err: unknown) => (err instanceof Error ? err.message : String(err));

export function createStatusMessage() {
  let message = $state("");
  let variant = $state<StatusVariant>("success");

  return {
    get message() {
      return message;
    },
    get variant() {
      return variant;
    },
    success(msg: string) {
      message = msg;
      variant = "success";
    },
    error(err: unknown, prefix = "Error") {
      message = `${prefix}: ${errorText(err)}`;
      variant = "error";
    },
    clear() {
      message = "";
    },
  };
}
