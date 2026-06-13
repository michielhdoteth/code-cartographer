package parser

import (
	"testing"
)

func TestTreeSitterParserAllLanguages(t *testing.T) {
	// Test that tree-sitter parsers can be created for all languages
	// that have grammar packages available
	languages := []Language{
		LanguageGo,
		LanguageJavaScript,
		LanguageTypeScript,
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageShell,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			parser := NewTreeSitterParser(lang)
			if parser == nil {
				t.Logf("NewTreeSitterParser(%q) returned nil (falls back to regex)", lang)
				return
			}
			if parser.Language() != lang {
				t.Errorf("parser.Language() = %q, want %q", parser.Language(), lang)
			}
		})
	}
}

func TestUnifiedParserAllLanguages(t *testing.T) {
	languages := []Language{
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageShell,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range languages {
		t.Run(string(lang), func(t *testing.T) {
			parser := NewUnifiedParser(lang)
			if parser == nil {
				t.Fatalf("NewUnifiedParser(%q) returned nil", lang)
			}
			if parser.Language() != lang {
				t.Errorf("parser.Language() = %q, want %q", parser.Language(), lang)
			}
		})
	}
}

func TestParseJava(t *testing.T) {
	code := `package com.example;

import java.util.List;
import java.util.Map;

public class UserService {
    private List<String> users;
    
    public void addUser(String name) {
        users.add(name);
    }
    
    public String getUser(int index) {
        return users.get(index);
    }
}

interface UserRepository {
    void save(String user);
}

enum Status {
    ACTIVE, INACTIVE
}
`
	parser := NewUnifiedParser(LanguageJava)
	result, err := parser.Parse(code, "UserService.java")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Java parse")
	}

	// Check for class
	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "UserService" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find UserService class")
	}

	// Check for interface
	foundInterface := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeInterface && n.Name == "UserRepository" {
			foundInterface = true
		}
	}
	if !foundInterface {
		t.Error("Expected to find UserRepository interface")
	}

	// Check for enum
	foundEnum := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeEnum && n.Name == "Status" {
			foundEnum = true
		}
	}
	if !foundEnum {
		t.Error("Expected to find Status enum")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseRust(t *testing.T) {
	code := `use std::collections::HashMap;

pub struct User {
    name: String,
    age: u32,
}

pub trait Repository {
    fn save(&self, user: &User);
    fn find(&self, name: &str) -> Option<User>;
}

pub enum Status {
    Active,
    Inactive,
}

impl User {
    pub fn new(name: String, age: u32) -> Self {
        User { name, age }
    }
}

pub fn main() {
    let user = User::new("Alice".to_string(), 30);
}
`
	parser := NewUnifiedParser(LanguageRust)
	result, err := parser.Parse(code, "main.rs")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Rust parse")
	}

	// Check for struct
	foundStruct := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeStruct && n.Name == "User" {
			foundStruct = true
		}
	}
	if !foundStruct {
		t.Error("Expected to find User struct")
	}

	// Check for trait
	foundTrait := false
	for _, n := range result.Nodes {
		if n.Name == "Repository" {
			foundTrait = true
		}
	}
	if !foundTrait {
		t.Error("Expected to find Repository trait")
	}

	// Check for enum
	foundEnum := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeEnum && n.Name == "Status" {
			foundEnum = true
		}
	}
	if !foundEnum {
		t.Error("Expected to find Status enum")
	}

	// Check imports
	if len(result.Imports) == 0 {
		t.Error("Expected at least 1 import")
	}
}

