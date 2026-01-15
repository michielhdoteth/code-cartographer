"""
Framework Detector for Code Cartographer

Detects frameworks and libraries used in a codebase based on:
- Configuration files (package.json, requirements.txt, etc.)
- Import patterns
- File markers
"""

import os
import re
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Set, Tuple


@dataclass
class FrameworkInfo:
    name: str
    category: str
    confidence: float
    indicators: List[str]
    version: Optional[str] = None


@dataclass
class ProjectFrameworks:
    frameworks: List[FrameworkInfo] = field(default_factory=list)
    language: str = ""
    build_tools: List[str] = field(default_factory=list)
    testing_frameworks: List[str] = field(default_factory=list)
    databases: List[str] = field(default_factory=list)


class FrameworkDetector:
    FRAMEWORK_MARKERS = {
        "web": {
            "python": {
                "django": {
                    "files": ["requirements.txt", "Pipfile", "pyproject.toml"],
                    "patterns": [r"django", r"from django", r"import django"],
                    "markers": ["DJANGO_SETTINGS_MODULE", "django.setup()"],
                },
                "flask": {
                    "files": ["requirements.txt"],
                    "patterns": [r"flask", r"from flask", r"import flask"],
                    "markers": ["Flask", "app = Flask", "@app.route"],
                },
                "fastapi": {
                    "files": ["requirements.txt", "pyproject.toml"],
                    "patterns": [r"fastapi", r"from fastapi", r"import fastapi"],
                    "markers": ["FastAPI", "app = FastAPI", "@app.get"],
                },
                "pyramid": {
                    "files": ["requirements.txt", "setup.py"],
                    "patterns": [r"pyramid", r"from pyramid"],
                    "markers": ["pyramid.config"],
                },
                "bottle": {
                    "files": ["requirements.txt"],
                    "patterns": [r"bottle", r"from bottle"],
                    "markers": ["Bottle", "app = Bottle"],
                },
                "cherrypy": {
                    "files": ["requirements.txt"],
                    "patterns": [r"cherrypy", r"from cherrypy"],
                    "markers": ["cherrypy"],
                },
                "tornado": {
                    "files": ["requirements.txt"],
                    "patterns": [r"tornado", r"from tornado"],
                    "markers": ["tornado.web", "tornado.ioloop"],
                },
                "aiohttp": {
                    "files": ["requirements.txt"],
                    "patterns": [r"aiohttp", r"from aiohttp"],
                    "markers": ["aiohttp", "aiohttp.web"],
                },
                "starlette": {
                    "files": ["requirements.txt"],
                    "patterns": [r"starlette", r"from starlette"],
                    "markers": ["Starlette", "app = Starlette"],
                },
            },
            "javascript": {
                "react": {
                    "files": ["package.json"],
                    "patterns": [r"react", r"@react"],
                    "markers": ["import React", "import { useState }", "useEffect"],
                },
                "vue": {
                    "files": ["package.json"],
                    "patterns": [r"vue", r"@vue"],
                    "markers": ["import Vue", "new Vue", "Vue.component"],
                },
                "angular": {
                    "files": ["package.json", "angular.json"],
                    "patterns": [r"@angular", r"angular"],
                    "markers": ["@Component", "@NgModule", "import { Component }"],
                },
                "express": {
                    "files": ["package.json"],
                    "patterns": [r"express"],
                    "markers": ["express()", "app.use", "@app.get"],
                },
                "next": {
                    "files": ["package.json", "next.config.js"],
                    "patterns": [r"next", r"@next"],
                    "markers": ["next.config", "getStaticProps", "getServerSideProps"],
                },
                "nuxt": {
                    "files": ["package.json", "nuxt.config.js"],
                    "patterns": [r"nuxt", r"@nuxt"],
                    "markers": ["nuxt.config", "asyncData", "fetch"],
                },
                "svelte": {
                    "files": ["package.json"],
                    "patterns": [r"svelte"],
                    "markers": ["import Svelte", "<script>", "export let"],
                },
                "astro": {
                    "files": ["package.json", "astro.config.mjs"],
                    "patterns": [r"astro"],
                    "markers": ["astro.config", "---", "frontmatter"],
                },
                "vite": {
                    "files": ["package.json", "vite.config.js"],
                    "patterns": [r"vite"],
                    "markers": ["vite.config", "createVite"],
                },
                "webpack": {
                    "files": ["package.json", "webpack.config.js"],
                    "patterns": [r"webpack"],
                    "markers": ["webpack.config", "HtmlWebpackPlugin"],
                },
            },
            "typescript": {
                "react": {
                    "files": ["package.json"],
                    "patterns": [r"react", r"@react"],
                    "markers": ["import React", "React.FC", "useState<"],
                },
                "next": {
                    "files": ["package.json", "next.config.ts"],
                    "patterns": [r"next", r"@next"],
                    "markers": ["next.config", "getStaticProps"],
                },
                "nest": {
                    "files": ["package.json"],
                    "patterns": [r"@nestjs", r"nestjs"],
                    "markers": ["@Controller", "@Injectable", "@Module"],
                },
                "typeorm": {
                    "files": ["package.json"],
                    "patterns": [r"typeorm"],
                    "markers": ["@Entity", "@Column", "DataSource"],
                },
            },
            "java": {
                "spring": {
                    "files": ["pom.xml", "build.gradle", "build.gradle.kts"],
                    "patterns": [r"org\.springframework", r"springframework"],
                    "markers": ["@RestController", "@Service", "@Repository", "@Component"],
                },
                "quarkus": {
                    "files": ["pom.xml", "build.gradle"],
                    "patterns": [r"io.quarkus"],
                    "markers": ["@ApplicationPath", "@Path", "Quarkus"],
                },
                "micronaut": {
                    "files": ["pom.xml", "build.gradle"],
                    "patterns": [r"io.micronaut"],
                    "markers": ["@Controller", "@Singleton", "Micronaut"],
                },
                "javalin": {
                    "files": ["pom.xml"],
                    "patterns": [r"javalin"],
                    "markers": ["Javalin", "app.get", "app.post"],
                },
                "spark": {
                    "files": ["pom.xml"],
                    "patterns": [r"com.sparkjava"],
                    "markers": ["Spark.get", "Spark.post"],
                },
            },
            "go": {
                "gin": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/gin-gonic/gin"],
                    "markers": ["gin.Engine", "gin.Default", "r.GET"],
                },
                "echo": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/labstack/echo"],
                    "markers": ["echo.New", "e.GET", "e.POST"],
                },
                "fiber": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/gofiber/fiber"],
                    "markers": ["fiber.New", "app.Get", "app.Post"],
                },
                "chi": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/go-chi/chi"],
                    "markers": ["chi.NewRouter", "r.Get", "r.Post"],
                },
                "gorilla": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/gorilla/mux"],
                    "markers": ["mux.NewRouter", "r.Handle", "r.Methods"],
                },
                "beego": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/astaxie/beego"],
                    "markers": ["beego.App", "beego.Router", "get,post"],
                },
                "revel": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/revel/revel"],
                    "markers": ["revel.Controller", "@Get", "@Post"],
                },
                "echo": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/labstack/echo"],
                    "markers": ["echo.New", "e.GET", "e.POST"],
                },
            },
            "rust": {
                "actix": {
                    "files": ["Cargo.toml"],
                    "patterns": [r"actix-web", r"actix"],
                    "markers": ["#[actix_web]", "HttpServer", "App::new()"],
                },
                "rocket": {
                    "files": ["Cargo.toml"],
                    "patterns": [r"rocket"],
                    "markers": ["#[rocket]", "rocket::", "#[get]"],
                },
                "warp": {
                    "files": ["Cargo.toml"],
                    "patterns": [r"warp"],
                    "markers": ["warp::", "Filter::boxed"],
                },
                "axum": {
                    "files": ["Cargo.toml"],
                    "patterns": [r"axum"],
                    "markers": ["axum::", "Router::new"],
                },
                "tokio": {
                    "files": ["Cargo.toml"],
                    "patterns": [r"tokio"],
                    "markers": ["tokio::", "#[tokio::main]"],
                },
                "yew": {
                    "files": ["Cargo.toml"],
                    "patterns": [r"yew"],
                    "markers": ["yew::", "#[function_component]"],
                },
            },
            "cpp": {
                "qt": {
                    "files": ["CMakeLists.txt", ".pro"],
                    "patterns": [r"Qt", r"Q_"],
                    "markers": ["Q_OBJECT", "QApplication", "QWidget"],
                },
                "boost": {
                    "files": ["CMakeLists.txt"],
                    "patterns": [r"boost"],
                    "markers": ["boost::", "#include <boost/"],
                },
                "gtk": {
                    "files": ["CMakeLists.txt"],
                    "patterns": [r"gtk"],
                    "markers": ["GtkWidget", "gtk_init"],
                },
                "wxwidgets": {
                    "files": ["CMakeLists.txt"],
                    "patterns": [r"wx"],
                    "markers": ["wxApp", "wxFrame", "wxWidget"],
                },
            },
            "ruby": {
                "rails": {
                    "files": ["Gemfile", "config/application.rb"],
                    "patterns": [r"rails", r"rails-"],
                    "markers": ["Rails::Application", "ApplicationController"],
                },
                "sinatra": {
                    "files": ["Gemfile"],
                    "patterns": [r"sinatra"],
                    "markers": ["require 'sinatra'", "get '/", "post '/'"],
                },
                "hanami": {
                    "files": ["Gemfile", "config/app.rb"],
                    "patterns": [r"hanami"],
                    "markers": ["Hanami::", "Action::Base"],
                },
                "padrino": {
                    "files": ["Gemfile"],
                    "patterns": [r"padrino"],
                    "markers": ["Padrino::", "get '/", "post '/'"],
                },
                "grape": {
                    "files": ["Gemfile"],
                    "patterns": [r"grape"],
                    "markers": ["class API < Grape::API"],
                },
            },
            "php": {
                "laravel": {
                    "files": ["composer.json", "artisan"],
                    "patterns": [r"laravel", r"illuminate"],
                    "markers": ["Illuminate\\", "Route::", "app/Http/Controllers"],
                },
                "symfony": {
                    "files": ["composer.json", "config/routes.yaml"],
                    "patterns": [r"symfony", r"symfony/"],
                    "markers": ["Symfony\\", "Controller\\", "@Route"],
                },
                "yii": {
                    "files": ["composer.json", "config/web.php"],
                    "patterns": [r"yii", r"yiisoft"],
                    "markers": ["Yii::$", "yii\web", "Controller"],
                },
                "codeigniter": {
                    "files": ["composer.json", "app/Controllers"],
                    "patterns": [r"codeigniter", r"CodeIgniter"],
                    "markers": ["CodeIgniter\\", "Controllers\\"],
                },
                "cakephp": {
                    "files": ["composer.json", "config/bootstrap.php"],
                    "patterns": [r"cakephp", r"cakephp/"],
                    "markers": ["Cake\\", "Controller\\", "App::"],
                },
                "drupal": {
                    "files": ["composer.json", "core/lib/Drupal"],
                    "patterns": [r"drupal"],
                    "markers": ["Drupal\\", "@Block", "@Controller"],
                },
                "wordpress": {
                    "files": ["wp-config.php", "wp-content/themes"],
                    "patterns": [r"wordpress", r"wp-"],
                    "markers": ["WP_Query", "add_action", "the_post"],
                },
                "slim": {
                    "files": ["composer.json"],
                    "patterns": [r"slim", r"slimframework"],
                    "markers": ["Slim\\App", "app->get", "app->post"],
                },
                "phpunit": {
                    "files": ["phpunit.xml", "phpunit.xml.dist"],
                    "patterns": [r"phpunit"],
                    "markers": ["PHPUnit\\Framework", "@Test", "TestCase"],
                },
            },
        },
        "ml_ai": {
            "python": {
                "tensorflow": {
                    "files": ["requirements.txt", "pyproject.toml"],
                    "patterns": [r"tensorflow", r"tf\."],
                    "markers": ["import tensorflow", "tf.keras", "tf.data"],
                },
                "pytorch": {
                    "files": ["requirements.txt", "pyproject.toml"],
                    "patterns": [r"torch", r"torch\."],
                    "markers": ["import torch", "nn.Module", "torch.nn"],
                },
                "keras": {
                    "files": ["requirements.txt"],
                    "patterns": [r"keras", r"tf\.keras"],
                    "markers": ["import keras", "Sequential", "Dense"],
                },
                "scikit_learn": {
                    "files": ["requirements.txt"],
                    "patterns": [r"sklearn", r"scikit-learn"],
                    "markers": ["import sklearn", "from sklearn"],
                },
                "pandas": {
                    "files": ["requirements.txt"],
                    "patterns": [r"pandas", r"pd\."],
                    "markers": ["import pandas", "pd.DataFrame", "pd.Series"],
                },
                "numpy": {
                    "files": ["requirements.txt"],
                    "patterns": [r"numpy", r"np\."],
                    "markers": ["import numpy", "np.array", "np.ndarray"],
                },
                "opencv": {
                    "files": ["requirements.txt"],
                    "patterns": [r"opencv", r"cv2"],
                    "markers": ["import cv2", "cv2.imread", "cv2.resize"],
                },
                "transformers": {
                    "files": ["requirements.txt"],
                    "patterns": [r"transformers", r"huggingface"],
                    "markers": ["from transformers", "AutoModel", "AutoTokenizer"],
                },
                "jax": {
                    "files": ["requirements.txt"],
                    "patterns": [r"jax", r"jaxlib"],
                    "markers": ["import jax", "jax.numpy", "jax.grad"],
                },
            },
        },
        "database": {
            "python": {
                "sqlalchemy": {
                    "files": ["requirements.txt"],
                    "patterns": [r"sqlalchemy", r"sa\."],
                    "markers": ["import sqlalchemy", "from sqlalchemy", "Base.metadata"],
                },
                "django_orm": {
                    "files": ["requirements.txt"],
                    "patterns": [r"django"],
                    "markers": ["models.Model", "class Meta:", "ForeignKey"],
                },
                "peewee": {
                    "files": ["requirements.txt"],
                    "patterns": [r"peewee"],
                    "markers": ["import peewee", "Model, CharField"],
                },
                "tortoise": {
                    "files": ["requirements.txt"],
                    "patterns": [r"tortoise"],
                    "markers": ["from tortoise", "Model", "@backward_fk"],
                },
            },
            "javascript": {
                "mongoose": {
                    "files": ["package.json"],
                    "patterns": [r"mongoose"],
                    "markers": ["mongoose.connect", "Schema", "model()"],
                },
                "prisma": {
                    "files": ["package.json", "prisma/schema.prisma"],
                    "patterns": [r"prisma"],
                    "markers": ["prisma.client", "PrismaClient"],
                },
                "typeorm": {
                    "files": ["package.json"],
                    "patterns": [r"typeorm"],
                    "markers": ["@Entity", "DataSource", "@Column"],
                },
                "sequelize": {
                    "files": ["package.json"],
                    "patterns": [r"sequelize"],
                    "markers": ["Sequelize", "define", "belongsTo"],
                },
            },
            "go": {
                "gorm": {
                    "files": ["go.mod"],
                    "patterns": [r"gorm.io/gorm"],
                    "markers": ["gorm.Open", "DB.AutoMigrate", "Model struct"],
                },
                "sqlx": {
                    "files": ["go.mod"],
                    "patterns": [r"jmoiron/sqlx"],
                    "markers": ["sqlx.Connect", "sqlx.NamedStmt"],
                },
                "ent": {
                    "files": ["go.mod"],
                    "patterns": [r"entgo.io/ent"],
                    "markers": ["ent.Client", "ent.Schema"],
                },
            },
        },
        "testing": {
            "python": {
                "pytest": {
                    "files": ["pytest.ini", "pyproject.toml"],
                    "patterns": [r"pytest"],
                    "markers": ["def test_", "import pytest", "@pytest.fixture"],
                },
                "unittest": {
                    "files": [],
                    "patterns": [r"unittest", r"TestCase"],
                    "markers": ["import unittest", "unittest.TestCase", "self.assert"],
                },
                "hypothesis": {
                    "files": ["requirements.txt"],
                    "patterns": [r"hypothesis"],
                    "markers": ["from hypothesis", "@given", "settings(database"],
                },
            },
            "javascript": {
                "jest": {
                    "files": ["package.json", "jest.config.js"],
                    "patterns": [r"jest"],
                    "markers": ["describe(", "test(", "expect("],
                },
                "mocha": {
                    "files": ["package.json"],
                    "patterns": [r"mocha"],
                    "markers": ["describe(", "it(", "before("],
                },
                "cypress": {
                    "files": ["package.json", "cypress.json"],
                    "patterns": [r"cypress"],
                    "markers": ["cy.visit", "cy.get", "cypress."],
                },
                "vitest": {
                    "files": ["package.json", "vitest.config.js"],
                    "patterns": [r"vitest"],
                    "markers": ["describe(", "test(", "vi."],
                },
            },
            "java": {
                "junit": {
                    "files": ["pom.xml", "build.gradle"],
                    "patterns": [r"junit", r"Jupiter"],
                    "markers": ["@Test", "TestCase", "Assertions.assert"],
                },
                "testng": {
                    "files": ["pom.xml"],
                    "patterns": [r"testng"],
                    "markers": ["@Test", "@BeforeMethod", "TestNG"],
                },
            },
            "go": {
                "testing": {
                    "files": ["go.mod"],
                    "patterns": [r"testing"],
                    "markers": ["func Test", "testing.T", "testing.M"],
                },
                "ginkgo": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/onsi/ginkgo"],
                    "markers": ["Describe(", "It(", "BeforeEach("],
                },
                "goconvey": {
                    "files": ["go.mod"],
                    "patterns": [r"github.com/smartystreets/goconvey"],
                    "markers": ["Convey(", "ShouldEqual", "So("],
                },
            },
        },
        "build_tools": {
            "python": {
                "poetry": {
                    "files": ["pyproject.toml", "poetry.lock"],
                    "patterns": [r"poetry"],
                    "markers": ["[tool.poetry]", "[build-system]"],
                },
                "pipenv": {
                    "files": ["Pipfile", "Pipfile.lock"],
                    "patterns": [r"pipenv"],
                    "markers": ["[packages]", "[dev-packages]"],
                },
                "setuptools": {
                    "files": ["setup.py", "setup.cfg", "pyproject.toml"],
                    "patterns": [r"setuptools"],
                    "markers": ["setup(", "setuptools.setup"],
                },
                "hatch": {
                    "files": ["pyproject.toml"],
                    "patterns": [r"hatch"],
                    "markers": ["[tool.hatch]", "hatchling.build"],
                },
            },
            "javascript": {
                "npm": {
                    "files": ["package.json", "package-lock.json"],
                    "patterns": [r"npm"],
                    "markers": ["scripts", "dependencies"],
                },
                "yarn": {
                    "files": ["package.json", "yarn.lock"],
                    "patterns": [r"yarn"],
                    "markers": ["yarn.lock"],
                },
                "pnpm": {
                    "files": ["package.json", "pnpm-lock.yaml"],
                    "patterns": [r"pnpm"],
                    "markers": ["pnpm-lock.yaml"],
                },
            },
            "java": {
                "maven": {
                    "files": ["pom.xml", ".mvn"],
                    "patterns": [r"maven"],
                    "markers": ["<groupId>", "<artifactId>", "<version>"],
                },
                "gradle": {
                    "files": ["build.gradle", "build.gradle.kts", "gradlew"],
                    "patterns": [r"gradle"],
                    "markers": ["plugins {", "dependencies {", "repositories {"],
                },
            },
            "rust": {
                "cargo": {
                    "files": ["Cargo.toml", "Cargo.lock"],
                    "patterns": [r"cargo"],
                    "markers": ["[package]", "[dependencies]"],
                },
            },
        },
    }

    def __init__(self):
        self._detected_frameworks: List[FrameworkInfo] = []
        self._build_tools: List[str] = []
        self._testing_frameworks: List[str] = []

    def detect(self, root_path: str) -> ProjectFrameworks:
        self._detected_frameworks = []
        self._build_tools = []
        self._testing_frameworks = []

        if not os.path.exists(root_path):
            return ProjectFrameworks()

        for category, languages in self.FRAMEWORK_MARKERS.items():
            for language, frameworks in languages.items():
                for framework_name, config in frameworks.items():
                    self._check_framework(root_path, framework_name, category, language, config)

        return ProjectFrameworks(
            frameworks=self._detected_frameworks,
            build_tools=self._build_tools,
            testing_frameworks=self._testing_frameworks,
        )

    def _check_framework(
        self,
        root_path: str,
        framework_name: str,
        category: str,
        language: str,
        config: Dict,
    ):
        indicators: List[str] = []
        confidence = 0.0

        for file_pattern in config.get("files", []):
            file_path = os.path.join(root_path, file_pattern)
            if os.path.exists(file_path):
                confidence += 0.3
                indicators.append(f"Found: {file_pattern}")

        content = self._read_project_files(root_path)

        for pattern in config.get("patterns", []):
            if re.search(pattern, content, re.IGNORECASE):
                confidence += 0.2
                indicators.append(f"Pattern match: {pattern}")

        for marker in config.get("markers", []):
            if marker in content:
                confidence += 0.2
                indicators.append(f"Marker found: {marker}")

        if confidence >= 0.3:
            self._detected_frameworks.append(FrameworkInfo(
                name=framework_name,
                category=category,
                confidence=min(confidence, 1.0),
                indicators=indicators,
            ))

            if category == "build_tools":
                self._build_tools.append(framework_name)
            elif category == "testing":
                self._testing_frameworks.append(framework_name)

    def _read_project_files(self, root_path: str) -> str:
        content_parts = []

        for root, dirs, files in os.walk(root_path):
            if "node_modules" in root or ".git" in root or "__pycache__" in root:
                continue

            for file in files:
                file_path = os.path.join(root, file)
                ext = os.path.splitext(file)[1].lower()

                if ext in [".py", ".js", ".ts", ".java", ".go", ".rs", ".cpp", ".c", ".h", ".rb", ".php", ".json", ".xml", ".yaml", ".yml", ".toml", ".gradle", ".md"]:
                    try:
                        with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
                            content_parts.append(f.read())
                    except Exception:
                        pass

        return "\n".join(content_parts)

    def get_detected_frameworks(self) -> List[FrameworkInfo]:
        return sorted(self._detected_frameworks, key=lambda f: f.confidence, reverse=True)

    def get_build_tools(self) -> List[str]:
        return self._build_tools

    def get_testing_frameworks(self) -> List[str]:
        return self._testing_frameworks


def detect_frameworks(root_path: str) -> ProjectFrameworks:
    detector = FrameworkDetector()
    return detector.detect(root_path)
