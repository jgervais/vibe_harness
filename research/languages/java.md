# Java — AI Code Smells & Tree-Sitter Detection

## 1. Common AI-Generated Java Anti-Patterns

### Over-Engineering
- **God classes**: AI tends to create massive classes doing everything (e.g., a `UserService` that handles CRUD, email, auth, and reporting). Detect by counting methods/fields per class.
- **Premature abstraction**: Unnecessary interfaces for single implementations (`UserServiceImpl` with no `UserService` alternative). Detect interface with exactly one implementor.
- **Excessive design patterns**: Factory-of-factories, strategy patterns with one strategy, abstract classes never extended.
- **Unnecessary generic types**: `<T>` on classes that only ever use one type.

### Error Handling
- **Empty catch blocks**: `catch (Exception e) {}` — AI's favorite way to "handle" errors. Very common.
- **Swallowed exceptions**: `catch (Exception e) { /* do nothing */ }` or logging at wrong level.
- **Overly broad catches**: `catch (Exception e)` instead of specific types; `throws Exception` on method signatures.
- **Missing finally/closing**: Try blocks without try-with-resources for `AutoCloseable` resources.

### Logging & Observability
- **No logging at all**: AI often skips `Logger` entirely in service classes.
- **System.out.println instead of Logger**: Debugging output left in production code.
- **Inconsistent log levels**: Everything at `INFO`, or `ERROR` for non-errors.

### Structural Smells
- **All-static utility classes**: `public class Util { public static String foo()... }` when DI/beans are appropriate.
- **Anemic domain models**: Data classes with only getters/setters, all logic in separate service classes (common AI pattern that fights Spring idioms).
- **String concatenation in loops**: `result += item` instead of `StringBuilder`.
- **Mutable static state**: `private static Map cache = new HashMap<>()` without synchronization.
- **Missing package-info.java**: AI generates classes but skips package documentation.
- **Hardcoded config**: URLs, passwords, connection strings embedded in source instead of `application.properties`/environment.

## 2. Maven/Gradle Ecosystem — AI Pitfalls

### Dependency Management
- **Mixed dependency styles**: AI sometimes mixes `<dependencyManagement>` and direct `<dependencies>` incorrectly.
- **Version hardcoding**: Placing version numbers directly in `<dependency>` instead of `<properties>` or BOM.
- **Scope mismatches**: `compile` scope for test dependencies, missing `<scope>provided</scope>` for servlet APIs.
- **Spring Boot BOM ignored**: AI adds explicit versions for dependencies covered by `spring-boot-dependencies` BOM.
- **Ghost dependencies**: Declaring deps not actually used in code (AI copies from examples).
- **Missing transitive deps**: AI assumes transitive deps are stable; they can shift across versions.

### Build Config Smells
- **Pom.xml with no `<properties>` section**: Versions scattered throughout.
- **No `<pluginManagement>`**: Plugin versions unspecified (builds break on different machines).
- **Gradle: mixing `implementation` and `api`**: AI often uses `api` when `implementation` suffices.
- **Missing `maven-compiler-plugin` source/target**: Defaults to Java 5 in old archetypes.

## 3. Tree-Sitter Java AST — Key Node Types

### Class/Interface Hierarchy
| Node Type | What It Captures |
|---|---|
| `class_declaration` | Class name, modifiers, extends, implements |
| `interface_declaration` | Interface definition |
| `enum_declaration` | Java enum (often misused by AI) |
| `record_declaration` | Java 16+ records |
| `annotation_type_declaration` | `@interface` custom annotations |
| `extends` / `implements` | Type hierarchy relationships |

### Method Level
| Node Type | What It Captures |
|---|---|
| `method_declaration` | Name, return type, params, throws |
| `constructor_declaration` | Constructor (AI often generates unnecessary ones) |
| `static_initializer` | `static { }` blocks |
| `method_invocation` | Method calls — useful for detecting System.out, printStackTrace |

