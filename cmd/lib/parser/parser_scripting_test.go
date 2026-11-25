package parser

import (
	"testing"
)

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

	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "User" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find User class")
	}

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

	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "UserService" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find UserService class")
	}

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

	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "UserService" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find UserService class")
	}

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

	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

// NodeTypeTrait is a custom node type for traits (used in Rust, PHP, Scala, Kotlin)
const NodeTypeTrait NodeType = "trait"

