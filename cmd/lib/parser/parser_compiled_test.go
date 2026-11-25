package parser

import (
	"testing"
)

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

	foundClass := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeClass && n.Name == "UserService" {
			foundClass = true
		}
	}
	if !foundClass {
		t.Error("Expected to find UserService class")
	}

	foundInterface := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeInterface && n.Name == "UserRepository" {
			foundInterface = true
		}
	}
	if !foundInterface {
		t.Error("Expected to find UserRepository interface")
	}

	foundEnum := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeEnum && n.Name == "Status" {
			foundEnum = true
		}
	}
	if !foundEnum {
		t.Error("Expected to find Status enum")
	}

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

	foundStruct := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeStruct && n.Name == "User" {
			foundStruct = true
		}
	}
	if !foundStruct {
		t.Error("Expected to find User struct")
	}

	foundTrait := false
	for _, n := range result.Nodes {
		if n.Name == "Repository" {
			foundTrait = true
		}
	}
	if !foundTrait {
		t.Error("Expected to find Repository trait")
	}

	foundEnum := false
	for _, n := range result.Nodes {
		if n.Type == NodeTypeEnum && n.Name == "Status" {
			foundEnum = true
		}
	}
	if !foundEnum {
		t.Error("Expected to find Status enum")
	}

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

	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
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

	if len(result.Imports) < 2 {
		t.Errorf("Expected at least 2 imports, got %d", len(result.Imports))
	}
}