func TestParseC(t *testing.T) {
	code := `#include <stdio.h>
#include "utils.h"

typedef struct {
    int x;
    int y;
} Point;

enum Color {
    RED, GREEN, BLUE
};

int calculate(int a, int b) {
    return a + b;
}

void print_point(Point p) {
    printf("(%d, %d)\\n", p.x, p.y);
}
`
	parser := NewUnifiedParser(LanguageC)
	result, err := parser.Parse(code, "main.c")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from C parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseRuby(t *testing.T) {
	code := `require 'json'
require_relative 'helpers'

class User
  attr_reader :name, :age

  def initialize(name, age)
    @name = name
    @age = age
  end

  def to_json
    { name: @name, age: @age }.to_json
  end
end

module Utils
  def self.format_name(name)
    name.strip.capitalize
  end
end
`
	parser := NewUnifiedParser(LanguageRuby)
	result, err := parser.Parse(code, "user.rb")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Ruby parse")
	}

	// Check for class
	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "User" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find User class")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseShell(t *testing.T) {
	code := `#!/bin/bash

source ./utils.sh
. ./config.sh

function greet() {
    echo "Hello, $1!"
}

deploy() {
    echo "Deploying..."
}

main() {
    greet "World"
    deploy
}
`
	parser := NewUnifiedParser(LanguageShell)
	result, err := parser.Parse(code, "script.sh")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Shell parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParsePython(t *testing.T) {
	code := `import os
from pathlib import Path

class UserService:
    def __init__(self):
        self.users = []
    
    def add_user(self, name):
        self.users.append(name)

def main():
    service = UserService()
    service.add_user("Alice")

if __name__ == "__main__":
    main()
`
	parser := NewUnifiedParser(LanguagePython)
	result, err := parser.Parse(code, "main.py")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Python parse")
	}

	// Check for class
	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "UserService" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find UserService class")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseCSharp(t *testing.T) {
	code := `using System;
using System.Collections.Generic;

namespace MyApp
{
    public class UserService
    {
        private List<string> _users = new List<string>();
        
        public void AddUser(string name)
        {
            _users.Add(name);
        }
    }

    public interface IRepository
    {
        void Save(string item);
    }

    public enum Status
    {
        Active,
        Inactive
    }
}
`
	parser := NewUnifiedParser(LanguageCSharp)
	result, err := parser.Parse(code, "UserService.cs")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from C# parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParsePHP(t *testing.T) {
	code := `<?php
namespace App\\Services;

use App\\Models\\User;
use App\\Repositories\\UserRepository;

class UserService
{
    private $repository;
    
    public function __construct(UserRepository $repo)
    {
        $this->repository = $repo;
    }
    
    public function addUser(string $name): void
    {
        $user = new User($name);
        $this->repository->save($user);
    }
}

interface CacheInterface
{
    public function get(string $key);
    public function set(string $key, $value);
}

trait Serializable
{
    public function serialize()
    {
        return json_encode($this);
    }
}
`
	parser := NewUnifiedParser(LanguagePHP)
	result, err := parser.Parse(code, "UserService.php")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from PHP parse")
	}

	// Check for class
	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "UserService" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find UserService class")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseSwift(t *testing.T) {
	code := `import Foundation
import UIKit

class ViewController: UIViewController {
    override func viewDidLoad() {
        super.viewDidLoad()
    }
    
    func fetchData() {
        print("Fetching...")
    }
}

struct User {
    let name: String
    let age: Int
}

enum Direction {
    case north, south, east, west
}

protocol Repository {
    func save(_ item: Any)
}
`
	parser := NewUnifiedParser(LanguageSwift)
	result, err := parser.Parse(code, "ViewController.swift")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Swift parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseKotlin(t *testing.T) {
	code := `package com.example

import kotlin.collections.List
import kotlinx.coroutines.*

class UserService {
    private val users = mutableListOf<String>()
    
    fun addUser(name: String) {
        users.add(name)
    }
    
    fun getUser(index: Int): String? {
        return users.getOrNull(index)
    }
}

interface Repository {
    fun save(item: String)
    fun find(id: Int): String?
}

object AppConfig {
    val appName = "MyApp"
    var debug = false
}
`
	parser := NewUnifiedParser(LanguageKotlin)
	result, err := parser.Parse(code, "UserService.kt")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Kotlin parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseScala(t *testing.T) {
	code := `package com.example

import scala.collection.mutable
import akka.actor.Actor

class UserService {
  private val users = mutable.ListBuffer[String]()
  
  def addUser(name: String): Unit = {
    users += name
  }
}

trait Repository {
  def save(item: String): Unit
  def find(id: Int): Option[String]
}

object AppConfig {
  val appName = "MyApp"
}
`
	parser := NewUnifiedParser(LanguageScala)
	result, err := parser.Parse(code, "UserService.scala")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Scala parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseHTML(t *testing.T) {
	code := `<!DOCTYPE html>
<html>
<head>
    <title>My Page</title>
    <link rel="stylesheet" href="style.css">
    <script src="app.js"></script>
</head>
<body>
    <h1>Hello World</h1>
    <div class="content">
        <p>Content here</p>
    </div>
</body>
</html>
`
	parser := NewUnifiedParser(LanguageHTML)
	result, err := parser.Parse(code, "index.html")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from HTML parse")
	}

	// Check imports (link/script sources)
	if len(result.Imports) < 1 {
		t.Error("Expected at least 1 import from HTML")
	}
}

func TestParseCSS(t *testing.T) {
	code := `@media (max-width: 768px) {
    .container {
        width: 100%;
    }
}

.header {
    background-color: #333;
    color: white;
}

#main-content {
    padding: 20px;
}

@keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}
`
	parser := NewUnifiedParser(LanguageCSS)
	result, err := parser.Parse(code, "style.css")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// CSS parser should work without errors
	if result == nil {
		t.Fatal("Expected non-nil result from CSS parse")
	}
}

func TestParseSQL(t *testing.T) {
	code := `CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(255)
);

CREATE INDEX idx_users_email ON users(email);

CREATE VIEW active_users AS
SELECT * FROM users WHERE active = true;

CREATE FUNCTION get_user(id INT)
RETURNS VARCHAR(255)
BEGIN
    RETURN SELECT name FROM users WHERE id = id;
END;
`
	parser := NewUnifiedParser(LanguageSQL)
	result, err := parser.Parse(code, "schema.sql")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result from SQL parse")
	}
}

