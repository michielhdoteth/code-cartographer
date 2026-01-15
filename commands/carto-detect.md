---
description: Detect frameworks and libraries used in the codebase
parameters: {}
---

# carto-detect

Detect frameworks, libraries, and tools used in the codebase.

## Usage

```
/carto:carto-detect
```

## What It Detects

### Web Frameworks

| Language | Frameworks |
|----------|------------|
| Python | Django, Flask, FastAPI, Pyramid, Bottle, Tornado, Aiohttp, Starlette |
| JavaScript/TypeScript | React, Vue, Angular, Express, Next.js, Nuxt, Svelte, Astro, Vite, Webpack |
| Java | Spring, Quarkus, Micronaut, Javalin, Spark |
| Go | Gin, Echo, Fiber, Chi, Gorilla, Beego, Revel |
| Rust | Actix, Rocket, Warp, Axum, Tokio, Yew |
| PHP | Laravel, Symfony, Yii, CodeIgniter, CakePHP, Drupal, WordPress, Slim |
| Ruby | Rails, Sinatra, Hanami, Padrino, Grape |
| C++ | Qt, Boost, GTK, wxWidgets |

### ML/AI Frameworks

| Language | Frameworks |
|----------|------------|
| Python | TensorFlow, PyTorch, Keras, scikit-learn, Pandas, NumPy, OpenCV, Transformers, JAX |

### Database Tools

| Language | Tools |
|----------|-------|
| Python | SQLAlchemy, Django ORM, Peewee, Tortoise |
| JavaScript/TypeScript | Mongoose, Prisma, TypeORM, Sequelize |
| Go | GORM, sqlx, ent |

### Testing Frameworks

| Language | Frameworks |
|----------|------------|
| Python | pytest, unittest, hypothesis |
| JavaScript/TypeScript | Jest, Mocha, Cypress, Vitest |
| Java | JUnit, TestNG |
| Go | testing, Ginkgo, GoConvey |

### Build Tools

| Language | Tools |
|----------|-------|
| Python | Poetry, Pipenv, Setuptools, Hatch |
| JavaScript/TypeScript | npm, yarn, pnpm |
| Java | Maven, Gradle |
| Rust | Cargo |

## How Detection Works

1. **Configuration Files**: Looks for package.json, requirements.txt, Cargo.toml, pom.xml, etc.
2. **Import Patterns**: Scans source code for framework-specific imports
3. **Markers**: Detects framework-specific markers (decorators, classes, functions)

## Output

Returns detected frameworks with:
- **Name**: Framework name
- **Category**: Web, ML/AI, Database, Testing, Build Tools
- **Confidence**: Detection confidence (0.0 - 1.0)
- **Indicators**: What was found (files, patterns, markers)

## Examples

```
/carto:carto-detect

Output:
{
  "status": "success",
  "frameworks": [
    {"name": "react", "category": "web", "confidence": 0.9, "indicators": ["Found: package.json", "Pattern match: react"]},
    {"name": "express", "category": "web", "confidence": 0.7, "indicators": ["Pattern match: express"]},
    {"name": "jest", "category": "testing", "confidence": 0.8, "indicators": ["Found: jest.config.js"]}
  ],
  "build_tools": ["npm", "yarn"],
  "testing_frameworks": ["jest"],
  "total_frameworks": 3
}
```

## Use Cases

- Understand project technology stack
- Identify migration dependencies
- Check for outdated frameworks
- Verify build configuration
