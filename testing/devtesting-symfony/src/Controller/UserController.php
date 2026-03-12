<?php

namespace App\Controller;

use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Attribute\Route;

class UserController
{
    private \PDO $db;

    public function __construct()
    {
        $dbPath = dirname(__DIR__, 2) . '/var/devtesting.db';
        $this->db = new \PDO("sqlite:$dbPath");
        $this->db->setAttribute(\PDO::ATTR_ERRMODE, \PDO::ERRMODE_EXCEPTION);

        $this->db->exec('CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            first_name TEXT NOT NULL,
            last_name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL
        )');
    }

    #[Route('/users', methods: ['GET'])]
    public function list(): JsonResponse
    {
        $stmt = $this->db->query('SELECT id, first_name, last_name, email FROM users');
        $users = $stmt->fetchAll(\PDO::FETCH_ASSOC);

        return new JsonResponse($users);
    }

    #[Route('/users/{id}', methods: ['GET'], requirements: ['id' => '\d+'])]
    public function get(int $id): JsonResponse
    {
        $stmt = $this->db->prepare('SELECT id, first_name, last_name, email FROM users WHERE id = ?');
        $stmt->execute([$id]);
        $user = $stmt->fetch(\PDO::FETCH_ASSOC);

        if (!$user) {
            return new JsonResponse(['error' => 'user not found'], Response::HTTP_NOT_FOUND);
        }

        return new JsonResponse($user);
    }

    #[Route('/users', methods: ['POST'])]
    public function create(Request $request): JsonResponse
    {
        $data = json_decode($request->getContent(), true);

        if (!$data || !isset($data['first_name'], $data['last_name'], $data['email'])) {
            return new JsonResponse(['error' => 'missing required fields'], Response::HTTP_BAD_REQUEST);
        }

        $stmt = $this->db->prepare('INSERT INTO users (first_name, last_name, email) VALUES (?, ?, ?)');
        $stmt->execute([$data['first_name'], $data['last_name'], $data['email']]);

        $id = (int) $this->db->lastInsertId();

        return new JsonResponse([
            'id' => $id,
            'first_name' => $data['first_name'],
            'last_name' => $data['last_name'],
            'email' => $data['email'],
        ], Response::HTTP_CREATED);
    }

    #[Route('/users/{id}', methods: ['PUT'], requirements: ['id' => '\d+'])]
    public function update(int $id, Request $request): JsonResponse
    {
        $data = json_decode($request->getContent(), true);

        if (!$data || !isset($data['first_name'], $data['last_name'], $data['email'])) {
            return new JsonResponse(['error' => 'missing required fields'], Response::HTTP_BAD_REQUEST);
        }

        $stmt = $this->db->prepare('UPDATE users SET first_name = ?, last_name = ?, email = ? WHERE id = ?');
        $stmt->execute([$data['first_name'], $data['last_name'], $data['email'], $id]);

        if ($stmt->rowCount() === 0) {
            return new JsonResponse(['error' => 'user not found'], Response::HTTP_NOT_FOUND);
        }

        return new JsonResponse([
            'id' => $id,
            'first_name' => $data['first_name'],
            'last_name' => $data['last_name'],
            'email' => $data['email'],
        ]);
    }

    #[Route('/users/{id}', methods: ['DELETE'], requirements: ['id' => '\d+'])]
    public function delete(int $id): JsonResponse
    {
        $stmt = $this->db->prepare('DELETE FROM users WHERE id = ?');
        $stmt->execute([$id]);

        if ($stmt->rowCount() === 0) {
            return new JsonResponse(['error' => 'user not found'], Response::HTTP_NOT_FOUND);
        }

        return new JsonResponse(['message' => 'user deleted']);
    }
}
