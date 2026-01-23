package com.example.service;

import java.util.ArrayList;
import java.util.List;

public class UserService {
    private Database database;
    private AuthManager authManager;
    private List<User> cachedUsers;

    public UserService(Database db, AuthManager auth) {
        this.database = db;
        this.authManager = auth;
        this.cachedUsers = new ArrayList<>();
    }

    public User getUserById(int id) {
        if (!authManager.isAuthenticated()) {
            throw new RuntimeException("User not authenticated");
        }

        // Check cache first
        for (User user : cachedUsers) {
            if (user.getId() == id) {
                return user;
            }
        }

        // Query database
        User user = database.getUserById(id);
        if (user != null) {
            cachedUsers.add(user);
        }
        return user;
    }

    public void createUser(String email, String password) {
        if (email == null || email.isEmpty()) {
            throw new IllegalArgumentException("Email cannot be empty");
        }

        User user = new User(email, password);
        database.insertUser(user);
        cachedUsers.add(user);
    }

    public List<User> getAllUsers() {
        return database.getAllUsers();
    }

    public void deleteUser(int id) {
        database.deleteUser(id);
        cachedUsers.removeIf(user -> user.getId() == id);
    }

    public void clearCache() {
        cachedUsers.clear();
    }
}
