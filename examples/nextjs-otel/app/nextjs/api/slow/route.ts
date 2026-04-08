import { withRoute } from "@/lib/with-route";

export const GET = withRoute("/nextjs/api/slow", async () => {
  await new Promise((resolve) => setTimeout(resolve, 300));
  return Response.json({ message: "Slow response" });
});
