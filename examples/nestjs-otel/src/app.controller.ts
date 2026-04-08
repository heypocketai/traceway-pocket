import { Controller, Get, Post, Param, Body } from "@nestjs/common";

const users = [
  { id: "1", name: "Alice", email: "alice@example.com" },
  { id: "2", name: "Bob", email: "bob@example.com" },
  { id: "3", name: "Charlie", email: "charlie@example.com" },
];

@Controller("nestjs/api")
export class AppController {
  @Get("users")
  getUsers() {
    return users;
  }

  @Get("users/:id")
  getUser(@Param("id") id: string) {
    const user = users.find((u) => u.id === id);
    if (!user) {
      return { error: "User not found" };
    }
    return user;
  }

  @Post("users")
  createUser(@Body() body: { name: string; email: string }) {
    const newUser = { id: String(users.length + 1), ...body };
    users.push(newUser);
    return newUser;
  }

  @Get("slow")
  async slow() {
    await new Promise((resolve) => setTimeout(resolve, 300));
    return { message: "Slow response" };
  }

  @Get("test-error")
  testError() {
    throw new Error("Test error from NestJS");
  }
}
