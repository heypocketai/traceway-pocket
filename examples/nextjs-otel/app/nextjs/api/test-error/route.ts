import { withRoute } from "@/lib/with-route";

export const GET = withRoute("/nextjs/api/test-error", async () => {
  throw new Error("Test error from Next.js");
});
