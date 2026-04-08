import { withRoute } from "@/lib/with-route";

const users = [
  { id: "1", name: "Alice", email: "alice@example.com" },
  { id: "2", name: "Bob", email: "bob@example.com" },
  { id: "3", name: "Charlie", email: "charlie@example.com" },
];

export const GET = withRoute("/nextjs/api/users", async () => {
  return Response.json(users);
});

export const POST = withRoute("/nextjs/api/users", async (req) => {
  const body = await req.json();
  const newUser = { id: String(users.length + 1), ...body };
  users.push(newUser);
  return Response.json(newUser, { status: 201 });
});