func TestParseYAML(t *testing.T) {
	code := `name: my-project
version: 1.0.0
dependencies:
  - react: ^18.0.0
  - next: ^14.0.0
scripts:
  dev: next dev
  build: next build
`
	parser := NewUnifiedParser(LanguageYAML)
	result, err := parser.Parse(code, "package.yaml")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result from YAML parse")
	}
}

func TestParseJSON(t *testing.T) {
	code := `{
  "name": "my-project",
  "version": "1.0.0",
  "dependencies": {
    "react": "^18.0.0"
  }
}`
	parser := NewUnifiedParser(LanguageJSON)
	result, err := parser.Parse(code, "package.json")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result from JSON parse")
	}
}

func TestParseTOML(t *testing.T) {
	code := `[package]
name = "my-project"
version = "0.1.0"

[dependencies]
serde = "1.0"
tokio = { version = "1.0", features = ["full"] }

[dev-dependencies]
criterion = "0.5"
`
	parser := NewUnifiedParser(LanguageTOML)
	result, err := parser.Parse(code, "Cargo.toml")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result from TOML parse")
	}
}

func TestParseMarkdown(t *testing.T) {
	code := `# My Project

## Installation

Run the following:

` + "```bash" + `
npm install
` + "```" + `

## Usage

Use it like this.
`
	parser := NewUnifiedParser(LanguageMarkdown)
	result, err := parser.Parse(code, "README.md")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result from Markdown parse")
	}
}

func TestParseCpp(t *testing.T) {
	code := `#include <iostream>
#include <vector>
#include "utils.h"

namespace MyApp {

class UserService {
public:
    void addUser(const std::string& name) {
        users_.push_back(name);
    }
    
    std::string getUser(int index) const {
        return users_[index];
    }

private:
    std::vector<std::string> users_;
};

} // namespace MyApp
`
	parser := NewUnifiedParser(LanguageCpp)
	result, err := parser.Parse(code, "UserService.cpp")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from C++ parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseGo(t *testing.T) {
	code := `package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Hello, World!")
}

func helper(x int) int {
    return x * 2
}
`
	parser := NewUnifiedParser(LanguageGo)
	result, err := parser.Parse(code, "main.go")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from Go parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

func TestParseJavaScript(t *testing.T) {
	code := `import React from 'react';
import { useState } from 'react';

function App() {
    const [count, setCount] = useState(0);
    return <div>{count}</div>;
}

class MyComponent extends React.Component {
    render() {
        return <div>Hello</div>;
    }
}

export default App;
`
	parser := NewUnifiedParser(LanguageJavaScript)
	result, err := parser.Parse(code, "App.js")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(result.Nodes) == 0 {
		t.Error("Expected nodes from JavaScript parse")
	}

	// Check imports
	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

// NodeTypeTrait is a custom node type for traits (used in Rust, PHP, Scala, Kotlin)
const NodeTypeTrait NodeType = "trait"

func TestRegisterAllParsers(t *testing.T) {
	// Reset the registry
	parserRegistry = make(map[Language]Parser)

	// Register all parsers
	RegisterTreeSitterParsers()
	RegisterUnifiedParsers()

	// Check that all 21 languages have parsers
	supported := GetSupportedLanguages()
	if len(supported) < 21 {
		t.Errorf("Expected at least 21 supported languages, got %d: %v", len(supported), supported)
	}

	// Verify specific languages have parsers
	mustHave := []Language{
		LanguageGo,
		LanguageJavaScript,
		LanguageTypeScript,
		LanguagePython,
		LanguageJava,
		LanguageC,
		LanguageCpp,
		LanguageCSharp,
		LanguageRuby,
		LanguagePHP,
		LanguageRust,
		LanguageSwift,
		LanguageKotlin,
		LanguageScala,
		LanguageHTML,
		LanguageCSS,
		LanguageSQL,
		LanguageShell,
		LanguageYAML,
		LanguageJSON,
		LanguageTOML,
		LanguageMarkdown,
	}

	for _, lang := range mustHave {
		if GetParser(lang) == nil {
			t.Errorf("No parser registered for %q", lang)
		}
	}
}
