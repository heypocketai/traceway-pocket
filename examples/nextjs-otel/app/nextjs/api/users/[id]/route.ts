import { withRoute } from "@/lib/with-route";

const users = [
  { id: "1", name: "Alice", email: "alice@example.com" },
  { id: "2", name: "Bob", email: "bob@example.com" },
  { id: "3", name: "Charlie", email: "charlie@example.com" },
];

export const GET = withRoute(
  "/nextjs/api/users/[id]",
  async (req, { params }) => {
    const { id } = await params;
    const user = users.find((u) => u.id === id);
    if (!user) {
      return Response.json({ error: "User not found" }, { status: 404 });
    }
    return Response.json(user);
  }
);