### Error Handling
| Node Type | What It Captures |
|---|---|
| `catch_clause` | Catch block — detect empty bodies |
| `try_statement` | Try-with-resources vs plain try |
| `throw_statement` | Throw expressions |
| `try_with_resources_statement` | Resource management |

### Annotations (Critical for Spring/Jakarta)
| Node Type | What It Captures |
|---|---|
| `annotation` | `@Override`, `@Autowired`, `@Bean`, etc. |
| `marker_annotation` | No-value annotations |
| `annotation_argument_list` | Annotation parameters |

### Other Useful Nodes
| Node Type | Use For |
|---|---|
| `field_declaration` | Count fields per class (god class detection) |
| `local_variable_declaration` | Local vars, unused detection |
| `import_declaration` | Wildcard imports (`.*`), unused imports |
| `modifiers` | `public`, `static`, `final`, `synchronized` |
| `block` | Statement block — empty body detection |
| `synchronized_statement` | Concurrency patterns |
| `lambda_expression` | Java 8+ functional style |

## 4. Framework-Specific AI Issues

### Spring Boot
- **`@Autowired` on fields**: AI uses field injection instead of constructor injection (Spring best practice).
- **Missing `@Transactional`**: Service methods modifying data without transaction annotations.
- **`@Controller` instead of `@RestController`**: AI picks the wrong stereotype for REST APIs.
- **Overuse of `@Component`**: When `@Service`/`@Repository` are more appropriate.
- **Missing `@ConfigurationProperties`**: Hardcoding values that should be externalized.
- **God `@RestController`**: Endpoints + business logic + data access in one class.
- **`@SpringBootApplication` on wrong package**: Not in root package, missing component scan.

### Jakarta EE (formerly Java EE)
- **Outdated `javax.` imports**: AI uses `javax.servlet` instead of `jakarta.servlet`.
- **Missing `@Inject`**: Using `new` for managed beans instead of CDI.
- **Wrong JPA annotations**: `@Column` without `nullable=false`, missing `@Entity` on DTOs.

## 5. Detection Rules — Tree-Sitter Queries

### Empty Catch Block
```
(catch_clause body: (block . (_) .))  → NOT this (has content)
(catch_clause body: (block))          → empty block = smell
```
Detect: `catch_clause` whose `body` is a `block` with zero child statements.

### God Class (Heuristic)
- Count `method_declaration` + `field_declaration` children under a `class_declaration`.
- Flag if methods > 15 or fields > 20 (tunable).

### Static Utility Class (Anti-pattern in Spring)
```
(class_declaration
  modifiers: (modifiers "public")
  (method_declaration modifiers: (modifiers "public" "static"))
  ...)
```
Class with only `static` methods = likely utility class, suspicious in Spring context.

### System.out.println
```
(method_invocation
  object: (field_access object: (identifier) @obj field: (identifier) @field)
  (#eq? @obj "System")
  (#match? @field "out"))
```

### Field Injection (Spring Anti-pattern)
```
(field_declaration
  modifiers: (modifiers (annotation name: (identifier) @ann))
  (#eq? @ann "Autowired"))
```
Constructor injection is preferred.

### Missing try-with-resources
If `local_variable_declaration` assigns a type implementing `AutoCloseable` but is NOT inside a `try_with_resources_statement`, that's a smell.

## 6. Quick Reference — Node Count for Smells

| Smell | Primary AST Signal | Metric |
|---|---|---|
| God class | `class_declaration` children | methods + fields > threshold |
| Empty catch | `catch_clause` → empty `block` | body stmt count == 0 |
| No logging | No `Logger` field in service class | 0 logger fields |
| Static util | All methods `static` | 100% static methods |
| Field injection | `@Autowired` on field | annotation on field_declaration |
| Swallowed exc | `catch_clause` with only comments | body is comment-only |
| Missing resources | `AutoCloseable` not in try-with-resources | type mismatch |
| println debug | `System.out` or `System.err` calls | method_invocation pattern |