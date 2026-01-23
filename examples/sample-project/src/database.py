"""Database Module - Handles all database operations"""

import sqlite3
from typing import List, Dict, Any

class Database:
    """Main database connection handler"""

    def __init__(self, db_path: str):
        """Initialize database connection"""
        self.db_path = db_path
        self.connection = None

    def connect(self):
        """Connect to database"""
        self.connection = sqlite3.connect(self.db_path)
        self.create_tables()

    def create_tables(self):
        """Create required tables"""
        cursor = self.connection.cursor()

        # Users table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY,
                email TEXT UNIQUE,
                password TEXT,
                created_at TIMESTAMP
            )
        ''')

        # Posts table
        cursor.execute('''
            CREATE TABLE IF NOT EXISTS posts (
                id INTEGER PRIMARY KEY,
                user_id INTEGER,
                content TEXT,
                created_at TIMESTAMP,
                FOREIGN KEY (user_id) REFERENCES users(id)
            )
        ''')

        self.connection.commit()

    def get_user(self, user_id: int) -> Dict[str, Any]:
        """Get user by ID"""
        cursor = self.connection.cursor()
        cursor.execute('SELECT * FROM users WHERE id = ?', (user_id,))
        return cursor.fetchone()

    def create_user(self, email: str, password: str) -> int:
        """Create new user"""
        cursor = self.connection.cursor()
        cursor.execute(
            'INSERT INTO users (email, password) VALUES (?, ?)',
            (email, password)
        )
        self.connection.commit()
        return cursor.lastrowid

    def get_user_posts(self, user_id: int) -> List[Dict[str, Any]]:
        """Get all posts by user"""
        cursor = self.connection.cursor()
        cursor.execute(
            'SELECT * FROM posts WHERE user_id = ?',
            (user_id,)
        )
        return cursor.fetchall()

    def close(self):
        """Close database connection"""
        if self.connection:
            self.connection.close()
