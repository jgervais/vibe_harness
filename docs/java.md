# Java — Build System & Existing Linters

## Build Systems

### Maven
- **Config:** `pom.xml`
- **Integration:** Exec plugin
  ```xml
  <plugin>
    <groupId>org.codehaus.mojo</groupId>
    <artifactId>exec-maven-plugin</artifactId>
    <version>3.1.0</version>
    <executions>
      <execution>
        <id>vibe-check</id>
        <phase>validate</phase>
        <goals><goal>exec</goal></goals>
        <configuration>
          <executable>vibe-harness</executable>
          <arguments><argument>src/main/java</argument></arguments>
        </configuration>
      </execution>
    </executions>
  </plugin>
  ```
  ```bash
  mvn validate
  ```

### Gradle
- **Config:** `build.gradle` or `build.gradle.kts`
- **Integration:**
  ```kotlin
  tasks.register("vibeCheck", Exec::class) {
      commandLine("vibe-harness", "src/main/java")
  }
  tasks.check { dependsOn("vibeCheck") }
  ```
  ```bash
  ./gradlew vibeCheck
  ```

### Gradle (Kotlin DSL)
- **Config:** `build.gradle.kts`
- **Integration:** Same as above, Kotlin DSL syntax

### Bazel
- **Config:** `WORKSPACE`, `BUILD` files
- **Integration:** Custom rule or test target
  ```python
  # BUILD
  sh_test(
      name = "vibe_check",
      srcs = ["vibe_check.sh"],
      args = ["$(location //:srcs)"],
  )
  ```
  Or shell script that runs `vibe-harness` on source directories.

## Frameworks

### Spring Boot
- **Build:** Maven or Gradle
- **Integration:** Add to existing build lifecycle
  - Maven: `validate` phase (runs before compile)
  - Gradle: `check` task dependency
- **Specific concerns:** Spring annotations (@Autowired, @Transactional, @RestController) are checked by VH language-specific rules

### Jakarta EE
- **Build:** Maven
- **Integration:** Same as Maven above
- **Specific concerns:** javax vs jakarta imports, @Inject usage

### Quarkus
- **Build:** Maven or Gradle
- **Integration:** Add to quarkus-maven-plugin or Gradle check

### Micronaut
- **Build:** Maven or Gradle
- **Integration:** Standard build integration

## Existing Linters to Leverage

| Linter | What It Catches | Overlap with Vibe Harness |
|--------|----------------|--------------------------|
| **Checkstyle** | Style, complexity, design | File length (LineLength), method length (MethodLength), empty catch (EmptyCatchBlock), god class (ClassFanOut) |
| **PMD** | Bug patterns, design, style | God class, empty catch, missing break, system.out |
| **SpotBugs** | Bug patterns | FindBugs successor, null pointer, bad casts |
| **Error Prone** | Compiler-error-level bugs | Injected constructors, missing @Override |
| **detekt** (Kotlin) | Kotlin-specific | Analogous rules for Kotlin codebases |
| **ktlint** (Kotlin) | Kotlin style | Formatting only |
| **Snyk** | Dependency vulnerabilities | Complementary — different domain |

### Leverage Strategy
- **Checkstyle + PMD first** for style and conventional checks
- **Vibe Harness adds what they miss:** missing logging in service methods, @Autowired field injection (instead of constructor), missing @Transactional, System.out.println detection, swallowed exceptions
- **Checkstyle MethodLength** can match VH-G002 threshold
- **PMD EmptyCatchBlock** overlaps with VH-G004 — but VH is non-configurable where PMD can be suppressed
- **SpotBugs** catches runtime bugs — complementary domain

### Checkstyle Configuration for Maximum Overlap
```xml
<!-- checkstyle.xml -->
<module name="Checker">
  <module name="FileLength">
    <property name="max" value="300"/>
  </module>
  <module name="TreeWalker">
    <module name="MethodLength">
      <property name="max" value="50"/>
    </module>
    <module name="EmptyCatchBlock"/>
    <module name="ExecutableStatementCount">
      <property name="max" value="50"/>
    </module>
  </module>
</module>
```

### PMD Configuration for Maximum Overlap
```xml
<!-- ruleset.xml -->
<ruleset>
  <rule ref="category/java/design.xml/GodClass"/>
  <rule ref="category/java/design.xml/TooManyMethods"/>
  <rule ref="category/java/errorprone.xml/EmptyCatchBlock"/>
  <rule ref="category/java/bestpractices.xml/SystemPrintln"/>
  <rule ref="category/java/design.xml/ExcessiveMethodLength"/>
</ruleset>
```